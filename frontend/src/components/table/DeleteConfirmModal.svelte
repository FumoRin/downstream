<script lang="ts">
  import {
    TrashIcon,
    WarningCircleIcon,
    XIcon,
  } from "phosphor-svelte";
  import type { DownloadItem } from "../../types/download";

  interface Props {
    open?: boolean;
    items?: DownloadItem[];
    onConfirm?: (deleteFiles: boolean) => Promise<void> | void;
    onClose?: () => void;
  }

  let {
    open = $bindable(false),
    items = [],
    onConfirm,
    onClose,
  }: Props = $props();

  let deleteFiles = $state(false);
  let isDeleting = $state(false);

  function handleClose() {
    if (isDeleting) return;
    deleteFiles = false;
    open = false;
    onClose?.();
  }

  async function handleConfirm() {
    isDeleting = true;
    try {
      if (onConfirm) {
        await onConfirm(deleteFiles);
      }
    } catch (e) {
      console.error("Delete confirmation failed:", e);
    } finally {
      isDeleting = false;
      deleteFiles = false;
      open = false;
      onClose?.();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      handleClose();
    } else if (e.key === "Enter" && !isDeleting) {
      e.preventDefault();
      handleConfirm();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- Modal Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4 animate-in fade-in duration-150"
    onclick={(e) => {
      if (e.target === e.currentTarget) handleClose();
    }}
  >
    <!-- Modal Card -->
    <div
      class="relative w-full max-w-md bg-tokyo-bgDark border border-tokyo-border rounded-xl shadow-2xl flex flex-col overflow-hidden text-xs text-tokyo-textMain animate-in zoom-in-95 duration-150"
    >
      <!-- Header Bar -->
      <div
        class="flex items-center justify-between px-5 py-3.5 border-b border-tokyo-borderSubtle bg-tokyo-bgDark"
      >
        <div class="flex items-center gap-2 font-semibold text-tokyo-red">
          <TrashIcon size={16} weight="bold" />
          <span class="text-sm">
            {items.length === 1 ? "Delete Download" : `Delete ${items.length} Downloads`}
          </span>
        </div>
        <button
          type="button"
          onclick={handleClose}
          disabled={isDeleting}
          class="text-tokyo-textMuted hover:text-tokyo-textMain p-1 rounded-md hover:bg-tokyo-bgSurface transition-colors disabled:opacity-50"
        >
          <XIcon size={14} />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-5 space-y-4">
        <div class="flex items-start gap-3">
          <div
            class="p-2 rounded-lg bg-tokyo-red/10 text-tokyo-red shrink-0 mt-0.5"
          >
            <WarningCircleIcon size={20} weight="fill" />
          </div>
          <div class="space-y-1 min-w-0">
            {#if items.length === 1}
              <p class="text-tokyo-textMain text-xs font-medium leading-relaxed">
                Are you sure you want to remove this download?
              </p>
              <p
                class="text-tokyo-teal font-mono text-[11px] truncate max-w-[340px] font-semibold bg-tokyo-bgSurface/70 px-2 py-1 rounded border border-tokyo-borderSubtle"
              >
                {items[0]?.filename}
              </p>
            {:else}
              <p class="text-tokyo-textMain text-xs font-medium leading-relaxed">
                Are you sure you want to remove these <strong class="text-tokyo-red">{items.length}</strong> downloads?
              </p>
              <div
                class="bg-tokyo-bgSurface/70 p-2 rounded border border-tokyo-borderSubtle max-h-24 overflow-y-auto space-y-1"
              >
                {#each items.slice(0, 4) as item}
                  <div class="text-[11px] font-mono text-tokyo-textSub truncate">
                    • {item.filename}
                  </div>
                {/each}
                {#if items.length > 4}
                  <div class="text-[10px] text-tokyo-textMuted italic">
                    + {items.length - 4} more
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        </div>

        <!-- Delete Files Checkbox Option -->
        <label
          class="flex items-start gap-3 p-3 rounded-lg bg-tokyo-bgSurface/40 border border-tokyo-borderSubtle hover:border-tokyo-border cursor-pointer transition-colors group select-none"
        >
          <input
            type="checkbox"
            bind:checked={deleteFiles}
            class="mt-0.5 rounded bg-tokyo-bgDark border-tokyo-border text-tokyo-red focus:ring-tokyo-red/40 focus:ring-offset-0 cursor-pointer"
          />
          <div class="space-y-0.5">
            <span
              class="text-xs font-medium text-tokyo-textMain group-hover:text-white transition-colors"
            >
              Also delete downloaded file(s) from disk
            </span>
            <p class="text-[11px] text-tokyo-textMuted leading-relaxed">
              Permanently removes downloaded files and temporary part data from your computer.
            </p>
          </div>
        </label>
      </div>

      <!-- Action Footer -->
      <div
        class="px-5 py-3 border-t border-tokyo-borderSubtle bg-tokyo-bgDark/50 flex justify-end gap-2.5"
      >
        <button
          type="button"
          onclick={handleClose}
          disabled={isDeleting}
          class="px-3 py-1.5 rounded-lg border border-tokyo-border text-tokyo-textSub hover:bg-tokyo-bgSurface hover:text-tokyo-textMain transition-colors text-xs font-medium disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="button"
          onclick={handleConfirm}
          disabled={isDeleting}
          class="px-4 py-1.5 rounded-lg bg-tokyo-red hover:bg-tokyo-red/90 text-white font-medium transition-colors text-xs flex items-center gap-1.5 shadow-sm shadow-tokyo-red/20 disabled:opacity-50"
        >
          <TrashIcon size={13} weight="bold" />
          <span>{isDeleting ? "Deleting..." : "Delete"}</span>
        </button>
      </div>
    </div>
  </div>
{/if}
