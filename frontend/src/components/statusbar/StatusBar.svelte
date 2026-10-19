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
        activeCount > 0 ? `${activeCount} downloading` : "Network online",
    };
  });
</script>

<footer
  class="h-6.5 bg-tokyo-bgDark border-t border-tokyo-borderSubtle px-3 flex items-center justify-between text-[11px] font-mono select-none shrink-0 text-tokyo-textMuted"
>
  <!-- Left Telemetry Stats -->

  <div class="flex items-center gap-2 min-w-0 truncate">
    <span class="whitespace-nowrap flex items-center">
      <span>Session:&nbsp;</span>
      <span
        class="inline-block min-w-17.5 tabular-nums text-tokyo-textSub font-medium"
      >
        {formatBytes(sessionDownloaded)}
      </span>
    </span>

    <span class="text-tokyo-textDark select-none shrink-0">•</span>

    <span class="whitespace-nowrap flex items-center">
      <span>Bandwidth:&nbsp;</span>
      <span
        class="inline-block min-w-23.75 tabular-nums text-tokyo-textSub font-medium"
      >
        {speedText} ↓
      </span>
    </span>

    <span class="text-tokyo-textDark select-none shrink-0 hidden sm:inline"
      >•</span
    >

    <span class="whitespace-nowrap hidden sm:inline">
      Limit: <span class="text-tokyo-textSub font-medium">{rateLimitLabel}</span
      >
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

    <div class="relative group flex items-center justify-center cursor-pointer">
      <span
        class="w-2 h-2 rounded-full {connectionStatus.color} inline-block transition-colors duration-200"
      ></span>

      <!-- Custom Tokyo Night Tooltip -->
      <div
        class="pointer-events-none absolute bottom-full mb-1.5 right-0 hidden group-hover:flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-tokyo-bgDark border border-tokyo-border text-[11px] font-sans text-
  tokyo-textSub shadow-2xl whitespace-nowrap z-50 animate-in fade-in zoom-in-95 duration-100"
      >
        <span class="w-1.5 h-1.5 rounded-full {connectionStatus.color} shrink-0"
        ></span>
        <span class="font-medium text-tokyo-textMain"
          >{connectionStatus.label}</span
        >
        <span class="text-tokyo-textMuted">•</span>
        <span>{connectionStatus.description}</span>
      </div>
    </div>
  </div>
</footer>
