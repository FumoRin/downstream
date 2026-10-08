export function formatBytes(bytes: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

export function formatSpeed(bytesPerSec: number) {
  if (!bytesPerSec || bytesPerSec <= 0) return "-";
  return `${formatBytes(bytesPerSec)}/s`;
}

export function formatETA(seconds: number): string {
  if (seconds === undefined || seconds === null || seconds <= 0) return "--";
  if (!isFinite(seconds)) return "--";
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

export function formatDate(isoStr: string): string {
  if (!isoStr) return "--";
  const d = new Date(isoStr);
  if (isNaN(d.getTime())) return "--";
  const m = (d.getMonth() + 1).toString().padStart(2, "0");
  const day = d.getDate().toString().padStart(2, "0");
  const h = d.getHours().toString().padStart(2, "0");
  const min = d.getMinutes().toString().padStart(2, "0");
  return `${m}-${day} ${h}:${min}`;
}

export function splitPath(fullPath: string): {filename: string; dir: string} {
  if (!fullPath) return { filename: "download", dir: ""}
  const normalized = fullPath.replace(/\\/g, "/")
  const lastSlash = normalized.lastIndexOf("/")
  if (lastSlash === -1 ) {
    return {filename: fullPath, dir: ""}
  }
  return {
    filename: normalized.slice(lastSlash + 1),
    dir: normalized.slice(0, lastSlash)
  }
}
