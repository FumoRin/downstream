<script lang="ts">
  import { onMount } from "svelte";
  import { SvelteSet } from "svelte/reactivity";
  import AppHeader from "./components/header/AppHeader.svelte";
  import Sidebar from "./components/sidebar/Sidebar.svelte";
  import DownloadTable from "./components/table/DownloadTable.svelte";
  import DetailsPane from "./components/details/DetailsPane.svelte";
  import StatusBar from "./components/statusbar/StatusBar.svelte";
  import AddDownloadModal from "./components/header/AddDownloadModal.svelte";
  import DeleteConfirmModal from "./components/table/DeleteConfirmModal.svelte";

  import { splitPath } from "./utils/formatters";

  // Wails auto-generated Go bindings
  import {
    StartDownload,
    GetAllDownloads,
    StopDownload,
    ResumeDownload,
    DeleteDownload,
  } from "../wailsjs/go/main/App";
  import { EventsOn } from "../wailsjs/runtime/runtime";

  import type {
    DownloadItem,
    FilterState,
    SortField,
    SortOrder,
  } from "./types/download";

  let downloads = $state<DownloadItem[]>([]);
  let selectedIds = new SvelteSet<string>();
  let multiSelectMode = $state(false);
  let lastClickedId = $state<string | null>(null);
  let searchQuery = $state("");
  let activeFilter = $state<FilterState>({ type: "status", value: "all" });
  let sortField = $state<SortField>("createdAt");
  let sortOrder = $state<SortOrder>("desc");
  let isAddModalOpen = $state(false);

  // Delete Confirmation Modal State
  let isDeleteModalOpen = $state(false);
  let pendingDeleteIds = $state<string[]>([]);
  let pendingDeleteItems = $derived(
    downloads.filter((d) => pendingDeleteIds.includes(d.id)),
  );

  // Live counters for Sidebar
  let counts = $derived.by(() => {
    const res = {
      all: downloads.length,
      active: 0,
      scheduled: 0,
      paused: 0,
      completed: 0,
      failed: 0,
      categories: {} as Record<string, number>,
    };
    for (const d of downloads) {
      if (d.status === 1) res.active++;
      else if (d.status === 5) res.scheduled++;
      else if (d.status === 2) res.paused++;
      else if (d.status === 3) res.completed++;
      else if (d.status === 4) res.failed++;

      if (d.category) {
        res.categories[d.category] = (res.categories[d.category] || 0) + 1;
      }
    }
    return res;
  });

  // Selected single item for DetailsPane
  let selectedItem = $derived(
    selectedIds.size === 1
      ? downloads.find((d) => selectedIds.has(d.id)) || null
      : null,
  );

  // Filtered & Sorted downloads list for DownloadTable
  let filteredDownloads = $derived(
    downloads
      .filter((item) => {
        if (searchQuery.trim()) {
          const q = searchQuery.toLowerCase();
          if (
            !item.filename.toLowerCase().includes(q) &&
            !item.url.toLowerCase().includes(q)
          ) {
            return false;
          }
        }
        if (activeFilter.value === "all") return true;
        if (activeFilter.type === "status") {
          return item.status.toString() === activeFilter.value;
        } else {
          return (
            item.category?.toLowerCase() === activeFilter.value.toLowerCase()
          );
        }
      })
      .sort((a, b) => {
        const valA = a[sortField];
        const valB = b[sortField];
        if (sortField === "createdAt") {
          return sortOrder === "asc"
            ? new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
            : new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        }
        if (typeof valA === "string") {
          return sortOrder === "asc"
            ? (valA as string).localeCompare(valB as string)
            : (valB as string).localeCompare(valA as string);
        }
        return sortOrder === "asc"
          ? (valA as number) - (valB as number)
          : (valB as number) - (valA as number);
      }),
  );

  // Load downloads from Go backend
  async function loadDownloads() {
    try {
      const raw = await GetAllDownloads();
      if (raw && Array.isArray(raw)) {
        downloads = raw.map((d: any) => {
          const id = d.id || d.ID || "";
          const existing = downloads.find((item) => item.id === id);
          const rawPath = d.filename || d.Filename || existing?.filename || "download"
          const {filename: baseName, dir: parsedDir} = splitPath(rawPath)
          const rawStatus = (d.status !== undefined ? d.status : d.Status) ?? 0;
          const status =
            existing?.status === 2 && rawStatus === 1 ? 2 : rawStatus;
          const totalSize =
            (d.totalSize ?? d.TotalSize) || existing?.totalSize || 0;
          const currentSize =
            (d.currentSize ?? d.CurrentSize) || existing?.currentSize || 0;
          let percentage =
            (d.percentage ?? d.Percentage) ?? existing?.percentage ?? 0;

          if (status === 3) {
            percentage = 100;
          } else if (percentage === 0 && totalSize > 0 && currentSize > 0) {
            percentage = (currentSize / totalSize) * 100;
          }

          return {
            id,
            filename: baseName,
            category:
              d.category || d.Category || existing?.category || "General",
            url: d.url || d.URL || existing?.url || "",
            totalSize,
            currentSize,
            percentage,
            speed:
              status === 1
                ? (d.speed ?? d.Speed ?? existing?.speed ?? 0)
                : 0,
            eta:
              status === 1
                ? (d.eta ?? d.ETA ?? existing?.eta ?? 0)
                : 0,
            status,
            priority: "normal",
            createdAt:
              d.createdAt ||
              d.CreatedAt ||
              existing?.createdAt ||
              new Date().toISOString(),
            saveDir: d.saveDir || d.SaveDir || parsedDir || existing?.saveDir || "~/Downloads",
            parts: d.parts || existing?.parts || [],
          };
        });
      }
    } catch (e) {
      console.warn("Wails backend not connected or error fetching downloads:", e);
    }
  }

  // Download actions
  async function handleAddDownload(url: string, filename: string, dir: string) {
    try {
      await StartDownload(url, filename, dir);
      await loadDownloads();
    } catch (e) {
      console.error("StartDownload error:", e);
      throw e;
    }
  }

  async function handleResume(id: string) {
    try {
      const idx = downloads.findIndex((d) => d.id === id);
      if (idx !== -1) {
        downloads[idx] = {
          ...downloads[idx],
          status: 1, // Downloading immediately
        };
      }
      await ResumeDownload(id);
      await loadDownloads();
    } catch (e) {
      console.error("ResumeDownload error:", e);
    }
  }

  async function handlePause(id: string) {
    try {
      const idx = downloads.findIndex((d) => d.id === id);
      if (idx !== -1) {
        downloads[idx] = {
          ...downloads[idx],
          status: 2, // Paused immediately
          speed: 0,
          eta: 0,
        };
      }
      await StopDownload(id);
      await loadDownloads();
    } catch (e) {
      console.error("StopDownload error:", e);
    }
  }

  function requestDelete(ids: string[]) {
    if (ids.length === 0) return;
    pendingDeleteIds = ids;
    isDeleteModalOpen = true;
  }

  async function handleConfirmDelete(deleteFiles: boolean) {
    const toDelete = [...pendingDeleteIds];
    pendingDeleteIds = [];
    isDeleteModalOpen = false;

    // Immediately remove from local state so UI updates without waiting
    downloads = downloads.filter((d) => !toDelete.includes(d.id));
    for (const id of toDelete) {
      selectedIds.delete(id);
    }

    for (const id of toDelete) {
      try {
        await DeleteDownload(id, deleteFiles);
      } catch (e) {
        console.error("DeleteDownload error:", e);
      }
    }
    await loadDownloads();
  }

  async function handleBatchResume() {
    for (const id of selectedIds) await handleResume(id);
  }

  async function handleBatchPause() {
    for (const id of selectedIds) await handlePause(id);
  }

  function handleBatchDelete() {
    requestDelete(Array.from(selectedIds));
  }

  function handleSortChange(field: SortField) {
    if (sortField === field) {
      sortOrder = sortOrder === "asc" ? "desc" : "asc";
    } else {
      sortField = field;
      sortOrder = [
        "totalSize",
        "percentage",
        "speed",
        "createdAt",
        "eta",
      ].includes(field)
        ? "desc"
        : "asc";
    }
  }

  function handleToggleSelect(id: string, isMulti: boolean, isShift: boolean) {
    if (isShift && lastClickedId && lastClickedId !== id) {
      const fromIdx = filteredDownloads.findIndex((d) => d.id === lastClickedId);
      const toIdx = filteredDownloads.findIndex((d) => d.id === id);
      if (fromIdx !== -1 && toIdx !== -1) {
        const [start, end] =
          fromIdx < toIdx ? [fromIdx, toIdx] : [toIdx, fromIdx];
        for (let i = start; i <= end; i++) {
          selectedIds.add(filteredDownloads[i].id);
        }
        lastClickedId = id;
        return;
      }
    }

    if (multiSelectMode || isMulti) {
      if (selectedIds.has(id)) selectedIds.delete(id);
      else selectedIds.add(id);
      lastClickedId = id;
    } else {
      if (selectedIds.size === 1 && selectedIds.has(id)) {
        selectedIds.clear();
        lastClickedId = null;
      } else {
        selectedIds.clear();
        selectedIds.add(id);
        lastClickedId = id;
      }
    }
  }

  function handleToggleSelectAll() {
    if (filteredDownloads.every((d) => selectedIds.has(d.id))) {
      selectedIds.clear();
    } else {
      for (const d of filteredDownloads) selectedIds.add(d.id);
    }
  }

  function unfocus() {
    selectedIds.clear();
    lastClickedId = null;
  }

  onMount(async () => {
    await loadDownloads();

    // Stream real-time progress updates from Go backend
    EventsOn("download:progress", (prog: any) => {
      const progId = prog.id || prog.ID || "";
      const progFilename = prog.filename || prog.Filename || "";
      const idx = downloads.findIndex(
        (d) =>
          (progId && d.id === progId) ||
          (progFilename && (
            d.filename === progFilename ||
            progFilename.endsWith(d.filename) ||
            d.filename.endsWith(progFilename)
          )),
      );
      if (idx !== -1) {
        // If download was paused by the user, ignore late in-flight progress packets
        if (downloads[idx].status === 2) {
          return;
        }

        const percentage = prog.percentage ?? prog.Percentage ?? 0;
        const currentSize = prog.currentSize ?? prog.CurrentSize ?? 0;
        const totalSize =
          (prog.totalSize ?? prog.TotalSize) || downloads[idx].totalSize;
        const speed = prog.speed ?? prog.Speed ?? 0;
        const rawEta = prog.eta ?? prog.ETA ?? 0;
        const eta =
          typeof rawEta === "number"
            ? (rawEta > 1e6 ? Math.floor(rawEta / 1e9) : rawEta)
            : 0;

        downloads[idx] = {
          ...downloads[idx],
          percentage,
          currentSize,
          totalSize,
          speed,
          eta,
          status: percentage >= 100 ? 3 : 1,
        };
      }
    });
  });
