<script lang="ts">
  import { onMount } from "svelte";
  import {
    CaretUpIcon,
    CaretDownIcon,
    PlayIcon,
    PauseIcon,
    TrashIcon,
    ColumnsIcon,
    CheckIcon,
    ArrowSquareOutIcon,
    FolderIcon,
    LinkIcon,
    CopyIcon,
  } from "phosphor-svelte";
  import type {
    DownloadItem,
    SortField,
    SortOrder,
  } from "../../types/download";
  import FileTypeIcon from "./FileTypeIcons.svelte";
  import {
    formatBytes,
    formatSpeed,
    formatETA,
    formatDate,
  } from "../../utils/formatters";
  import { cn } from "../../utils/cn";
  import { showToast } from "../../utils/toast.svelte";

  interface Props {
    items: DownloadItem[];
    selectedIds: Set<string>;
    multiSelectMode?: boolean;
    sortField?: SortField;
    sortOrder?: SortOrder;
    onToggleSelect?: (id: string, isMulti: boolean, isShift: boolean) => void;
    onToggleSelectAll?: () => void;
    onSortChange?: (field: SortField) => void;
    onResume?: (id: string) => void;
    onPause?: (id: string) => void;
    onDelete?: (id: string) => void;
    onOpenFile?: (item: DownloadItem) => void;
    onShowInFolder?: (item: DownloadItem) => void;
    onUnfocus?: () => void;
  }

  let {
    items = [],
    selectedIds,
    multiSelectMode = false,
    sortField = "createdAt",
    sortOrder = "desc",
    onToggleSelect,
    onToggleSelectAll,
    onSortChange,
    onResume,
    onPause,
    onDelete,
    onOpenFile,
    onShowInFolder,
    onUnfocus,
  }: Props = $props();

  const STORAGE_KEY = "downstream_table_columns";

  // Optional column visibility
  let showDateAdded = $state(false);
  let showCategory = $state(false);
  let isColumnsMenuOpen = $state(false);

  // Right Click context menu
  let contextMenuOpen = $state(false);
  let contextMenuPos = $state({ x: 0, y: 0 });
  let contextMenuItem = $state<DownloadItem | null>(null);

  onMount(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        showDateAdded = !!parsed.dateAdded;
        showCategory = !!parsed.category;
      }
    } catch (e) {
      console.error("Failed to load table column preferences:", e);
    }
  });

  function toggleColumn(col: "dateAdded" | "category") {
    if (col === "dateAdded") showDateAdded = !showDateAdded;
    if (col === "category") showCategory = !showCategory;
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          dateAdded: showDateAdded,
          category: showCategory,
        }),
      );
    } catch (e) {
      console.error("Failed to save table column preferences:", e);
    }
  }

  function handleContextMenu(e: MouseEvent, item: DownloadItem) {
    e.preventDefault();
    e.stopPropagation();

    // If right-clicked item isn't in selection, select it
    if (!selectedIds.has(item.id)) {
      onToggleSelect?.(item.id, false, false);
    }

    // Clamp coordinates inside window boundaries
    const menuWidth = 190;
    const menuHeight = 220;
    const x = Math.min(e.clientX, window.innerWidth - menuWidth - 10);
    const y = Math.min(e.clientY, window.innerHeight - menuHeight - 10);

    contextMenuPos = { x, y };
    contextMenuItem = item;
    contextMenuOpen = true;
  }

  function closeContextMenu() {
    contextMenuOpen = false;
    contextMenuItem = null;
  }

  function handleWindowClick() {
    if (isColumnsMenuOpen) isColumnsMenuOpen = false;
    if (contextMenuOpen) closeContextMenu();
  }

  function handleWindowKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      if (isColumnsMenuOpen) isColumnsMenuOpen = false;
      if (contextMenuOpen) closeContextMenu();
      onUnfocus?.();
    }
  }

  const isAllSelected = $derived(
    items.length > 0 && items.every((d) => selectedIds.has(d.id)),
  );

  function getStatusMeta(status: number) {
    switch (status) {
      case 1:
        return {
          label: "Downloading",
          barColor: "bg-tokyo-teal",
          badgeColor: "text-tokyo-teal bg-tokyo-teal/10",
          dotColor: "bg-tokyo-teal",
        };
      case 2:
        return {
          label: "Paused",
          barColor: "bg-tokyo-orange",
          badgeColor: "text-tokyo-orange bg-tokyo-orange/10",
          dotColor: "bg-tokyo-orange",
        };
      case 3:
        return {
          label: "Done",
          barColor: "bg-tokyo-green",
          badgeColor: "text-tokyo-green bg-tokyo-green/10",
          dotColor: "bg-tokyo-green",
        };
      case 4:
        return {
          label: "Failed",
          barColor: "bg-tokyo-red",
          badgeColor: "text-tokyo-red bg-tokyo-red/10",
          dotColor: "bg-tokyo-red",
        };
      case 5:
        return {
          label: "Scheduled",
          barColor: "bg-tokyo-purple",
          badgeColor: "text-tokyo-purple bg-tokyo-purple/10",
          dotColor: "bg-tokyo-purple",
        };
      default:
        return {
          label: "Queued",
          barColor: "bg-tokyo-purple",
          badgeColor: "text-tokyo-purple bg-tokyo-purple/10",
          dotColor: "bg-tokyo-purple",
        };
    }
  }

  function handleRowClick(e: MouseEvent, id: string) {
    e.stopPropagation();
    onToggleSelect?.(id, e.ctrlKey || e.metaKey, e.shiftKey);
  }

  function handleRowDblClick(e: MouseEvent, item: DownloadItem) {
    e.stopPropagation();
    onOpenFile?.(item);
  }
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleWindowKeydown} />

