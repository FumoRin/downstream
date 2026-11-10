let interceptionEnabled = true;
browser.storage.local.get("enabled").then((data) => {
  interceptionEnabled = data.enabled !== false;
});
browser.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.enabled !== undefined) {
    interceptionEnabled = changes.enabled.newValue !== false;
  }
});

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

function isRecentlySent(url) {
  const now = Date.now();
  if (recentDownloads.has(url)) {
    const lastTime = recentDownloads.get(url);
    if (now - lastTime < 3000) {
      return true;
    }
  }
  return false;
}

function markAsSent(url) {
  recentDownloads.set(url, Date.now());
  if (recentDownloads.size > 100) {
    const oldestKey = recentDownloads.keys().next().value;
    recentDownloads.delete(oldestKey);
  }
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

function parseContentDispositionFilename(headerValue) {
  if (!headerValue) return "";
  // Check for filename*=UTF-8''filename.ext
  const utf8Match = headerValue.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match) {
    try {
      return decodeURIComponent(utf8Match[1]);
    } catch (_) { }
  }
  // Check for standard filename="filename.ext" or filename=filename.ext
  const match = headerValue.match(/filename=["']?([^"';]+)["']?/i);
  return match ? match[1].trim() : "";
}

function isDownloadableResponse(headers, urlStr) {
  let isAttachment = false;
  let filename = "";
  let contentType = "";

  for (const h of headers) {
    const name = h.name.toLowerCase();
    if (name === "content-disposition") {
      if (h.value.toLowerCase().includes("attachment")) {
        isAttachment = true;
      }
      filename = parseContentDispositionFilename(h.value);
    } else if (name === "content-type") {
      contentType = h.value.toLowerCase();
    }
  }

  // Never intercept standard web browsing pages / scripts
  if (
    contentType.includes("text/html") ||
    contentType.includes("application/json") ||
    contentType.includes("text/css") ||
    contentType.includes("javascript")
  ) {
    return { shouldIntercept: false, filename: "" };
  }

  // Intercept if server explicitly designated attachment
  if (isAttachment) {
    return { shouldIntercept: true, filename: filename };
  }

  // Or if content-type is binary octet-stream / known archive
  if (
    contentType.includes("application/octet-stream") ||
    contentType.includes("application/x-msdownload") ||
    contentType.includes("application/zip")
  ) {
    return { shouldIntercept: true, filename: filename };
  }

  return { shouldIntercept: false, filename: "" };
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

async function getCookiesForUrl(urlStr, storeId) {
  try {
    const details = { url: urlStr };
    if (storeId) {
      details.storeId = storeId;
    }
    const cookies = await browser.cookies.getAll(details);
    if (!cookies || cookies.length === 0) {
      return "";
    }
    return cookies.map((c) => `${c.name}=${c.value}`).join("; ");
  } catch (err) {
    console.warn("[Downstream] Could not retrieve cookies for URL:", err);
    return "";
  }
}

async function forwardToDownstream(targetUrl, filename, referrer, storeId) {
  if (isRecentlySent(targetUrl)) {
    return;
  }
  markAsSent(targetUrl);

  const displayName = filename || extractFilename(targetUrl) || "download";

  // 1. Try targetUrl cookies; fallback to referrer if target is a CDN (like googleusercontent.com)
  let cookies = await getCookiesForUrl(targetUrl, storeId);
  if (!cookies && referrer && referrer.startsWith("http")) {
    cookies = await getCookiesForUrl(referrer, storeId);
  }

  const headers = {
    "User-Agent": navigator.userAgent,
  };
  if (referrer) {
    headers["Referer"] = referrer;
  }

  const payload = {
    action: "download",
    url: targetUrl,
    filename: filename,
    referrer: referrer,
    cookies: cookies,
    headers: headers,
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
  await forwardToDownstream(targetUrl, filename, referrer, tab?.cookieStoreId);
});

// 2. DOM Click Interception listener from content.js
browser.runtime.onMessage.addListener(async (message, sender) => {
  if (message.action === "intercepted_click") {
    const enabled = await isInterceptionEnabled();
    if (!enabled) return;

    const filename = extractFilename(message.url, message.filename);
    await forwardToDownstream(
      message.url,
      filename,
      message.referrer || "",
      sender?.tab?.cookieStoreId,
    );
  }
});

// 3. Network-layer interception: stops downloads BEFORE Firefox's download UI ever triggers
browser.webRequest.onHeadersReceived.addListener(
  (details) => {
    if (
      !interceptionEnabled ||
      details.tabId === -1 ||
      !details.responseHeaders
    )
      return;
    // 1. Ignore background extension requests, data/blob URLs, and disabled state
    if (details.tabId === -1 || !details.responseHeaders) return;

    // 2. Check if the HTTP response is a downloadable file
    const { shouldIntercept, filename } = isDownloadableResponse(
      details.responseHeaders,
      details.url,
    );

    if (shouldIntercept) {
      const finalFilename = filename || extractFilename(details.url);

      // Forward to Downstream asynchronously
      forwardToDownstream(
        details.url,
        finalFilename,
        details.originUrl || details.documentUrl || "",
        details.cookieStoreId,
      );

      // Cancel at network level: browser download UI NEVER triggers!
      return { cancel: true };
    }
  },
  { urls: ["<all_urls>"], types: ["main_frame", "sub_frame", "other"] },
  ["blocking", "responseHeaders"],
);
