<script lang="ts">
  import type { DownloadItem, PartState } from "../../types/download";
  import { formatBytes } from "../../utils/formatters";
  import {
    StackIcon,
    ArrowsMergeIcon,
    CheckCircleIcon,
    CircleNotchIcon,
    PauseCircleIcon,
    InfoIcon,
  } from "phosphor-svelte";

  interface Props {
    item: DownloadItem;
  }

  let { item }: Props = $props();

  let parts = $derived(item.parts || []);
  let sortedParts = $derived(
    [...parts].sort((a, b) => a.startByte - b.startByte),
  );
  let isMultiPart = $derived(sortedParts.length > 1);

  function getPartSize(part: PartState): number {
    return Math.max(1, part.endByte - part.startByte + 1);
  }

  function getPartPercentage(part: PartState): number {
    const size = getPartSize(part);
    if (size <= 0) return 0;
    return Math.min(100, Math.max(0, (part.currentByte / size) * 100));
  }

  function getPartStatus(
    part: PartState,
  ): "done" | "downloading" | "paused" | "idle" {
    if (item.status === 3 || part.currentByte >= getPartSize(part)) {
      return "done";
    }
    if (item.status === 1) {
      return "downloading";
    }
    if (item.status === 2) {
      return "paused";
    }
    return "idle";
  }
</script>

