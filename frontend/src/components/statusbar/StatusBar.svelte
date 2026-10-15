<script lang="ts">
  import { onMount } from "svelte";
  import { formatBytes } from "../../utils/formatters";

  interface Props {
    sessionDownloaded?: number;
    aggregateSpeed?: number;
    rateLimitLabel?: string;
    diskStorage?: { free: number; total: number } | null;
    activeCount?: number;
    onRefreshDisk?: () => void;
  }

  let {
    sessionDownloaded = 0,
    aggregateSpeed = 0,
    rateLimitLabel = "Unlimited",
    diskStorage = null,
    activeCount = 0,
    onRefreshDisk,
  }: Props = $props();

  let isOnline = $state(
    typeof navigator !== "undefined" ? navigator.onLine : true,
  );

  onMount(() => {
    const handleOnline = () => {
      isOnline = true;
    };
    const handleOffline = () => {
      isOnline = false;
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  });

  let speedText = $derived(
    aggregateSpeed > 0 ? `${formatBytes(aggregateSpeed)}/s` : "0 B/s",
  );

  let diskText = $derived(
    diskStorage
      ? `${formatBytes(diskStorage.free)} free / ${formatBytes(diskStorage.total)}`
      : "-- free / --",
  );

  let connectionStatus = $derived.by(() => {
    if (!isOnline) {
      return {
        color: "bg-tokyo-red",
        label: "Not connected",
        description: "Network offline",
      };
    }
    return {
      color: "bg-tokyo-green",
      label: "Connected",
      description:
        activeCount > 0
          ? `${activeCount} downloading • ${speedText}`
          : "Network online",
    };
  });
</script>

<footer
  class="h-[26px] bg-tokyo-bgDark border-t border-tokyo-borderSubtle px-3 flex items-center justify-between text-[11px] font-mono select-none overflow-hidden shrink-0 text-tokyo-textMuted"
>
  <!-- Left Telemetry Stats -->
  <div class="flex items-center gap-2 min-w-0 truncate">
    <span class="whitespace-nowrap">
      Session: <span class="text-tokyo-textSub font-medium">{formatBytes(sessionDownloaded)}</span>
    </span>
    <span class="text-tokyo-textDark select-none shrink-0">•</span>
    <span class="whitespace-nowrap">
      Bandwidth: <span class="text-tokyo-textSub font-medium">{speedText} ↓</span>
    </span>
    <span class="text-tokyo-textDark select-none shrink-0 hidden sm:inline">•</span>
    <span class="whitespace-nowrap hidden sm:inline">
      Limit: <span class="text-tokyo-textSub font-medium">{rateLimitLabel}</span>
    </span>
  </div>

  <!-- Right Disk & Connection Indicator -->
  <div class="flex items-center gap-2 shrink-0 pl-2">
    <button
      type="button"
      class="hover:text-tokyo-textSub transition-colors text-left flex items-center gap-1 cursor-pointer"
      onclick={onRefreshDisk}
      title="Click to refresh disk space"
    >
      <span class="text-tokyo-textMuted">Disk:</span>
      <span class="text-tokyo-textSub font-medium">{diskText}</span>
    </button>

    <span class="text-tokyo-textDark select-none">•</span>

    <div
      class="flex items-center justify-center cursor-help"
      title="{connectionStatus.label} ({connectionStatus.description})"
    >
      <span
        class="w-2 h-2 rounded-full {connectionStatus.color} inline-block transition-colors duration-200"
      ></span>
    </div>
  </div>
</footer>
