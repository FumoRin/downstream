<script lang="ts">
  import {
    XIcon,
    GearIcon,
    FolderIcon,
    HardDriveIcon,
    SlidersIcon,
    InfoIcon,
    CheckIcon,
    WarningCircleIcon,
    CircleNotchIcon,
    CaretDownIcon,
  } from "phosphor-svelte";
  import {
    GetSettings,
    UpdateSettings,
    SelectDirectory,
  } from "../../../wailsjs/go/main/App";

  interface Props {
    open?: boolean;
    showStatusBar?: boolean;
    onClose?: () => void;
    onSaved?: () => void;
  }

  let {
    open = $bindable(false),
    showStatusBar = $bindable(true),
    onClose,
    onSaved,
  }: Props = $props();

  // Active Tab
  let activeTab = $state<"downloads" | "about">("downloads");

  // Form Fields
  let downloadDir = $state("");
  let maxConcurrency = $state(3);
  let partsPerDownload = $state(4);
  let minMultiPartMB = $state(5);
  let showStatusBarLocal = $state(true);

  // Status & Feedback States
  let initialDownloadDir = $state("");
  let isLoading = $state(false);
  let isSaving = $state(false);
  let errorMsg = $state("");
  let saveSuccess = $state(false);

  // Download Speed limit
  let isSpeedLimited = $state(false);
  let speedLimitValue = $state(5);
  let speedLimitUnit = $state<"MB" | "KB">("MB");

  // Load existing settings whenever modal opens
  $effect(() => {
    if (open) {
      loadCurrentSettings();
    }
  });

  async function loadCurrentSettings() {
    isLoading = true;
    errorMsg = "";
    saveSuccess = false;
    showStatusBarLocal = showStatusBar;
    try {
      const s = await GetSettings();
      if (s) {
        initialDownloadDir = s.DownloadDir || "";
        downloadDir = s.DownloadDir || "";
        maxConcurrency = s.MaxConcurrencyDownload || 3;
        partsPerDownload = s.PartsPerDownload || 4;
        minMultiPartMB = Math.max(
          1,
          Math.round(s.MinMultiPartDownload || 5242880) / (1024 * 1024),
        );

        const limitBytes = s.SpeedLimit || 0;
        if (limitBytes > 0) {
          isSpeedLimited = true;
          if (limitBytes >= 1024 * 1024 && limitBytes % (1024 * 1024) === 0) {
            speedLimitValue = limitBytes / (1024 * 1024);
            speedLimitUnit = "MB";
          } else {
            speedLimitValue = Math.round(limitBytes / 1024);
            speedLimitUnit = "KB";
          }
        } else {
          isSpeedLimited = false;
          const savedLimit =
            typeof localStorage !== "undefined"
              ? Number(localStorage.getItem("downstream_last_speed_limit")) || 0
              : 0;
          if (savedLimit > 0) {
            if (savedLimit >= 1024 * 1024 && savedLimit % (1024 * 1024) === 0) {
              speedLimitValue = savedLimit / (1024 * 1024);
              speedLimitUnit = "MB";
            } else {
              speedLimitValue = Math.round(savedLimit / 1024);
              speedLimitUnit = "KB";
            }
          } else {
            speedLimitValue = 5;
            speedLimitUnit = "MB";
          }
        }
      }
    } catch (err: any) {
      errorMsg = err?.message || "Failed to load current settings";
    } finally {
      isLoading = false;
    }
  }

  async function handleBrowseFolder() {
    try {
      const selected = await SelectDirectory();
      if (selected) {
        downloadDir = selected;
      }
    } catch (err: any) {
      console.error("Failed to select directory:", err);
    }
  }

  async function handleSave(e?: Event) {
    if (e) e.preventDefault();

    const finalDownloadDir = downloadDir.trim() || initialDownloadDir;
    if (!finalDownloadDir) {
      errorMsg = "Default download directory cannot be empty";
      return;
    }

    if (maxConcurrency < 1 || maxConcurrency > 10) {
      errorMsg = "Concurrent downloads must be between 1 and 10";
      return;
    }
    if (partsPerDownload < 1 || partsPerDownload > 16) {
      errorMsg = "Connections per download must be between 1 and 16";
      return;
    }
    if (minMultiPartMB < 1) {
      errorMsg = "Minimum multi-part size must be at least 1 MB";
      return;
    }

    const calculateSpeedLimit = isSpeedLimited
      ? Number(speedLimitValue) * (speedLimitUnit === "MB" ? 1024 * 1024 : 1024)
      : 0;

    if (calculateSpeedLimit > 0) {
      localStorage.setItem(
        "downstream_last_speed_limit",
        String(calculateSpeedLimit),
      );
    }

    isSaving = true;
    errorMsg = "";
    try {
      await UpdateSettings({
        DownloadDir: finalDownloadDir,
        MaxConcurrencyDownload: Number(maxConcurrency),
        PartsPerDownload: Number(partsPerDownload),
        MinMultiPartDownload: Number(minMultiPartMB) * 1024 * 1024,
        SpeedLimit: calculateSpeedLimit,
      });

      showStatusBar = showStatusBarLocal;
      localStorage.setItem(
        "downstream_show_statusbar",
        String(showStatusBarLocal),
      );

      saveSuccess = true;
      onSaved?.();

      setTimeout(() => {
        handleClose();
      }, 500);
    } catch (err: any) {
      errorMsg = err?.message || "Failed to save settings";
    } finally {
      isSaving = false;
    }
  }

  function handleClose() {
    if (isSaving) return;
    open = false;
    errorMsg = "";
    saveSuccess = false;
    onClose?.();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      handleClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 animate-in fade-in duration-150 select-none"
    onclick={(e) => {
      if (e.target === e.currentTarget) handleClose();
    }}
  >
    <!-- Modal Card -->
    <div
      class="relative w-full max-w-lg bg-tokyo-bgDark border border-tokyo-border rounded-xl shadow-2xl flex flex-col overflow-hidden text-xs text-tokyo-textMain animate-in zoom-in-95 duration-150"
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between px-5 py-3.5 border-b border-tokyo-borderSubtle bg-tokyo-bgDark"
      >
        <div class="flex items-center gap-2.5 font-semibold text-sm">
          <div
            class="w-6 h-6 rounded bg-tokyo-teal/15 text-tokyo-teal flex items-center justify-center"
          >
            <GearIcon size={14} weight="bold" />
          </div>
          <span>Preferences & Settings</span>
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

      <!-- Tab Segmented Bar -->
      <div
        class="flex border-b border-tokyo-borderSubtle bg-tokyo-bgDark px-5 gap-4"
      >
        <button
          type="button"
          onclick={() => (activeTab = "downloads")}
          class="py-2 text-xs flex items-center gap-1.5 transition-colors border-b-2 {activeTab ===
          'downloads'
            ? 'border-tokyo-teal text-tokyo-teal font-medium'
            : 'border-transparent text-tokyo-textMuted hover:text-tokyo-textSub'}"
        >
          <SlidersIcon size={13} />
          <span>Downloads</span>
        </button>
        <button
          type="button"
          onclick={() => (activeTab = "about")}
          class="py-2 text-xs flex items-center gap-1.5 transition-colors border-b-2 {activeTab ===
          'about'
            ? 'border-tokyo-teal text-tokyo-teal font-medium'
            : 'border-transparent text-tokyo-textMuted hover:text-tokyo-textSub'}"
        >
          <InfoIcon size={13} />
          <span>About</span>
        </button>
      </div>

      <!-- Body -->
      <div class="p-5 max-h-[65vh] overflow-y-auto">
        {#if isLoading}
          <div
            class="py-12 flex flex-col items-center justify-center gap-2 text-tokyo-textMuted"
          >
            <CircleNotchIcon size={20} class="animate-spin text-tokyo-teal" />
            <span>Loading preferences...</span>
          </div>
        {:else if activeTab === "downloads"}
          <form onsubmit={handleSave} class="space-y-4">
            {#if errorMsg}
              <div
                class="p-3 rounded bg-tokyo-red/10 border border-tokyo-red/30 text-tokyo-red text-xs flex items-center gap-2"
              >
                <WarningCircleIcon size={16} weight="fill" class="shrink-0" />
                <span>{errorMsg}</span>
              </div>
            {/if}

            <!-- Download Directory -->
            <div class="space-y-1.5">
              <label
                for="download-dir"
                class="text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
              >
                <FolderIcon size={13} class="text-tokyo-blue" />
                <span>Default Download Directory</span>
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="download-dir"
                  type="text"
                  bind:value={downloadDir}
                  placeholder="/home/user/Downloads"
                  class="flex-1 h-9 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono focus:outline-none focus:border-tokyo-teal transition-colors"
                />
                <button
                  type="button"
                  onclick={handleBrowseFolder}
                  class="h-9 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border hover:bg-tokyo-bgHover text-tokyo-textSub hover:text-tokyo-textMain transition-colors flex items-center gap-1.5 font-medium shrink-0"
                >
                  <FolderIcon size={13} />
                  <span>Browse...</span>
                </button>
              </div>
              <p class="text-[11px] text-tokyo-textMuted">
                Default location for new downloads. Categories can organize
                files into subfolders.
              </p>
            </div>

            <!-- Concurrency & Parts Grid -->
            <div class="grid grid-cols-2 gap-3 pt-1">
              <!-- Max Concurrency -->
              <div class="space-y-1.5">
                <label
                  for="max-concurrency"
                  class="text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
                >
                  <SlidersIcon size={13} class="text-tokyo-teal" />
                  <span>Max Concurrent</span>
                </label>
                <input
                  id="max-concurrency"
                  type="number"
                  min="1"
                  max="10"
                  bind:value={maxConcurrency}
                  class="w-full h-9 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono focus:outline-none focus:border-tokyo-teal transition-colors"
                />
                <p class="text-[11px] text-tokyo-textMuted leading-tight">
                  Simultaneous downloads (1 – 10).
                </p>
              </div>

              <!-- Parts Per Download -->
              <div class="space-y-1.5">
                <label
                  for="parts-per-download"
                  class="text-xs font-medium text-tokyo-textSub flex items-center gap-1.5"
                >
                  <HardDriveIcon size={13} class="text-tokyo-purple" />
                  <span>Connections / Parts</span>
                </label>
                <div class="relative">
                  <select
                    id="parts-per-download"
                    bind:value={partsPerDownload}
                    class="w-full h-9 pl-3 pr-8 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono appearance-none cursor-pointer focus:outline-none focus:border-tokyo-teal transition-colors"
                  >
                    <option
                      value={1}
                      class="bg-tokyo-bgDark text-tokyo-textMain"
                      >1 (Single Part)</option
                    >
                    <option
                      value={2}
                      class="bg-tokyo-bgDark text-tokyo-textMain"
                      >2 Parts</option
                    >
                    <option
                      value={4}
                      class="bg-tokyo-bgDark text-tokyo-textMain"
                      >4 Parts (Default)</option
                    >
                    <option
                      value={8}
                      class="bg-tokyo-bgDark text-tokyo-textMain"
                      >8 Parts</option
                    >
                    <option
                      value={16}
                      class="bg-tokyo-bgDark text-tokyo-textMain"
                      >16 Parts</option
                    >
                  </select>
                  <div
                    class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-tokyo-textMuted"
                  >
                    <CaretDownIcon size={14} />
                  </div>
                </div>
                <p class="text-[11px] text-tokyo-textMuted leading-tight">
                  Parallel streams per download.
                </p>
              </div>
            </div>

            <!-- Min Multi-part Size -->
            <div class="space-y-1.5 pt-1">
              <label
                for="min-part-size"
                class="block text-xs font-medium text-tokyo-textSub"
              >
                Minimum Multi-Part Size (MB)
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="min-part-size"
                  type="number"
                  min="1"
                  max="1000"
                  bind:value={minMultiPartMB}
                  class="w-32 h-9 px-3 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono focus:outline-none focus:border-tokyo-teal transition-colors"
                />
                <span class="text-tokyo-textMuted font-mono">MB</span>
              </div>
              <p class="text-[11px] text-tokyo-textMuted">
                Files smaller than this threshold download as a single chunk
                without splitting.
              </p>
            </div>

            <!-- Speed Limit -->
            <div class="space-y-2 pt-2 border-t border-tokyo-borderSubtle">
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  bind:checked={isSpeedLimited}
                  class="rounded bg-tokyo-bgSurface border border-tokyo-border text-tokyo-teal focus:ring-0 focus:ring-offset-0 cursor-pointer accent-tokyo-teal"
                />
                <span class="text-xs font-medium text-tokyo-textSub">
                  Limit download speed
                </span>
              </label>

              {#if isSpeedLimited}
                <div
                  class="flex items-center gap-2 pl-5 pt-0.5 animate-in fade-in duration-150"
                >
                  <input
                    type="number"
                    min="1"
                    max="10000"
                    bind:value={speedLimitValue}
                    class="w-24 h-8 px-2.5 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono focus:outline-none focus:border-tokyo-teal transition-colors"
                  />
                  <select
                    bind:value={speedLimitUnit}
                    class="h-8 px-2 rounded bg-tokyo-bgSurface border border-tokyo-border text-xs text-tokyo-textMain font-mono cursor-pointer focus:outline-none focus:border-tokyo-teal"
                  >
                    <option value="MB" class="bg-tokyo-bgDark">MB/s</option>
                    <option value="KB" class="bg-tokyo-bgDark">KB/s</option>
                  </select>
                </div>
              {/if}
              <p class="text-[11px] text-tokyo-textMuted pl-5">
                Throttle maximum bandwidth across all active downloads.
              </p>
            </div>

            <!-- Interface / Appearance Options -->
            <div class="pt-3 border-t border-tokyo-borderSubtle space-y-2">
              <span class="text-xs font-medium text-tokyo-textSub block">
                Interface
              </span>
              <label class="flex items-start gap-2.5 cursor-pointer group">
                <input
                  type="checkbox"
                  bind:checked={showStatusBarLocal}
                  class="mt-0.5 rounded bg-tokyo-bgSurface border border-tokyo-border text-tokyo-teal focus:ring-0 focus:ring-offset-0 cursor-pointer accent-tokyo-teal"
                />
                <div class="flex flex-col">
                  <span
                    class="text-xs text-tokyo-textMain group-hover:text-tokyo-teal transition-colors"
                  >
                    Show status bar
                  </span>
                  <span class="text-[11px] text-tokyo-textMuted leading-tight">
                    Display session statistics, live bandwidth, disk storage,
                    and connection status at the bottom.
                  </span>
                </div>
              </label>
            </div>
          </form>
        {:else}
          <!-- About Tab -->
          <div class="space-y-4 py-1">
            <div
              class="p-3.5 rounded-lg bg-tokyo-bgSurface/50 border border-tokyo-borderSubtle space-y-1"
            >
              <div class="font-semibold text-tokyo-textMain text-sm">
                Downstream
              </div>
              <div class="text-[11px] text-tokyo-textMuted">
                A modern, high-performance desktop download manager built with
                Go, Wails v2, Svelte 5, and Tokyo Night aesthetic.
              </div>
            </div>

            <div class="space-y-2">
              <span class="text-xs font-semibold text-tokyo-textSub block">
                Keyboard Shortcuts
              </span>
              <div
                class="rounded border border-tokyo-borderSubtle divide-y divide-tokyo-borderSubtle text-[11px]"
              >
                <div class="flex justify-between px-3 py-1.5">
                  <span class="text-tokyo-textMuted">Close Modal / Unfocus</span
                  >
                  <kbd
                    class="font-mono text-tokyo-textSub bg-tokyo-bgSurface px-1.5 py-0.5 rounded border border-tokyo-border"
                    >Esc</kbd
                  >
                </div>
                <div class="flex justify-between px-3 py-1.5">
                  <span class="text-tokyo-textMuted">Open Downloaded File</span>
                  <span class="text-tokyo-textSub">Double Click Row</span>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer Bar -->
      <div
        class="px-5 py-3 border-t border-tokyo-borderSubtle bg-tokyo-bgDark flex items-center justify-between"
      >
        <div>
          {#if saveSuccess}
            <span
              class="text-tokyo-teal flex items-center gap-1 font-medium animate-in fade-in"
            >
              <CheckIcon size={14} weight="bold" />
              Settings saved!
            </span>
          {/if}
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            onclick={handleClose}
            disabled={isSaving}
            class="h-8 px-4 rounded border border-tokyo-border text-xs font-medium text-tokyo-textSub hover:bg-tokyo-bgSurface hover:text-tokyo-textMain transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          {#if activeTab === "downloads"}
            <button
              type="button"
              onclick={handleSave}
              disabled={isSaving || isLoading}
              class="h-8 px-4 rounded bg-tokyo-teal text-tokyo-bgDark text-xs font-semibold hover:bg-tokyo-teal/90 transition-colors flex items-center gap-1.5 disabled:opacity-50"
            >
              {#if isSaving}
                <CircleNotchIcon size={13} class="animate-spin" />
                <span>Saving...</span>
              {:else}
                <span>Save Changes</span>
              {/if}
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
