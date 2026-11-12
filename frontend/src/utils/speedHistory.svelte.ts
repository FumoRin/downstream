import { SvelteMap } from "svelte/reactivity";

export interface ItemSpeedHistory {
  history: number[];
  peak: number;
}

class SpeedHistoryTracker {
  private map = new SvelteMap<string, ItemSpeedHistory>();

  recordSample(id: string, speed: number) {
    const existing = this.map.get(id);
    const history = existing ? existing.history.slice() : new Array(60).fill(0);
    history.shift();
    history.push(speed);

    const peak = existing ? Math.max(existing.peak, speed) : speed;
    this.map.set(id, { history, peak });
  }

  getHistory(id: string): ItemSpeedHistory {
    return this.map.get(id) ?? { history: new Array(60).fill(0), peak: 0 };
  }

  deleteItem(id: string) {
    this.map.delete(id);
  }
}

export const speedTracker = new SpeedHistoryTracker();
