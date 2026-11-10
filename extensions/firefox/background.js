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

async function isInterceptionEnabled() {
  const data = await browser.storage.local.get("enabled");
  return data.enabled !== false; // default to true
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

    // Check if the HTTP response is a downloadable file
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
