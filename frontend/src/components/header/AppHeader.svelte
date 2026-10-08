<script lang="ts">
  import {
    CheckSquareIcon,
    PlayIcon,
    PauseIcon,
    TrashIcon,
    LightningIcon,
    XIcon,
    MagnifyingGlassIcon,
  } from "phosphor-svelte";

  interface Props {
    multiSelectMode?: boolean;
    selectedCount?: number;
    searchQuery?: string;
    speedLimitLabel?: string;
    isSpeedLimited?: boolean;
    onToggleMultiSelect?: () => void;
    onResume?: () => void;
    onPause?: () => void;
    onDelete?: () => void;
    onClearSelection?: () => void;
    onOpenSpeedLimit?: () => void;
  }

  let {
    multiSelectMode = $bindable(false),
    selectedCount = 0,
    searchQuery = $bindable(""),
    speedLimitLabel = "Speed Limit: Off",
    isSpeedLimited = false,
    onToggleMultiSelect,
    onResume,
    onPause,
    onDelete,
    onClearSelection,
    onOpenSpeedLimit,
  }: Props = $props();
</script>

<header
  class="h-11.5 bg-tokyo-bgDark border-b border-tokyo-borderSubtle px-4 flex items-center gap-3 shrink-0 z-10"
>
  <div class="w-47 flex items-center gap-2.5 mr-2">
    <!-- Need to be replaced with actual icons later -->
    <div
      class="w-6 h-6 rounded bg-tokyo-teal text-tokyo-bgDark font-black text-xs flex items-center justify-center shadow-md"
    >
      D
    </div>
    <span class="font-bold text-sm text-tokyo-textMain tracking-wide"
      >Downstream</span
    >
  </div>

  <div class="flex items-center gap-1.5">
    <button
      type="button"
      onclick={onToggleMultiSelect}
      title="Toggle multi-selection"
      class="h-7 px-2.5 rounded text-xs font-medium flex items-center gap-1.5 border transition-colors {multiSelectMode
        ? 'bg-tokyo-bgHighlight text-tokyo-teal border-tokyo-teal/50'
        : 'bg-tokyo-bgSurface text-tokyo-textSub border-tokyo-border hover:bg-tokyo-bgHover hover:text-tokyo-textMuted'}"
    >
      <CheckSquareIcon size="14" />
      <span>Multi-Select</span>
    </button>
    <button
      type="button"
      disabled={selectedCount === 0}
      onclick={onResume}
      title="Resume Selected"
      class="btn-icon"
    >
      <PlayIcon size="13" fill="currentColor" />
    </button>
    <button
      type="button"
      disabled={selectedCount === 0}
      onclick={onPause}
      title="Pause Selected"
      class="btn-icon"
    >
      <PauseIcon size="13" fill="currentColor" />
    </button>
    <button
      type="button"
      disabled={selectedCount === 0}
      onclick={onDelete}
      title="Delete Selected"
      class="btn-icon-danger"
    >
      <TrashIcon size="13" fill="currentColor" />
    </button>

    <div class="h-4 w-px bg-tokyo-borderSubtle mx-1"></div>

    <button
      type="button"
      onclick={onOpenSpeedLimit}
      title="Configure Bandwidth Limit"
      class="h-7 px-2.5 rounded text-xs font-mono flex items-center gap-1.5 border transition-colors {isSpeedLimited
        ? 'bg-tokyo-yellow/15 text-tokyo-yellow border-tokyo-yellow/40'
        : 'bg-tokyo-bgSurface text-tokyo-textSub border-tokyo-border hover:bg-tokyo-bgHover hover:text-tokyo-textMain'}"
    >
      <LightningIcon size="13" />
      <span>{speedLimitLabel}</span>
    </button>

    {#if selectedCount > 0}
      <div
        class="flex items-center gap-1.5 bg-tokyo-bgHighlight border border-tokyo-teal/40 rounded-full px-2.5 py-0.5 text-xs animate-in fade-in duration-75"
      >
        <span class="font-mono font-medium">{selectedCount} selected</span>
        <button
          type="button"
          onclick={onClearSelection}
          title="Clear Selection"
          class="hover:bg-tokyo-teal/20 rounded-full w-3.5 h-3.5 flex items-center justify-center"
        >
          <XIcon size="10" weight="bold" />
        </button>
      </div>
    {/if}
  </div>
  <div class="ml-auto relative flex items-center w-60">
    <MagnifyingGlassIcon size="13" class="absolute left-2.5 text-tokyo-textMuted pointer-events-none" />
    <input 
      type="text" 
      bind:value={searchQuery}
      placeholder="Search downloads ..."
      class="w-full h-7 pl-8 pr-2.5 bg-tokyo-bgSurface border border-tokyo-border rounded text-xs text-tokyo-textMain placeholder:text-tokyo-textMuted focus:outline-none focus:border-tokyo-teal transition-colors"
    >
  </div>
</header>
