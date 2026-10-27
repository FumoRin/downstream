<script lang="ts">
  import { FilesIcon, PlayIcon, PauseIcon, TrashIcon } from "phosphor-svelte";
  import type { DownloadItem } from "../../types/download";
  import MiniSpeedGraph from "./MiniSpeedGraph.svelte";
  import { formatBytes, formatSpeed } from "../../utils/formatters";

  interface Props {
    selectedItems: DownloadItem[];
    onResumeBatch?: (ids: string[]) => void;
    onPauseBatch?: (ids: string[]) => void;
    onDeleteBatch?: (ids: string[]) => void;
  }

  let {
    selectedItems = [],
    onResumeBatch,
    onPauseBatch,
    onDeleteBatch,
  }: Props = $props();

  const totalBytes = $derived(
    selectedItems.reduce((acc, item) => acc + (item.totalSize || 0), 0),
  );
  const totalDownloaded = $derived(
    selectedItems.reduce((acc, item) => acc + (item.currentSize || 0), 0),
  );
  const totalSpeed = $derived(
    selectedItems.reduce(
      (acc, item) => acc + (item.status === 1 ? item.speed || 0 : 0),
      0,
    ),
  );
  const aggregatePercentage = $derived(
    totalBytes > 0 ? Math.min(100, (totalDownloaded / totalBytes) * 100) : 0,
  );

  const activeItems = $derived(selectedItems.filter((i) => i.status === 1));
  const pausedItems = $derived(selectedItems.filter((i) => i.status === 2));
  const doneItems = $derived(selectedItems.filter((i) => i.status === 3));
  const failedItems = $derived(selectedItems.filter((i) => i.status === 4));
  const queuedItems = $derived(
    selectedItems.filter((i) => i.status === 0 || i.status === 5),
  );

  const resumableIds = $derived([
    ...pausedItems.map((i) => i.id),
    ...queuedItems.map((i) => i.id),
    ...failedItems.map((i) => i.id),
  ]);
  const activeIds = $derived(activeItems.map((i) => i.id));
  const allIds = $derived(selectedItems.map((i) => i.id));

  const batchKey = $derived(`batch-${allIds.slice().sort().join("-")}`);
</script>

<div class="flex items-center gap-6 h-full px-5 py-3 select-none">
  <!-- Left: Stacked Files Vector Badge (80x80) -->
  <div
    class="w-20 h-20 rounded-xl bg-tokyo-bgSurface border border-tokyo-border flex items-center justify-center shrink-0 shadow-inner"
  >
    <FilesIcon size={38} class="text-tokyo-purple" />
  </div>

  <!-- Center: Batch Stats, Progress & Action Chips -->
  <div class="flex-1 min-w-0 space-y-2.5">
    <!-- Row 1: Batch Header Title + Status Distribution Pills -->
    <div class="flex items-center gap-2 flex-wrap">
      <span class="font-bold text-sm text-tokyo-textMain">
        {selectedItems.length} downloads selected
      </span>

      <div class="flex items-center gap-1.5 text-[10px] font-medium">
        {#if activeItems.length > 0}
          <span class="px-2 py-0.5 rounded bg-tokyo-teal/10 text-tokyo-teal">
            ● {activeItems.length} Active
          </span>
        {/if}
        {#if pausedItems.length > 0}
          <span
            class="px-2 py-0.5 rounded bg-tokyo-orange/10 text-tokyo-orange"
          >
            ● {pausedItems.length} Paused
          </span>
        {/if}
        {#if doneItems.length > 0}
          <span class="px-2 py-0.5 rounded bg-tokyo-green/10 text-tokyo-green">
            ● {doneItems.length} Done
          </span>
        {/if}
        {#if failedItems.length > 0}
          <span class="px-2 py-0.5 rounded bg-tokyo-red/10 text-tokyo-red">
            ● {failedItems.length} Failed
          </span>
        {/if}
      </div>
    </div>

    <!-- Row 2: Batch Action Chips -->
    <div class="flex items-center gap-2">
      {#if resumableIds.length > 0}
        <button
          type="button"
          onclick={() => onResumeBatch?.(resumableIds)}
          class="px-2.5 py-1 rounded text-[11px] font-medium flex items-center gap-1.5 bg-tokyo-teal/10 text-tokyo-teal hover:bg-tokyo-teal/20 transition-colors"
        >
          <PlayIcon size={12} weight="bold" />
          <span>Resume ({resumableIds.length})</span>
        </button>
      {/if}

      {#if activeIds.length > 0}
        <button
          type="button"
          onclick={() => onPauseBatch?.(activeIds)}
          class="px-2.5 py-1 rounded text-[11px] font-medium flex items-center gap-1.5 bg-tokyo-orange/10 text-tokyo-orange hover:bg-tokyo-orange/20 transition-colors"
        >
          <PauseIcon size={12} weight="bold" />
          <span>Pause ({activeIds.length})</span>
        </button>
      {/if}

      <button
        type="button"
        onclick={() => onDeleteBatch?.(allIds)}
        class="px-2.5 py-1 rounded text-[11px] font-medium flex items-center gap-1.5 bg-tokyo-red/10 text-tokyo-red hover:bg-tokyo-red/20 transition-colors"
      >
        <TrashIcon size={12} weight="bold" />
        <span>Delete Selected ({allIds.length})</span>
      </button>
    </div>

    <!-- Row 3: Aggregate Progress Bar (Solid color, no gradient) -->
    <div class="flex items-center gap-3">
      <div
        class="flex-1 h-2.5 bg-tokyo-bgSurface rounded-full overflow-hidden border border-tokyo-borderSubtle"
      >
        <div
          class="h-full rounded-full transition-[width] duration-150 ease-out bg-tokyo-teal"
          style="width: {aggregatePercentage}%"
        ></div>
      </div>
      <span class="font-mono font-bold text-xs text-tokyo-textMain shrink-0">
        {aggregatePercentage.toFixed(1)}%
      </span>
    </div>

    <!-- Row 4: Aggregate Metrics -->
    <div
      class="text-[11px] font-mono text-tokyo-textSub flex items-center gap-2 truncate"
    >
      <span class="shrink-0">
        Downloaded: <strong class="text-tokyo-textMain"
          >{formatBytes(totalDownloaded)}</strong
        >
        / {formatBytes(totalBytes)}
      </span>
      <span class="text-tokyo-borderSubtle hidden sm:inline">•</span>
      <span class="hidden sm:inline shrink-0">
        Combined Speed: <strong class="text-tokyo-teal"
          >{formatSpeed(totalSpeed)}</strong
        >
      </span>
      <span class="text-tokyo-borderSubtle hidden lg:inline">•</span>
      <span class="hidden lg:inline shrink-0">
        Active Transfers: {activeItems.length}
      </span>
    </div>
  </div>

  <!-- Right: Docked Speed Graph with Combined Speed -->
  <MiniSpeedGraph
    speed={totalSpeed}
    status={activeItems.length > 0 ? 1 : 0}
    itemId={batchKey}
  />
</div>
