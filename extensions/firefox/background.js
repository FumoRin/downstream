// Register right-click context menu when extension is installed
browser.runtime.onInstalled.addListener(() => {
  browser.contextMenus.create({
    id: "download-with-downstream",
    title: "Download with Downstream",
    contexts: ["link", "image", "video", "audio"],
  });
});

// Cache recently sent URLs to prevent duplicate triggers
const recentDownloads = new Map();

function shouldSend(url) {
  const now = Date.now();
  if (recentDownloads.has(url)) {
    const lastTime = recentDownloads.get(url);
    if (now - lastTime < 3000) {
      return false; // deduplicate within 3 seconds
    }
  }
  recentDownloads.set(url, now);
  if (recentDownloads.size > 100) {
    const oldestKey = recentDownloads.keys().next().value;
    recentDownloads.delete(oldestKey);
  }
  return true;
}

function extractFilename(urlStr, suggestedFilename) {
  if (suggestedFilename && suggestedFilename.trim()) {
    const clean = suggestedFilename.trim();
    const lastPart = clean.split(/[/\\]/).pop();
    if (lastPart && lastPart.includes(".")) {
      return lastPart;
    }
  }
  try {
    const url = new URL(urlStr);
    const pathname = url.pathname;
    const segments = pathname.split("/").filter(Boolean);
    if (segments.length > 0) {
      const last = decodeURIComponent(segments[segments.length - 1]);
      if (last.includes(".")) {
        return last;
      }
    }
  } catch (_) { }
  return "";
}

async function isInterceptionEnabled() {
  const data = await browser.storage.local.get("enabled");
  return data.enabled !== false; // default to true
}

function notify(title, message) {
  try {
    browser.notifications.create({
      type: "basic",
      title: title,
      message: message,
    });
  } catch (err) {
    console.warn("[Downstream] Could not trigger notification:", err);
  }
}

async function forwardToDownstream(targetUrl, filename, referrer) {
  if (!shouldSend(targetUrl)) {
    return;
  }

  const displayName = filename || extractFilename(targetUrl) || "download";

  const payload = {
    action: "download",
    url: targetUrl,
    filename: filename,
    referrer: referrer,
  };

  try {
    const response = await browser.runtime.sendNativeMessage(
      "downstream",
      payload,
    );
    if (response && response.status === "error") {
      console.error("[Downstream] Error from desktop:", response.message);
      notify(
        "Downstream Error",
        response.message || "Failed to start download",
      );
    } else if (response && response.status === "ok") {
      if (response.message === "prompted") {
        console.log(`Opening confirmation dialog for: ${displayName}`);
      } else {
        notify("Downstream", `Started downloading: ${displayName}`);
      }
    }
  } catch (err) {
    console.error("[Downstream] Failed to send native message:", err);
    notify(
      "Downstream Error",
      "Desktop application is not running. Please launch Downstream.",
    );
  }
}

// 1. Right-click context menu handler
browser.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId !== "download-with-downstream") {
    return;
  }

  const targetUrl = info.linkUrl || info.srcUrl;
  if (!targetUrl) return;

  const filename = extractFilename(targetUrl, "");
  const referrer = info.pageUrl || tab?.url || "";
  await forwardToDownstream(targetUrl, filename, referrer);
});

// 2. DOM Click Interception listener from content.js
browser.runtime.onMessage.addListener(async (message) => {
  if (message.action === "intercepted_click") {
    const enabled = await isInterceptionEnabled();
    if (!enabled) return;

    const filename = extractFilename(message.url, message.filename);
    await forwardToDownstream(message.url, filename, message.referrer || "");
  }
});

// 3. Fallback: browser.downloads.onCreated
browser.downloads.onCreated.addListener(async (downloadItem) => {
  const enabled = await isInterceptionEnabled();
  if (!enabled) return;

  const url = downloadItem.url;
  if (
    !url ||
    url.startsWith("blob:") ||
    url.startsWith("data:") ||
    url.startsWith("about:") ||
    url.startsWith("moz-extension:")
  ) {
    return;
  }

  // If already forwarded by content script in the last 3 seconds, just cancel native download
  if (!shouldSend(url)) {
    try {
      await browser.downloads.cancel(downloadItem.id);
      await browser.downloads.erase({ id: downloadItem.id });
    } catch (_) { }
    return;
  }

  // Otherwise, cancel and forward
  try {
    await browser.downloads.cancel(downloadItem.id);
    await browser.downloads.erase({ id: downloadItem.id });
  } catch (e) {
    console.warn("[Downstream] Could not cancel fallback download:", e);
  }

  const filename = extractFilename(url, downloadItem.filename);
  const referrer = downloadItem.referrer || "";
  await forwardToDownstream(url, filename, referrer);
});
