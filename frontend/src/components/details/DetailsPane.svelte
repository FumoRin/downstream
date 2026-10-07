<script lang="ts">
  import {
    CaretUpIcon,
    CaretDownIcon,
    CursorClickIcon,
    StackIcon,
  } from "phosphor-svelte";
  import type { DownloadItem } from "../../types/download";
  import GeneralView from "./GeneralView.svelte";

  interface Props {
    selectedItem: DownloadItem | null;
    selectedCount: number;
  }

  let { selectedItem, selectedCount = 0 }: Props = $props();

  let activeTab = $state<"general" | "parts">("general");
  let isCollapsed = $state(false);
</script>

<div
  class="bg-tokyo-bgDark border-t border-tokyo-borderSubtle flex flex-col shrink-0 transition-all duration-200 select-none {isCollapsed
    ? 'h-8'
    : 'h-[195px]'}"
>
  <!-- Pane Header Bar -->
  <div
    class="h-8 px-4 border-b border-tokyo-borderSubtle flex items-center justify-between bg-tokyo-bgDark/80"
  >
    <div
      class="flex items-center gap-2 text-xs font-medium text-tokyo-textMuted"
    >
      {#if selectedCount === 1 && selectedItem}
        <span class="text-tokyo-textMain font-semibold truncate max-w-[300px]">
          {selectedItem.filename}
        </span>
      {:else if selectedCount > 1}
        <span class="text-tokyo-textMain font-semibold">
          {selectedCount} downloads selected
        </span>
      {:else}
        <span>Details Inspector</span>
      {/if}
    </div>

    <!-- Right Controls: Tabs + Collapse Toggle -->
    <div class="flex items-center gap-2">
      <!-- 2-Tab Segmented Control (Visible when 1 item selected) -->
      {#if selectedCount === 1 && !isCollapsed}
        <div
          class="flex items-center bg-tokyo-bgSurface p-0.5 rounded border border-tokyo-borderSubtle"
        >
          <button
            type="button"
            onclick={() => (activeTab = "general")}
            class="px-2.5 py-0.5 rounded text-[11px] font-medium transition-colors {activeTab ===
            'general'
              ? 'bg-tokyo-bgHighlight text-tokyo-textMain shadow-sm'
              : 'text-tokyo-textMuted hover:text-tokyo-textMain'}"
          >
            General
          </button>
          <button
            type="button"
            onclick={() => (activeTab = "parts")}
            class="px-2.5 py-0.5 rounded text-[11px] font-medium transition-colors {activeTab ===
            'parts'
              ? 'bg-tokyo-bgHighlight text-tokyo-textMain shadow-sm'
              : 'text-tokyo-textMuted hover:text-tokyo-textMain'}"
          >
            Chunks & Parts
          </button>
        </div>
      {/if}

      <!-- Collapse / Expand Button -->
      <button
        type="button"
        onclick={() => (isCollapsed = !isCollapsed)}
        title={isCollapsed ? "Expand Details" : "Collapse Details"}
        class="p-1 rounded hover:bg-tokyo-bgSurface text-tokyo-textMuted hover:text-tokyo-textMain transition-colors"
      >
        {#if isCollapsed}
          <CaretUpIcon size={14} />
        {:else}
          <CaretDownIcon size={14} />
        {/if}
      </button>
    </div>
  </div>

  <!-- Pane Body (Hidden when collapsed) -->
  {#if !isCollapsed}
    <div class="flex-1 min-h-0 overflow-hidden bg-tokyo-bgDark">
      {#if selectedCount === 0}
        <!-- Zero Selection State -->
        <div
          class="h-full flex flex-col items-center justify-center text-tokyo-textMuted gap-2"
        >
          <CursorClickIcon size={24} class="opacity-40" />
          <p class="text-xs">Select a download to view details and telemetry</p>
        </div>
      {:else if selectedCount > 1}
        <!-- Multi-Selection Batch State -->
        <div
          class="h-full flex flex-col items-center justify-center text-tokyo-textMuted gap-2"
        >
          <StackIcon size={24} class="text-tokyo-teal opacity-60" />
          <p class="text-xs font-semibold text-tokyo-textMain">
            {selectedCount} downloads selected
          </p>
          <p class="text-[11px]">
            Use header actions to resume, pause, or delete in bulk
          </p>
        </div>
      {:else if selectedItem}
        <!-- Single Item Selected -->
        {#if activeTab === "general"}
          <GeneralView item={selectedItem} />
        {:else}
          <div
            class="h-full flex items-center justify-center text-tokyo-textMuted text-xs"
          >
            Chunks & Parts view (coming up next)
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>
