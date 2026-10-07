<script lang="ts">
  import { CopyIcon, CheckIcon } from "phosphor-svelte";
  import type { DownloadItem } from "../../types/download";
  import FileTypeIcon from "../table/FileTypeIcons.svelte";
  import MiniSpeedGraph from "./MiniSpeedGraph.svelte";
  import {
    formatBytes,
    formatSpeed,
    formatETA,
    formatDate,
  } from "../../utils/formatters";

  interface Props {
    item: DownloadItem;
  }

  let { item }: Props = $props();

  let copiedPath = $state(false);
  let copiedUrl = $state(false);

  function copyToClipboard(text: string, type: "path" | "url") {
    if (!text) return;
    navigator.clipboard.writeText(text);
    if (type === "path") {
      copiedPath = true;
      setTimeout(() => (copiedPath = false), 1500);
    } else {
      copiedUrl = true;
      setTimeout(() => (copiedUrl = false), 1500);
    }
  }

  const statusMeta = $derived.by(() => {
    switch (item.status) {
      case 1:
        return {
          label: "Active",
          barColor: "bg-tokyo-teal",
          badgeColor: "text-tokyo-teal bg-tokyo-teal/10",
        };
      case 2:
        return {
          label: "Paused",
          barColor: "bg-tokyo-orange",
          badgeColor: "text-tokyo-orange bg-tokyo-orange/10",
        };
      case 3:
        return {
          label: "Completed",
          barColor: "bg-tokyo-green",
          badgeColor: "text-tokyo-green bg-tokyo-green/10",
        };
      case 4:
        return {
          label: "Failed",
          barColor: "bg-tokyo-red",
          badgeColor: "text-tokyo-red bg-tokyo-red/10",
        };
      default:
        return {
          label: "Queued",
          barColor: "bg-tokyo-purple",
          badgeColor: "text-tokyo-purple bg-tokyo-purple/10",
        };
    }
  });
</script>

<div class="flex items-center gap-6 h-full px-4 py-2 select-none">
  <!-- Left: File Graphic Badge (70x70) -->
  <div
    class="w-[70px] h-[70px] rounded-lg bg-tokyo-bgSurface border border-tokyo-border flex items-center justify-center shrink-0 shadow-inner"
  >
    <FileTypeIcon
      category={item.category}
      size={32}
      wrapperClass="w-full h-full rounded-lg flex items-center justify-center shrink-0"
    />
  </div>

  <!-- Center: Metadata & Progress Info -->
  <div class="flex-1 min-w-0 space-y-1.5">
    <!-- Row 1: Filename + Status Pill -->
    <div class="flex items-center gap-2.5">
      <span
        class="font-bold text-sm text-tokyo-textMain truncate leading-tight"
      >
        {item.filename}
      </span>
      <span
        class="text-[10px] font-medium px-2 py-0.5 rounded shrink-0 {statusMeta.badgeColor}"
      >
        ● {statusMeta.label}
      </span>
    </div>

    <!-- Row 2: Destination Path with Inline Copy -->
    <div
      class="flex items-center gap-1.5 text-xs text-tokyo-textMuted font-mono"
    >
      <span class="truncate">{item.saveDir || "~/Downloads"}</span>
      <button
        type="button"
        onclick={() => copyToClipboard(item.saveDir || "~/Downloads", "path")}
        title="Copy download path"
        class="hover:text-tokyo-textMain p-0.5 rounded transition-colors"
      >
        {#if copiedPath}
          <CheckIcon size={12} class="text-tokyo-green" />
        {:else}
          <CopyIcon size={12} />
        {/if}
      </button>
    </div>

    <!-- Row 3: Progress Bar -->
    <div class="flex items-center gap-3">
      <div
        class="flex-1 h-2 bg-tokyo-bgSurface rounded-full overflow-hidden border border-tokyo-borderSubtle"
      >
        <div
          class="h-full rounded-full transition-all duration-200 {statusMeta.barColor}"
          style="width: {item.status === 3 ? 100 : item.percentage || 0}%"
        ></div>
      </div>
      <span class="font-mono font-bold text-xs text-tokyo-textMain shrink-0">
        {item.status === 3 ? "100%" : `${(item.percentage || 0).toFixed(1)}%`}
      </span>
    </div>

    <!-- Row 4: Stats Line -->
    <div
      class="text-[11px] font-mono text-tokyo-textSub flex items-center gap-2 truncate"
    >
      <span
        >Downloaded: <strong class="text-tokyo-textMain"
          >{formatBytes(item.currentSize || 0)}</strong
        > / {formatBytes(item.totalSize)}</span
      >
      <span class="text-tokyo-textMuted">•</span>
      <span
        >Speed: <strong class="text-tokyo-teal"
          >{item.status === 1 ? formatSpeed(item.speed) : "-"}</strong
        ></span
      >
      <span class="text-tokyo-textMuted">•</span>
      <span>ETA: {item.status === 3 ? "Done" : formatETA(item.eta)}</span>
      <span class="text-tokyo-textMuted">•</span>
      <span>Added: {formatDate(item.createdAt)}</span>
    </div>

    <!-- Row 5: File URL (Truncated at end with copy button) -->
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div
      class="flex items-center gap-1.5 text-[11px] font-mono text-tokyo-textMuted min-w-0"
    >
      <span>File:</span>
      <span
        class="text-tokyo-cyan hover:underline cursor-pointer truncate max-w-[450px]"
        title={item.url}
        onclick={() => copyToClipboard(item.url, "url")}
      >
        {item.url}
      </span>
      <button
        type="button"
        onclick={() => copyToClipboard(item.url, "url")}
        title="Copy URL"
        class="hover:text-tokyo-textMain p-0.5 rounded transition-colors"
      >
        {#if copiedUrl}
          <CheckIcon size={12} class="text-tokyo-green" />
        {:else}
          <CopyIcon size={12} />
        {/if}
      </button>
    </div>
  </div>

  <!-- Right: Docked Speed Graph -->
  <MiniSpeedGraph speed={item.speed} status={item.status} itemId={item.id} />
</div>
