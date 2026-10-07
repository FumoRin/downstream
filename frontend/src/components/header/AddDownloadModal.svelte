<script lang="ts">
  import {
    XIcon,
    DownloadSimpleIcon,
    LinkIcon,
    FolderIcon,
    FileTextIcon,
    CheckCircleIcon,
    WarningCircleIcon,
    CircleNotchIcon,
    ArrowLeftIcon,
    HardDriveIcon,
  } from "phosphor-svelte";
  import { ProbeURL } from "../../../wailsjs/go/main/App";
  import { formatBytes } from "../../utils/formatters";
  import type { downloader } from "../../../wailsjs/go/models";

  interface Props {
    open?: boolean;
    onClose?: () => void;
    onSubmit?: (url: string, filename: string, dir: string) => Promise<void>;
  }

  let { open = $bindable(false), onClose, onSubmit }: Props = $props();

  let step = $state<"url" | "confirm">("url");
  let url = $state("");
  let filename = $state("");
  let customDir = $state("");
  let isProbing = $state(false);
  let isSubmitting = $state(false);
  let errorMsg = $state("");
  let probeInfo = $state<downloader.TargetInfo | null>(null);

  // Step 1: Probe URL
  async function handleProbe(e?: Event) {
    if (e) e.preventDefault();
    if (!url.trim()) {
      errorMsg = "Please enter a valid URL";
      return;
    }
    errorMsg = "";
    isProbing = true;
    try {
      const info = await ProbeURL(url.trim());
      if (!info) {
        throw new Error("Unable to reach server. Please check the URL.");
      }
      probeInfo = info;
      filename = info.Filename || "";
      step = "confirm";
    } catch (err: any) {
      errorMsg =
        err?.message || String(err) || "Failed to reach server. Please check the URL.";
    } finally {
      isProbing = false;
    }
  }

  // Step 2: Final Submit
  async function handleFinalSubmit(e: SubmitEvent) {
    e.preventDefault();
    isSubmitting = true;
    errorMsg = "";
    try {
      if (onSubmit) {
        await onSubmit(url.trim(), filename.trim(), customDir.trim());
      }
      handleClose();
    } catch (err: any) {
      errorMsg = err?.message || String(err) || "Failed to start download";
    } finally {
      isSubmitting = false;
    }
  }

  function handleBack() {
    step = "url";
    errorMsg = "";
  }

  function handleClose() {
    step = "url";
    url = "";
    filename = "";
    customDir = "";
    probeInfo = null;
    errorMsg = "";
    open = false;
    onClose?.();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") handleClose();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4 animate-in fade-in duration-150 select-none"
    onclick={handleClose}
  >
    <!-- Modal Card -->
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
    <div
      class="w-full max-w-lg bg-tokyo-bgDark border border-tokyo-border rounded-lg shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <!-- Modal Header -->
      <div
        class="px-5 py-3.5 border-b border-tokyo-borderSubtle flex items-center justify-between bg-tokyo-bgDark/80"
      >
        <div
          class="flex items-center gap-2 text-tokyo-textMain font-semibold text-sm"
        >
          <div
            class="w-6 h-6 rounded bg-tokyo-teal/15 text-tokyo-teal flex items-center justify-center"
          >
            <DownloadSimpleIcon size={14} weight="bold" />
          </div>
          <span>
            {step === "url" ? "Add New Download" : "Confirm Download Details"}
          </span>
        </div>
        <button
          type="button"
          onclick={handleClose}
          title="Close (Esc)"
          class="p-1 rounded text-tokyo-textMuted hover:text-tokyo-textMain hover:bg-tokyo-bgSurface transition-colors"
        >
          <XIcon size={14} weight="bold" />
        </button>
      </div>

      <!-- STEP 1: URL Input & Probe -->
      {#if step === "url"}
        <form onsubmit={handleProbe} class="p-5 space-y-4">
          {#if errorMsg}
            <div
              class="p-3 rounded bg-tokyo-red/10 border border-tokyo-red/30 text-tokyo-red text-xs flex items-center gap-2"
            >
              <WarningCircleIcon size={16} weight="fill" class="shrink-0" />
              <span>{errorMsg}</span>
            </div>
          {/if}

          <div class="space-y-1.5">
            <label
              for="download-url"
              class="block text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
            >
              <LinkIcon size={13} class="text-tokyo-teal" />
              <span>Download Link URL <strong class="text-tokyo-red">*</strong></span>
            </label>
            <input
              id="download-url"
              type="url"
              bind:value={url}
              placeholder="Paste download URL (e.g. https://...)"
              required
              disabled={isProbing}
              class="w-full h-9 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain placeholder:text-tokyo-textMuted focus:outline-none focus:border-tokyo-teal transition-colors font-mono disabled:opacity-50"
            />
            <p class="text-[11px] text-tokyo-textMuted">
              The link will be tested to detect the file name, size, and multi-part support before downloading.
            </p>
          </div>

          <!-- Actions -->
          <div
            class="pt-2 flex items-center justify-end gap-2 border-t border-tokyo-borderSubtle"
          >
            <button
              type="button"
              onclick={handleClose}
              disabled={isProbing}
              class="h-8 px-3 rounded text-xs text-tokyo-textSub hover:bg-tokyo-bgSurface hover:text-tokyo-textMain transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isProbing || !url.trim()}
              class="h-8 px-4 rounded text-xs font-bold bg-tokyo-teal text-tokyo-bgDark hover:brightness-110 active:scale-[0.98] disabled:opacity-40 disabled:pointer-events-none transition-all flex items-center gap-2 shadow-sm"
            >
              {#if isProbing}
                <CircleNotchIcon size={14} class="animate-spin" />
                <span>Checking link...</span>
              {:else}
                <span>Check Link & Continue</span>
              {/if}
            </button>
          </div>
        </form>

      <!-- STEP 2: Metadata Review & Final Confirmation -->
      {:else if step === "confirm"}
        <form onsubmit={handleFinalSubmit} class="p-5 space-y-4">
          <!-- Active Link Status Card -->
          <div
            class="p-3 rounded bg-tokyo-teal/10 border border-tokyo-teal/30 flex items-center justify-between"
          >
            <div
              class="flex items-center gap-2 text-tokyo-teal text-xs font-semibold"
            >
              <CheckCircleIcon size={16} weight="fill" />
              <span>Server Reachable & Active</span>
            </div>
            {#if probeInfo?.SupportMultiPart}
              <span
                class="text-[10px] bg-tokyo-teal/20 text-tokyo-teal px-2 py-0.5 rounded font-mono font-medium"
              >
                Multi-part / Resume Ready
              </span>
            {:else}
              <span
                class="text-[10px] bg-tokyo-yellow/20 text-tokyo-yellow px-2 py-0.5 rounded font-mono font-medium"
              >
                Single Stream
              </span>
            {/if}
          </div>

          {#if errorMsg}
            <div
              class="p-2.5 rounded bg-tokyo-red/10 border border-tokyo-red/30 text-tokyo-red text-xs"
            >
              {errorMsg}
            </div>
          {/if}

          <!-- File Metadata Card -->
          <div class="grid grid-cols-2 gap-3 text-xs font-mono bg-tokyo-bgSurface/40 p-3 rounded border border-tokyo-borderSubtle">
            <div>
              <span class="text-[10px] uppercase text-tokyo-textMuted block font-sans font-semibold">Total Size</span>
              <span class="text-tokyo-textMain font-bold">
                {probeInfo?.TotalSize && probeInfo.TotalSize > 0
                  ? formatBytes(probeInfo.TotalSize)
                  : "Unknown / Dynamic"}
              </span>
            </div>
            <div>
              <span class="text-[10px] uppercase text-tokyo-textMuted block font-sans font-semibold">Resume Supported</span>
              <span class={probeInfo?.SupportMultiPart ? "text-tokyo-green font-bold" : "text-tokyo-textMuted"}>
                {probeInfo?.SupportMultiPart ? "Yes" : "No"}
              </span>
            </div>
          </div>

          <!-- Filename Input (Editable) -->
          <div class="space-y-1.5">
            <label
              for="confirm-filename"
              class="block text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
            >
              <FileTextIcon size={12} class="text-tokyo-teal" />
              <span>File Name</span>
            </label>
            <input
              id="confirm-filename"
              type="text"
              bind:value={filename}
              required
              class="w-full h-8 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain focus:outline-none focus:border-tokyo-teal transition-colors font-mono"
            />
          </div>

          <!-- Save Destination (Optional) -->
          <div class="space-y-1.5">
            <label
              for="confirm-dir"
              class="block text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
            >
              <FolderIcon size={12} />
              <span>Save Folder</span>
            </label>
            <input
              id="confirm-dir"
              type="text"
              bind:value={customDir}
              placeholder="Default downloads directory"
              class="w-full h-8 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain placeholder:text-tokyo-textMuted focus:outline-none focus:border-tokyo-teal transition-colors font-mono"
            />
          </div>

          <!-- Actions -->
          <div
            class="pt-2 flex items-center justify-between border-t border-tokyo-borderSubtle"
          >
            <button
              type="button"
              onclick={handleBack}
              class="h-8 px-3 rounded text-xs text-tokyo-textSub hover:bg-tokyo-bgSurface hover:text-tokyo-textMain transition-colors flex items-center gap-1"
            >
              <ArrowLeftIcon size={12} />
              <span>Change URL</span>
            </button>

            <div class="flex items-center gap-2">
              <button
                type="button"
                onclick={handleClose}
                class="h-8 px-3 rounded text-xs text-tokyo-textSub hover:bg-tokyo-bgSurface hover:text-tokyo-textMain transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting || !filename.trim()}
                class="h-8 px-4 rounded text-xs font-bold bg-tokyo-teal text-tokyo-bgDark hover:brightness-110 active:scale-[0.98] disabled:opacity-40 disabled:pointer-events-none transition-all flex items-center gap-1.5 shadow-sm"
              >
                {#if isSubmitting}
                  <CircleNotchIcon size={14} class="animate-spin" />
                  <span>Starting...</span>
                {:else}
                  <DownloadSimpleIcon size={14} weight="bold" />
                  <span>Start Download</span>
                {/if}
              </button>
            </div>
          </div>
        </form>
      {/if}
    </div>
  </div>
{/if}
