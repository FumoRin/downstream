const cp1252Map = {
  0x20ac: 0x80,
  0x201a: 0x82,
  0x0192: 0x83,
  0x201e: 0x84,
  0x2026: 0x85,
  0x2020: 0x86,
  0x2021: 0x87,
  0x02c6: 0x88,
  0x2030: 0x89,
  0x0160: 0x8a,
  0x2039: 0x8b,
  0x0152: 0x8c,
  0x017d: 0x8e,
  0x2018: 0x91,
  0x2019: 0x92,
  0x201c: 0x93,
  0x201d: 0x94,
  0x2022: 0x95,
  0x2013: 0x96,
  0x2014: 0x97,
  0x02dc: 0x98,
  0x2122: 0x99,
  0x0161: 0x9a,
  0x203a: 0x9b,
  0x0153: 0x9c,
  0x017e: 0x9e,
  0x0178: 0x9f,
};

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

function fixMojibake(str) {
  if (!str) return str;
  try {
    const bytes = [];
    for (let i = 0; i < str.length; i++) {
      const code = str.charCodeAt(i);
      if (code <= 0xff) {
        bytes.push(code);
      } else if (cp1252Map[code] !== undefined) {
        bytes.push(cp1252Map[code]);
      } else {
        return str; // Character outside cp1252 / latin-1 range
      }
    }
    const decoded = new TextDecoder("utf-8", { fatal: true }).decode(
      new Uint8Array(bytes),
    );
    if (decoded !== str) {
      return decoded;
    }
    return str;
  } catch (_) {
    return str; // Not valid UTF-8, preserve original
  }
}
function extractFilename(urlStr, suggestedFilename) {
  if (suggestedFilename && suggestedFilename.trim()) {
    let clean = suggestedFilename.trim();
    try {
      if (clean.includes("%")) {
        clean = decodeURIComponent(clean);
      }
    } catch (_) { }
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
      let last = segments[segments.length - 1];
      try {
        last = decodeURIComponent(last);
      } catch (_) { }
      if (last.includes(".")) {
        return last;
      }
    }
  } catch (_) { }
  return "";
}

function parseContentDispositionFilename(headerValue) {
  if (!headerValue) return "";

  // 1. RFC 5987 / RFC 6266: filename*=UTF-8'lang'%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf
  const utf8Match = headerValue.match(
    /filename\*=(?:UTF-8|utf-8)'[^']*'([^;\s]+)/i,
  );
  if (utf8Match) {
    try {
      const decoded = decodeURIComponent(
        utf8Match[1].trim().replace(/^["']|["']$/g, ""),
      );
      if (decoded) return decoded;
    } catch (_) { }
  }

  // 2. Standard filename="...": may contain percent-encoding, raw UTF-8, or cp1252 mojibake
  const match = headerValue.match(
    /\bfilename\s*=\s*(?:"([^"]+)"|'([^']+)'|([^;\s]+))/i,
  );
  if (match) {
    const raw = (match[1] || match[2] || match[3] || "").trim();
    if (raw) {
      try {
        if (raw.includes("%")) {
          return fixMojibake(decodeURIComponent(raw));
        }
      } catch (_) { }
      return fixMojibake(raw);
    }
  }

  return "";
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
