export type DownloadStatus = 0 | 1 | 2 | 3 | 4 | 5;
export const DownloadStatusEnum = {
  Queued: 0,
  Downloading: 1,
  Paused: 2,
  Completed: 3,
  Error: 4,
  Scheduled: 5,
};

export interface PartState {
  id: string;
  start: number;
  end: number;
  current: number;
  pct: number;
  speed?: number;
  status: "Done" | "Downloading" | "Paused" | "Idle";
}

export interface ChunkBlock {
  index: number;
  start: number;
  end: number;
  state: "done" | "active" | "pending";
  workerThread?: number;
}

export type DetailsTab = "general" | "parts";

export interface DownloadItem {
  id: string;
  filename: string;
  category: "Video" | "Document" | "Compressed" | "Programs" | "Audio" | string;
  url: string;
  totalSize: number;
  currentSize: number;
  percentage: number;
  speed: number;
  eta: number;
  status: DownloadStatus;
  priority: "low" | "normal" | "high";
  createdAt: string;
  scheduledAt?: string | null;
  saveDir: string;
  parts: PartState[];
}

export type SortField =
  | "filename"
  | "category"
  | "totalSize"
  | "percentage"
  | "speed"
  | "eta"
  | "status"
  | "createdAt";

export type SortOrder = "asc" | "desc";

export interface FilterState {
  type: "status" | "category";
  value: string;
}

export interface RateLimiterConfig {
  enabled: boolean;
  displayValue: number;
  unit: "KB" | "MB";
  limitBytesPerSec: number;
}
