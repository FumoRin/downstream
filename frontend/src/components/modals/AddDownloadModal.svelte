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
    ClockIcon,
    CalendarBlankIcon,
  } from "phosphor-svelte";
  import { ProbeURL } from "../../../wailsjs/go/main/App";
  import { formatBytes } from "../../utils/formatters";
  import type { downloader } from "../../../wailsjs/go/models";

  interface Props {
    open?: boolean;
    onClose?: () => void;
    onSubmit?: (
      url: string,
      filename: string,
      dir: string,
      scheduledUnix?: number | null,
    ) => Promise<void>;
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
  let isScheduled = $state(false);
  let scheduledDateTime = $state(""); // formatted YYYY-MM-DDHH:mm

  // Formatter for Date Object for scheduledTime format
  function toLocalDatetimeInput(date: Date): string {
    const pad = (n: number) => String(n).padStart(2, "0");
    const year = date.getFullYear();
    const month = pad(date.getMonth() + 1);
    const day = pad(date.getDate());
    const hours = pad(date.getHours());
    const minutes = pad(date.getMinutes());
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  }

  // Quick Presets
  function applyPresetMinutes(minutes: number) {
    const d = new Date(Date.now() + minutes * 60 * 1000);
    scheduledDateTime = toLocalDatetimeInput(d);
  }

  function setTonight() {
    const d = new Date();
    d.setHours(22, 0, 0, 0);
    if (d.getTime() <= Date.now()) {
      d.setDate(d.getDate() + 1); // If already past 22:00, schedule for tomorrow 22:00
    }
    scheduledDateTime = toLocalDatetimeInput(d);
  }

  function setTomorrowMorning() {
    const d = new Date();
    d.setDate(d.getDate() + 1);
    d.setHours(9, 0, 0, 0);
    scheduledDateTime = toLocalDatetimeInput(d);
  }

  function toggleScheduled() {
    isScheduled = !isScheduled;
    if (isScheduled && !scheduledDateTime) {
      applyPresetMinutes(60); // Default to +1 hour from now
    }
  }
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
        err?.message ||
        String(err) ||
        "Failed to reach server. Please check the URL.";
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
      let scheduledUnix: number | null = null;
      if (isScheduled) {
        if (!scheduledDateTime) {
          throw new Error(
            "Please select a date and time for the scheduled download",
          );
        }
        const schedTime = new Date(scheduledDateTime).getTime();
        if (isNaN(schedTime)) {
          throw new Error("Invalid date and time selected");
        }
        if (schedTime <= Date.now()) {
          throw new Error("Schedule time must be in the future");
        }
        scheduledUnix = Math.floor(schedTime / 1000);
      }
      if (onSubmit) {
        await onSubmit(
          url.trim(),
          filename.trim(),
          customDir.trim(),
          scheduledUnix,
        );
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
    scheduledDateTime = "";
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
    class="fixed inset-0 z-50 bg-black/70 flex items-center justify-center p-4 animate-in fade-in duration-150 select-none"
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
              <span
                >Download Link URL <strong class="text-tokyo-red">*</strong
                ></span
              >
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
              The link will be tested to detect the file name, size, and
              multi-part support before downloading.
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
          <div
            class="grid grid-cols-2 gap-3 text-xs font-mono bg-tokyo-bgSurface/40 p-3 rounded border border-tokyo-borderSubtle"
          >
            <div>
              <span
                class="text-[10px] uppercase text-tokyo-textMuted block font-sans font-semibold"
                >Total Size</span
              >
              <span class="text-tokyo-textMain font-bold">
                {probeInfo?.TotalSize && probeInfo.TotalSize > 0
                  ? formatBytes(probeInfo.TotalSize)
                  : "Unknown / Dynamic"}
              </span>
            </div>
            <div>
              <span
                class="text-[10px] uppercase text-tokyo-textMuted block font-sans font-semibold"
                >Resume Supported</span
              >
              <span
                class={probeInfo?.SupportMultiPart
                  ? "text-tokyo-green font-bold"
                  : "text-tokyo-textMuted"}
              >
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

          <!-- Schedule for Later Section -->
          <div class="pt-1">
            <div
              class="p-3 rounded-lg bg-tokyo-bgSurface/50 border border-tokyo-borderSubtle space-y-3 transition-colors"
            >
              <!-- Toggle Header -->
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <div
                    class="w-6 h-6 rounded bg-tokyo-purple/15 text-tokyo-purple flex items-center justify-center shrink-0"
                  >
                    <ClockIcon size={14} weight="bold" />
                  </div>
                  <div>
                    <span
                      class="text-xs font-semibold text-tokyo-textMain block"
                    >
                      Schedule for later
                    </span>
                    <span class="text-[11px] text-tokyo-textMuted block">
                      Queue download to start automatically at a set time
                    </span>
                  </div>
                </div>

                <!-- Custom Pill Switch -->
                <button
                  type="button"
                  role="switch"
                  aria-label="Toggle schedule for later"
                  aria-checked={isScheduled}
                  onclick={toggleScheduled}
                  class="w-8 h-4 rounded-full transition-colors relative flex items-center px-0.5 cursor-pointer {isScheduled
                    ? 'bg-tokyo-purple'
                    : 'bg-tokyo-bgSurface border border-tokyo-border'}"
                >
                  <span
                    class="w-3 h-3 rounded-full bg-white transition-transform duration-150 {isScheduled
                      ? 'translate-x-3.5'
                      : 'translate-x-0'}"
                  ></span>
                </button>
              </div>

              <!-- Collapsible Schedule Options -->
              {#if isScheduled}
                <div
                  class="space-y-2.5 pt-2 border-t border-tokyo-borderSubtle animate-in fade-in duration-100"
                >
                  <!-- Quick Preset Chips -->
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span
                      class="text-[10px] text-tokyo-textMuted uppercase font-mono mr-1"
                    >
                      Presets:
                    </span>
                    <button
                      type="button"
                      onclick={() => applyPresetMinutes(15)}
                      class="px-2 py-0.5 rounded text-[11px] font-mono bg-tokyo-bgDark hover:bg-tokyo-bgHighlight text-tokyo-textSub hover:text-tokyo-textMain border border-tokyo-borderSubtle transition-colors"
                    >
                      +15m
                    </button>
                    <button
                      type="button"
                      onclick={() => applyPresetMinutes(60)}
                      class="px-2 py-0.5 rounded text-[11px] font-mono bg-tokyo-bgDark hover:bg-tokyo-bgHighlight text-tokyo-textSub hover:text-tokyo-textMain border border-tokyo-borderSubtle transition-colors"
                    >
                      +1h
                    </button>
                    <button
                      type="button"
                      onclick={setTonight}
                      class="px-2 py-0.5 rounded text-[11px] font-mono bg-tokyo-bgDark hover:bg-tokyo-bgHighlight text-tokyo-textSub hover:text-tokyo-textMain border border-tokyo-borderSubtle transition-colors"
                    >
                      Tonight (22:00)
                    </button>
                    <button
                      type="button"
                      onclick={setTomorrowMorning}
                      class="px-2 py-0.5 rounded text-[11px] font-mono bg-tokyo-bgDark hover:bg-tokyo-bgHighlight text-tokyo-textSub hover:text-tokyo-textMain border border-tokyo-borderSubtle transition-colors"
                    >
                      Tomorrow (09:00)
                    </button>
                  </div>

                  <!-- Datetime-local Input -->
                  <div class="space-y-1">
                    <label
                      for="scheduled-datetime"
                      class="block text-[11px] text-tokyo-textMuted font-mono"
                    >
                      Start Date & Time
                    </label>
                    <input
                      id="scheduled-datetime"
                      type="datetime-local"
                      bind:value={scheduledDateTime}
                      min={toLocalDatetimeInput(new Date())}
                      required={isScheduled}
                      class="w-full h-8 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain focus:outline-none focus:border-tokyo-purple transition-colors font-mono"
                    />
                  </div>
                </div>
              {/if}
            </div>
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
                class="h-8 px-4 rounded text-xs font-bold transition-all flex items-center gap-1.5 shadow-sm active:scale-[0.98] disabled:opacity-40 disabled:pointer-events-none {isScheduled
                  ? 'bg-tokyo-purple text-tokyo-bgDark hover:brightness-110'
                  : 'bg-tokyo-teal text-tokyo-bgDark hover:brightness-110'}"
              >
                {#if isSubmitting}
                  <CircleNotchIcon size={14} class="animate-spin" />
                  <span>{isScheduled ? "Scheduling..." : "Starting..."}</span>
                {:else if isScheduled}
                  <CalendarBlankIcon size={14} weight="bold" />
                  <span>Schedule Download</span>
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