</script>

<div
  class="flex flex-col h-screen w-screen overflow-hidden bg-tokyo-bgDark text-tokyo-textMain select-none"
>
  <AppHeader
    bind:multiSelectMode
    bind:searchQuery
    selectedCount={selectedIds.size}
    onClearSelection={unfocus}
    onToggleMultiSelect={() => (multiSelectMode = !multiSelectMode)}
    onResume={handleBatchResume}
    onPause={handleBatchPause}
    onDelete={handleBatchDelete}
  />

  <div class="flex flex-1 min-h-0 overflow-hidden">
    <Sidebar
      {activeFilter}
      {counts}
      onSelectFilter={(f) => (activeFilter = f)}
      onOpenModal={() => (isAddModalOpen = true)}
    />
    <main class="flex flex-1 flex-col min-w-0 bg-tokyo-bgBase overflow-hidden">
      <DownloadTable
        items={filteredDownloads}
        {selectedIds}
        {multiSelectMode}
        {sortField}
        {sortOrder}
        onToggleSelect={handleToggleSelect}
        onToggleSelectAll={handleToggleSelectAll}
        onSortChange={handleSortChange}
        onUnfocus={unfocus}
        onResume={handleResume}
        onPause={handlePause}
        onDelete={(id) => requestDelete([id])}
        onOpenFile={(item) => console.log("Open file:", item.filename)}
      />
      <DetailsPane
        {selectedItem}
        selectedCount={selectedIds.size}
      />
    </main>
  </div>

  <StatusBar />

  <!-- Add Download Modal Dialog -->
  <AddDownloadModal
    bind:open={isAddModalOpen}
    onSubmit={handleAddDownload}
  />

  <!-- Delete Confirmation Modal Dialog -->
  <DeleteConfirmModal
    bind:open={isDeleteModalOpen}
    items={pendingDeleteItems}
    onConfirm={handleConfirmDelete}
  />
</div>
