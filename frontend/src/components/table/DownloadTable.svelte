<script lang="ts">
  import {
    CaretUpIcon,
    CaretDownIcon,
    PlayIcon,
    PauseIcon,
    TrashIcon,
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
    onUnfocus,
  }: Props = $props();

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

<!-- Outer Container with background click to unfocus -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div
  class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden bg-tokyo-bgBase select-none"
  onclick={onUnfocus}
>
  <table class="w-full table-fixed border-collapse text-left text-xs font-sans">
    <thead
      class="sticky top-0 z-10 bg-tokyo-bgDark border-b border-tokyo-borderSubtle text-tokyo-textMuted uppercase text-[10px] tracking-wider"
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

        <th
          class="w-24 th-header hidden xl:table-cell"
          onclick={() => onSortChange?.("category")}
        >
          <div class="flex items-center gap-1">
            <span>Category</span>
            {#if sortField === "category"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th
          class="w-24 th-header"
          onclick={() => onSortChange?.("totalSize")}
        >
          <div class="flex items-center gap-1">
            <span>Size</span>
            {#if sortField === "totalSize"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th
          class="w-36 th-header"
          onclick={() => onSortChange?.("percentage")}
        >
          <div class="flex items-center gap-1">
            <span>Progress</span>
            {#if sortField === "percentage"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th
          class="w-24 th-header"
          onclick={() => onSortChange?.("speed")}
        >
          <div class="flex items-center gap-1">
            <span>Speed</span>
            {#if sortField === "speed"}
              {#if sortOrder === "asc"}<CaretUpIcon
                  size={12}
                />{:else}<CaretDownIcon size={12} />{/if}
            {/if}
          </div>
        </th>

        <th
          class="w-16 th-header hidden lg:table-cell"
          onclick={() => onSortChange?.("eta")}
        >
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
          class="w-24 th-header"
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

        <th
          class="w-28 th-header hidden xl:table-cell"
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

        <th class="w-20 px-3 py-2 text-center">Actions</th>
      </tr>
    </thead>

    <tbody class="divide-y divide-tokyo-borderSubtle">
      {#each items as item, idx (item.id)}
        {@const selected = selectedIds.has(item.id)}
        {@const meta = getStatusMeta(item.status)}
        <tr
          onclick={(e) => handleRowClick(e, item.id)}
          ondblclick={(e) => handleRowDblClick(e, item)}
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

          <!-- Category -->
          <td class="td-cell hidden xl:table-cell"
            >{item.category || "General"}</td
          >

          <!-- Size -->
          <td class="td-mono"
            >{formatBytes(item.totalSize)}</td
          >

          <!-- Progress Bar & Percentage -->
          <td class="px-3 py-2.5">
            <div class="w-28 space-y-1">
              <div
                class="h-1.5 w-full bg-tokyo-bgSurface rounded-full overflow-hidden"
              >
                <div
                  class="h-full rounded-full transition-all duration-200 {meta.barColor}"
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
            class={cn("td-mono", item.status === 1 ? "text-tokyo-teal font-medium" : "text-tokyo-textMuted")}
          >
            {item.status === 1 ? formatSpeed(item.speed) : "-"}
          </td>

          <!-- ETA -->
          <td class="td-mono-muted hidden lg:table-cell">
            {item.status === 3
              ? "Done"
              : item.status === 1
                ? formatETA(item.eta)
                : "--"}
          </td>

          <!-- Status Badge -->
          <td class="px-3 py-2.5">
            <span
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-medium {meta.badgeColor}"
            >
              <span class="w-1.5 h-1.5 rounded-full {meta.dotColor}"></span>
              <span>{meta.label}</span>
            </span>
          </td>

          <!-- Date Added -->
          <td
            class="td-mono-muted whitespace-nowrap hidden xl:table-cell"
          >
            {formatDate(item.createdAt)}
          </td>

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
      {:else}
        <tr>
          <td
            colspan="100"
            class="text-center py-16 text-tokyo-textMuted"
          >
            <div class="space-y-1">
              <p class="text-sm font-medium">No downloads found</p>
              <p class="text-xs">
                Add a URL to get started or clear active filters
              </p>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>


