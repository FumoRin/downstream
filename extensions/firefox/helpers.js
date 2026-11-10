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
  const utf8Match = headerValue.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match) {
    try {
      return decodeURIComponent(utf8Match[1]);
    } catch (_) { }
  }
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