<div class="h-full flex flex-col p-3.5 gap-2.5 bg-tokyo-bgDark text-tokyo-textMain select-none overflow-hidden">
  <!-- Top Summary Strip -->
  <div class="flex items-center justify-between gap-3 shrink-0">
    <div class="flex items-center gap-2 min-w-0">
      {#if isMultiPart}
        <div class="w-6 h-6 rounded bg-tokyo-teal/10 text-tokyo-teal flex items-center justify-center shrink-0">
          <StackIcon size={14} weight="bold" />
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs font-bold text-tokyo-textMain">Multi-Part Stream</span>
          <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-tokyo-bgSurface border border-tokyo-border text-tokyo-teal font-medium">
            {sortedParts.length} Connections
          </span>
        </div>
      {:else}
        <div class="w-6 h-6 rounded bg-tokyo-blue/10 text-tokyo-blue flex items-center justify-center shrink-0">
          <ArrowsMergeIcon size={14} weight="bold" />
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs font-bold text-tokyo-textMain">Single Connection</span>
          <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-tokyo-bgSurface border border-tokyo-border text-tokyo-textMuted font-medium">
            1 Stream
          </span>
        </div>
      {/if}
    </div>

    <div class="flex items-center gap-2 font-mono text-[11px] text-tokyo-textMuted tabular-nums shrink-0">
      <span class="text-tokyo-textSub font-medium">{formatBytes(item.currentSize)}</span>
      <span>/</span>
      <span>{formatBytes(item.totalSize)}</span>
      <span class="text-tokyo-teal font-semibold ml-1">
        {item.status === 3 ? "100%" : `${(item.percentage || 0).toFixed(1)}%`}
      </span>
    </div>
  </div>

  <!-- Segmented Chunk Visualizer Track (IDM Style) -->
  <div class="shrink-0 space-y-1">
    <div class="h-3.5 w-full bg-tokyo-bgSurface border border-tokyo-borderSubtle rounded-sm flex overflow-hidden shadow-inner">
      {#if isMultiPart}
        {#each sortedParts as part, idx (part.id || idx)}
          {@const partSize = getPartSize(part)}
          {@const widthPct = item.totalSize > 0 ? (partSize / item.totalSize) * 100 : 100 / sortedParts.length}
          {@const pct = getPartPercentage(part)}
          {@const status = getPartStatus(part)}

          <div
            class="relative h-full bg-tokyo-bgDark/50 border-r border-tokyo-bgDark last:border-r-0 group/seg transition-colors"
            style="width: {widthPct}%;"
            title="{part.workerID || part.workerId || `Worker ${idx + 1}`}: {formatBytes(part.currentByte)} / {formatBytes(partSize)} ({pct.toFixed(1)}%)"
          >
            <div
              class="h-full transition-[width] duration-150 ease-out {status === 'done'
                ? 'bg-tokyo-green'
                : status === 'paused'
                ? 'bg-tokyo-orange'
                : 'bg-tokyo-teal'}"
              style="width: {pct}%;"
            ></div>
          </div>
        {/each}
      {:else}
        <!-- Single Stream Track -->
        <div class="w-full h-full bg-tokyo-bgDark/50 relative">
          <div
            class="h-full bg-tokyo-teal transition-[width] duration-150 ease-out"
            style="width: {item.status === 3 ? 100 : item.percentage || 0}%;"
          ></div>
        </div>
      {/if}
    </div>
  </div>

  <!-- Parts Breakdown Content -->
  {#if isMultiPart}
    <!-- Scrollable Table for Multi-Part Connections -->
    <div class="flex-1 min-h-0 overflow-y-auto border border-tokyo-borderSubtle rounded bg-tokyo-bgSurface/20">
      <table class="w-full text-left border-collapse text-xs select-none">
        <thead class="sticky top-0 z-10 text-[10px] uppercase font-mono text-tokyo-textMuted border-b border-tokyo-borderSubtle shadow-sm">
          <tr>
            <th class="py-1.5 px-3 font-medium bg-tokyo-bgDark">Worker</th>
            <th class="py-1.5 px-3 font-medium bg-tokyo-bgDark">Byte Range</th>
            <th class="py-1.5 px-3 font-medium bg-tokyo-bgDark">Downloaded</th>
            <th class="py-1.5 px-3 font-medium bg-tokyo-bgDark">Progress</th>
            <th class="py-1.5 px-3 font-medium text-right bg-tokyo-bgDark">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-tokyo-borderSubtle/60 font-mono text-[11px]">
          {#each sortedParts as part, idx (part.id || idx)}
            {@const partSize = getPartSize(part)}
            {@const pct = getPartPercentage(part)}
            {@const status = getPartStatus(part)}
            <tr class="hover:bg-tokyo-bgSurface/30 transition-colors">
              <td class="py-1.5 px-3 font-medium text-tokyo-textMain flex items-center gap-1.5">
                {#if status === "done"}
                  <span class="w-1.5 h-1.5 rounded-full bg-tokyo-green shrink-0"></span>
                {:else if status === "downloading"}
                  <span class="w-1.5 h-1.5 rounded-full bg-tokyo-teal animate-pulse shrink-0"></span>
                {:else if status === "paused"}
                  <span class="w-1.5 h-1.5 rounded-full bg-tokyo-orange shrink-0"></span>
                {:else}
                  <span class="w-1.5 h-1.5 rounded-full bg-tokyo-textMuted shrink-0"></span>
                {/if}
                <span>{part.workerID || part.workerId || `Worker ${idx + 1}`}</span>
              </td>
              <td class="py-1.5 px-3 text-tokyo-textMuted">
                {formatBytes(part.startByte)} – {formatBytes(part.endByte)}
              </td>
              <td class="py-1.5 px-3 text-tokyo-textSub tabular-nums">
                {formatBytes(part.currentByte)} / {formatBytes(partSize)}
              </td>
              <td class="py-1.5 px-3">
                <div class="flex items-center gap-2">
                  <div class="w-20 h-1.5 bg-tokyo-bgSurface rounded-full overflow-hidden shrink-0">
                    <div
                      class="h-full rounded-full transition-[width] duration-150 ease-out {status === 'done'
                        ? 'bg-tokyo-green'
                        : status === 'paused'
                        ? 'bg-tokyo-orange'
                        : 'bg-tokyo-teal'}"
                      style="width: {pct}%;"
                    ></div>
                  </div>
                  <span class="text-[10px] text-tokyo-textMuted tabular-nums">
                    {pct.toFixed(1)}%
                  </span>
                </div>
              </td>
              <td class="py-1.5 px-3 text-right">
                {#if status === "done"}
                  <span class="inline-flex items-center gap-1 text-[10px] font-sans font-medium text-tokyo-green bg-tokyo-green/10 px-1.5 py-0.5 rounded">
                    <CheckCircleIcon size={11} weight="fill" />
                    Done
                  </span>
                {:else if status === "downloading"}
                  <span class="inline-flex items-center gap-1 text-[10px] font-sans font-medium text-tokyo-teal bg-tokyo-teal/10 px-1.5 py-0.5 rounded">
                    <CircleNotchIcon size={11} class="animate-spin" />
                    Active
                  </span>
                {:else if status === "paused"}
                  <span class="inline-flex items-center gap-1 text-[10px] font-sans font-medium text-tokyo-orange bg-tokyo-orange/10 px-1.5 py-0.5 rounded">
                    <PauseCircleIcon size={11} weight="fill" />
                    Paused
                  </span>
                {:else}
                  <span class="text-[10px] font-sans text-tokyo-textMuted">Idle</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <!-- Single Connection / Non-segmented State -->
    <div class="flex-1 min-h-0 flex items-center justify-between p-3 rounded border border-tokyo-borderSubtle bg-tokyo-bgSurface/20 text-xs">
      <div class="space-y-1">
        <div class="flex items-center gap-1.5 font-medium text-tokyo-textSub">
          <InfoIcon size={14} class="text-tokyo-cyan shrink-0" />
          <span>Single-Stream Download</span>
        </div>
        <p class="text-[11px] text-tokyo-textMuted max-w-lg leading-relaxed">
          {#if item.status === 0 || item.status === 5}
            Connection parts will be initialized once the download starts.
          {:else}
            This download is operating over a single connection because the remote server does not support byte-range requests (HTTP 206 Partial Content) or the file size is under the multipart threshold.
          {/if}
        </p>
      </div>

      <div class="text-right font-mono text-[11px] space-y-0.5 shrink-0 pl-4 border-l border-tokyo-borderSubtle">
        <div class="text-tokyo-textMuted">Stream Progress</div>
        <div class="text-tokyo-teal font-bold text-sm">
          {item.status === 3 ? "100%" : `${(item.percentage || 0).toFixed(1)}%`}
        </div>
        <div class="text-[10px] text-tokyo-textMuted">
          {formatBytes(item.currentSize)} / {formatBytes(item.totalSize)}
        </div>
      </div>
    </div>
  {/if}
</div>
