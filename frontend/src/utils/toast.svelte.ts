export type ToastType = "info" | "success" | "warning" | "error";

export interface ToastItem {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
}

class ToastManager {
  toasts = $state<ToastItem[]>([]);

  show(message: string, type: ToastType = "info", duration = 3000): string {
    const id = Math.random().toString(36).substring(2, 9);
    const item: ToastItem = { id, message, type, duration };
    this.toasts.push(item);

    if (duration > 0) {
      setTimeout(() => {
        this.dismiss(id);
      }, duration);
    }
    return id;
  }

  dismiss(id: string) {
    this.toasts = this.toasts.filter((t) => t.id !== id);
  }
}

export const toastManager = new ToastManager();

export function showToast(
  message: string,
  type: ToastType = "info",
  duration = 3000,
) {
  return toastManager.show(message, type, duration);
}