<!-- Outer Container with background click to unfocus -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
  class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden bg-tokyo-bgBase select-none"
  onclick={onUnfocus}
>
  <table class="w-full table-fixed border-collapse text-left text-xs font-sans">
    <thead
      class="sticky top-0 z-10 bg-tokyo-bgDark border-b border-tokyo-borderSubtle text-tokyo-textMuted uppercase text-[10px] tracking-wider"
      oncontextmenu={(e) => {
        e.preventDefault();
        isColumnsMenuOpen = true;
      }}
    >
      <tr>
        {#if multiSelectMode}
          <th class="w-8 px-3 py-2 text-center">
            <input
              type="checkbox"
              checked={isAllSelected}
              onchange={onToggleSelectAll}
              class="rounded bg-tokyo-bgSurface border-tokyo-border text-tokyo-teal focus:ring-0 focus:outline-none cursor-pointer"
            />
          </th>
        {/if}
        <th class="w-8 px-2 py-2 text-center">#</th>

        <!-- Sortable Columns -->
        <th
          class="px-3 py-2 cursor-pointer hover:text-tokyo-textMain"
          onclick={() => onSortChange?.("filename")}
        >
          <div class="flex items-center gap-1">
            <span>Filename</span>
            {#if sortField === "filename"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        {#if showCategory}
          <th class="w-24 th-header" onclick={() => onSortChange?.("category")}>
            <div class="flex items-center gap-1">
              <span>Category</span>
              {#if sortField === "category"}
                {#if sortOrder === "asc"}<CaretUpIcon
                    size={12}
                  />{:else}<CaretDownIcon size={12} />{/if}
              {/if}
            </div>
          </th>
        {/if}

        <th class="w-20 th-header" onclick={() => onSortChange?.("totalSize")}>
          <div class="flex items-center gap-1">
            <span>Size</span>
            {#if sortField === "totalSize"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th class="w-32 th-header" onclick={() => onSortChange?.("percentage")}>
          <div class="flex items-center gap-1">
            <span>Progress</span>
            {#if sortField === "percentage"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th class="w-24 th-header" onclick={() => onSortChange?.("speed")}>
          <div class="flex items-center gap-1">
            <span>Speed</span>
            {#if sortField === "speed"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th class="w-16 th-header" onclick={() => onSortChange?.("eta")}>
          <div class="flex items-center gap-1">
            <span>ETA</span>
            {#if sortField === "eta"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th
          class="w-28 th-header pr-4"
          onclick={() => onSortChange?.("status")}
        >
          <div class="flex items-center gap-1">
            <span>Status</span>
            {#if sortField === "status"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        {#if showDateAdded}
          <th
            class="w-28 th-header"
            onclick={() => onSortChange?.("createdAt")}
          >
            <div class="flex items-center gap-1">
              <span>Date Added</span>
              {#if sortField === "createdAt"}
                {#if sortOrder === "asc"}<CaretUpIcon
                    size={12}
                  />{:else}<CaretDownIcon size={12} />{/if}
              {/if}
            </div>
          </th>
        {/if}

        <!-- Actions Column + Column Customizer Button -->
        <th class="w-22 px-2 py-2 text-center relative select-none">
          <div class="flex items-center justify-center gap-1">
            <span>Actions</span>
            <button
              type="button"
              onclick={(e) => {
                e.stopPropagation();
                isColumnsMenuOpen = !isColumnsMenuOpen;
              }}
              title="Customize columns"
              class="p-0.5 rounded text-tokyo-textMuted hover:text-tokyo-textMain hover:bg-tokyo-bgSurface transition-colors cursor-pointer"
            >
              <ColumnsIcon size={12} />
            </button>
          </div>

          <!-- Column Customizer Dropdown Menu -->
          {#if isColumnsMenuOpen}
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
            <div
              class="absolute right-0 top-full mt-1 w-44 bg-tokyo-bgDark border border-tokyo-border rounded-lg shadow-2xl p-1.5 text-xs text-tokyo-textMain z-50 normal-case tracking-normal text-left font-normal animate-in fade-in zoom-in-95 duration-100"
              onclick={(e) => e.stopPropagation()}
            >
              <div
                class="px-2 py-1 text-[10px] font-semibold text-tokyo-textMuted uppercase tracking-wider border-b border-tokyo-borderSubtle mb-1"
              >
                Table Columns
              </div>
              <button
                type="button"
                onclick={() => toggleColumn("dateAdded")}
                class="w-full px-2 py-1.5 rounded flex items-center justify-between text-left hover:bg-tokyo-bgSurface text-tokyo-textSub hover:text-tokyo-textMain transition-colors cursor-pointer"
              >
                <span>Date Added</span>
                {#if showDateAdded}
                  <CheckIcon size={13} weight="bold" class="text-tokyo-teal" />
                {/if}
              </button>
              <button
                type="button"
                onclick={() => toggleColumn("category")}
                class="w-full px-2 py-1.5 rounded flex items-center justify-between text-left hover:bg-tokyo-bgSurface text-tokyo-textSub hover:text-tokyo-textMain transition-colors cursor-pointer"
              >
                <span>Category</span>
                {#if showCategory}
                  <CheckIcon size={13} weight="bold" class="text-tokyo-teal" />
                {/if}
              </button>
            </div>
          {/if}
        </th>
      </tr>
    </thead>

    <tbody class="divide-y divide-tokyo-borderSubtle">
      {#each items as item, idx (item.id)}
        {@const selected = selectedIds.has(item.id)}
        {@const meta = getStatusMeta(item.status)}
        <tr
          onclick={(e) => handleRowClick(e, item.id)}
          ondblclick={(e) => handleRowDblClick(e, item)}
          oncontextmenu={(e) => handleContextMenu(e, item)}
          class="cursor-pointer transition-colors group {selected
            ? 'bg-tokyo-bgHighlight text-tokyo-textMain'
            : 'hover:bg-tokyo-bgSurface/40 text-tokyo-textSub'}"
        >
          {#if multiSelectMode}
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
            <td
              class="px-3 py-2.5 text-center cursor-pointer"
              onclick={(e) => {
                e.stopPropagation();
                onToggleSelect?.(item.id, true, false);
              }}
            >
              <input
                type="checkbox"
                checked={selected}
                class="rounded bg-tokyo-bgSurface border-tokyo-border text-tokyo-teal focus:ring-0 focus:outline-none pointer-events-none"
              />
            </td>
          {/if}

          <!-- Row Index -->
          <td
            class="px-2 py-2.5 text-center font-mono text-[11px] text-tokyo-textMuted"
            >{idx + 1}</td
          >

          <!-- Filename + FileTypeIcon + SaveDir -->
          <td class="px-3 py-2.5 min-w-0">
            <div class="flex items-center gap-2.5 min-w-0">
              <FileTypeIcon category={item.category} size={15} />
              <div class="min-w-0 flex-1 truncate">
                <div
                  class="font-semibold text-tokyo-textMain truncate leading-tight"
                >
                  {item.filename}
                </div>
                <div
                  class="text-[10px] font-mono text-tokyo-textMuted truncate"
                >
                  {item.saveDir || "~/Downloads"}
                </div>
              </div>
            </div>
          </td>

          {#if showCategory}
            <td class="td-cell">{item.category || "General"}</td>
          {/if}

          <!-- Size -->
          <td class="td-mono">{formatBytes(item.totalSize)}</td>

          <!-- Progress Bar & Percentage -->
          <td class="px-3 py-2.5">
            <div class="w-24 space-y-1">
              <div
                class="h-1.5 w-full bg-tokyo-bgSurface rounded-full overflow-hidden"
              >
                <div
                  class="h-full rounded-full transition-[width] duration-150 ease-out {meta.barColor}"
                  style="width: {item.status === 3
                    ? 100
                    : item.percentage || 0}%"
                ></div>
              </div>
              <div
                class="flex justify-between text-[10px] font-mono text-tokyo-textMuted leading-none"
              >
                <span
                  >{item.status === 3
                    ? 100
                    : item.percentage?.toFixed(1) || 0}%</span
                >
                <span>{formatBytes(item.currentSize || 0)}</span>
              </div>
            </div>
          </td>

          <!-- Speed -->
          <td
            class={cn(
              "td-mono whitespace-nowrap tabular-nums",
              item.status === 1
                ? "text-tokyo-teal font-medium"
                : "text-tokyo-textMuted",
            )}
          >
            {item.status === 1 ? formatSpeed(item.speed) : "-"}
          </td>

          <!-- ETA -->
          <td class="td-mono-muted">
            {item.status === 3
              ? "Done"
              : item.status === 1
                ? formatETA(item.eta)
                : "--"}
          </td>

          <!-- Status Badge -->
          <td class="px-3 py-2.5 pr-4">
            <span
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-medium {meta.badgeColor}"
            >
              <span class="w-1.5 h-1.5 rounded-full {meta.dotColor}"></span>
              <span>{meta.label}</span>
            </span>
          </td>

          {#if showDateAdded}
            <td class="td-mono-muted whitespace-nowrap">
              {formatDate(item.createdAt)}
            </td>
          {/if}

          <!-- Row Action Buttons -->
          <td class="px-3 py-2.5 text-center">
            <div
              class="flex items-center justify-center gap-1 opacity-60 group-hover:opacity-100 transition-opacity"
            >
              {#if item.status === 1}
                <button
                  type="button"
                  title="Pause"
                  onclick={(e) => {
                    e.stopPropagation();
                    onPause?.(item.id);
                  }}
                  class="btn-row-action hover:text-tokyo-orange"
                >
                  <PauseIcon size={13} fill="currentColor" />
                </button>
              {:else if item.status === 2 || item.status === 4}
                <button
                  type="button"
                  title="Resume"
                  onclick={(e) => {
                    e.stopPropagation();
                    onResume?.(item.id);
                  }}
                  class="btn-row-action hover:text-tokyo-teal"
                >
                  <PlayIcon size={13} fill="currentColor" />
                </button>
              {/if}
              <button
                type="button"
                title="Delete"
                onclick={(e) => {
                  e.stopPropagation();
                  onDelete?.(item.id);
                }}
                class="btn-row-action hover:bg-tokyo-red/10 hover:text-tokyo-red"
              >
                <TrashIcon size={13} />
              </button>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>

  {#if items.length === 0}
    <div
      class="py-20 flex flex-col items-center justify-center text-center text-tokyo-textMuted space-y-1"
    >
      <p class="text-sm font-medium">No downloads found</p>
      <p class="text-xs">Add a URL to get started or clear active filters</p>
    </div>
  {/if}

  {#if contextMenuOpen && contextMenuItem}
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div
      class="fixed z-50 min-w-48 bg-tokyo-bgDark border border-tokyo-border rounded-lg shadow-2xl py-1 text-xs text-tokyo-textMain select-none animate-in fade-in duration-75"
      style="left: {contextMenuPos.x}px; top: {contextMenuPos.y}px;"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Open File -->
      <button
        type="button"
        disabled={contextMenuItem.status !== 3}
        onclick={() => {
          if (contextMenuItem) onOpenFile?.(contextMenuItem);
          closeContextMenu();
        }}
        class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors disabled:opacity-40 disabled:pointer-events-none"
      >
        <ArrowSquareOutIcon size={14} class="text-tokyo-teal" />
        <span>Open File</span>
      </button>

      <!-- Show in Folder -->
      <button
        type="button"
        onclick={() => {
          if (contextMenuItem) onShowInFolder?.(contextMenuItem);
          closeContextMenu();
        }}
        class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors"
      >
        <FolderIcon size={14} class="text-tokyo-cyan" />
        <span>Show in Folder</span>
      </button>

      <div class="h-px bg-tokyo-borderSubtle my-1"></div>

      <!-- Copy Link -->
      <button
        type="button"
        onclick={() => {
          if (contextMenuItem)
            navigator.clipboard.writeText(contextMenuItem.url);
          showToast("Download link copied to clipboard", "info");
          closeContextMenu();
        }}
        class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors"
      >
        <LinkIcon size={14} class="text-tokyo-textMuted" />
        <span>Copy Download Link</span>
      </button>

      <!-- Copy Path -->
      <button
        type="button"
        onclick={() => {
          if (contextMenuItem) {
            const path = contextMenuItem.saveDir
              ? `${contextMenuItem.saveDir}/${contextMenuItem.filename}`
              : contextMenuItem.filename;
            navigator.clipboard.writeText(path);
            showToast("File path copied to clipboard", "info");
          }
          closeContextMenu();
        }}
        class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors"
      >
        <CopyIcon size={14} class="text-tokyo-textMuted" />
        <span>Copy File Path</span>
      </button>

      <div class="h-px bg-tokyo-borderSubtle my-1"></div>

      <!-- Resume / Pause -->
      {#if contextMenuItem.status === 1}
        <button
          type="button"
          onclick={() => {
            if (contextMenuItem) onPause?.(contextMenuItem.id);
            closeContextMenu();
          }}
          class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors text-tokyo-orange"
        >
          <PauseIcon size={14} />
          <span>Pause Download</span>
        </button>
      {:else if contextMenuItem.status === 2 || contextMenuItem.status === 0 || contextMenuItem.status === 4}
        <button
          type="button"
          onclick={() => {
            if (contextMenuItem) onResume?.(contextMenuItem.id);
            closeContextMenu();
          }}
          class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-bgSurface text-left transition-colors text-tokyo-teal"
        >
          <PlayIcon size={14} />
          <span>Resume Download</span>
        </button>
      {/if}

      <!-- Delete -->
      <button
        type="button"
        onclick={() => {
          if (contextMenuItem) onDelete?.(contextMenuItem.id);
          closeContextMenu();
        }}
        class="w-full px-3 py-1.5 flex items-center gap-2 hover:bg-tokyo-red/15 hover:text-tokyo-red text-left transition-colors text-tokyo-textSub"
      >
        <TrashIcon size={14} />
        <span>Delete...</span>
      </button>
    </div>
  {/if}
</div>
