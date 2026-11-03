const DOWNLOADABLE_EXTENSIONS = new Set([
  // Archives
  "zip",
  "rar",
  "7z",
  "tar",
  "gz",
  "bz2",
  "xz",
  "iso",
  "dmg",
  "pkg",
  "deb",
  "rpm",
  // Executables & Installers
  "exe",
  "msi",
  "apk",
  "appimage",
  "bin",
  // Documents & Media
  "pdf",
  "epub",
  "mp4",
  "mkv",
  "avi",
  "mov",
  "flv",
  "webm",
  "mp3",
  "flac",
  "wav",
  "ogg",
  "opus",
]);

function isDownloadableUrl(urlStr) {
  try {
    const url = new URL(urlStr, window.location.href);
    if (!["http:", "https:"].includes(url.protocol)) {
      return false;
    }
    const pathname = url.pathname;
    const lastSegment = pathname.split("/").filter(Boolean).pop();
    if (!lastSegment) return false;

    const parts = lastSegment.split(".");
    if (parts.length > 1) {
      const ext = parts.pop().toLowerCase();
      if (DOWNLOADABLE_EXTENSIONS.has(ext)) {
        return true;
      }
    }
  } catch (_) { }
  return false;
}

// Capture phase click listener
window.addEventListener(
  "click",
  (event) => {
    // Escape hatch: If user is holding Alt, let the browser download natively!
    if (event.altKey) {
      return;
    }

    // Check if target or ancestor is an <a> tag with href
    const anchor = event.target.closest("a[href]");
    if (!anchor) return;

    const href = anchor.href;
    if (!href || href.startsWith("javascript:") || href.startsWith("#")) {
      return;
    }

    // Intercept if it has the HTML5 download attribute or matches known extensions
    const hasDownloadAttr = anchor.hasAttribute("download");
    const isDownloadFile = isDownloadableUrl(href);

    if (hasDownloadAttr || isDownloadFile) {
      // 1. Prevent native navigation / browser download entirely
      event.preventDefault();
      event.stopPropagation();

      // 2. Extract suggested filename if present
      const suggestedName = anchor.getAttribute("download") || "";

      // 3. Dispatch to background script
      browser.runtime
        .sendMessage({
          action: "intercepted_click",
          url: href,
          filename: suggestedName,
          referrer: window.location.href,
        })
        .catch((err) => {
          console.warn(
            "[Downstream] Could not send message to background:",
            err,
          );
        });
    }
  },
  true, // useCapture = true is crucial to intercept before page scripts
);
