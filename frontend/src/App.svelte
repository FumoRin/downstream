<script lang="ts">
  import { onMount } from "svelte";
  // Wails auto-generated bindings
  import {
    StartDownload,
    GetAllDownloads,
    StopDownload,
    ResumeDownload,
    DeleteDownload,
  } from "../wailsjs/go/main/App";
  import { EventsOn } from "../wailsjs/runtime/runtime";

  interface DownloadItem {
    ID: string;
    URL: string;
    Filename: string;
    Category: string;
    TotalSize: number;
    Status: number;
    // Realtime progress fields
    percentage?: number;
    speed?: number;
    eta?: number;
  }

  let downloads = $state<DownloadItem[]>([]);
  let newUrl = $state("");
  let customDir = $state("");

  const statusNames = [
    "Queued",
    "Downloading",
    "Paused",
    "Completed",
    "Error",
    "Scheduled",
  ];

  function formatBytes(bytes: number): string {
    if (!bytes || bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  async function loadDownloads() {
    downloads = (await GetAllDownloads()) || [];
  }

  async function handleStart() {
    if (!newUrl.trim()) return;
    await StartDownload(newUrl, "", customDir);
    newUrl = "";
    await loadDownloads();
  }

  async function handleStop(id: string) {
    await StopDownload(id);
    await loadDownloads();
  }

  async function handleResume(id: string) {
    await ResumeDownload(id);
    await loadDownloads();
  }

  async function handleDelete(id: string) {
    await DeleteDownload(id, false);
    await loadDownloads();
  }

  onMount(async () => {
    await loadDownloads();

    // Listen to realtime progress updates streamed from Go ProgressChan
    EventsOn("download:progress", (prog: any) => {
      const idx = downloads.findIndex(
        (d) =>
          d.Filename === prog.Filename || d.Filename.endsWith(prog.Filename),
      );
      if (idx !== -1) {
        downloads[idx].percentage = prog.Percentage;
        downloads[idx].speed = prog.Speed;
      }
    });
  });
</script>

<main class="container">
  <header>
    <h1>GDL Download Manager</h1>
  </header>

  <!-- Add Download Section -->
  <section class="add-section">
    <input
      type="text"
      bind:value={newUrl}
      placeholder="Enter download URL..."
    />
    <input
      type="text"
      bind:value={customDir}
      placeholder="Optional custom directory (-d)..."
    />
    <button onclick={handleStart}>Download</button>
  </section>

  <!-- Downloads List Table -->
  <section class="list-section">
    <table>
      <thead>
        <tr>
          <th>Filename</th>
          <th>Category</th>
          <th>Size</th>
          <th>Status</th>
          <th>Progress</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each downloads as item (item.ID)}
          <tr>
            <td class="filename">{item.Filename}</td>
            <td><span class="badge">{item.Category || "General"}</span></td>
            <td>{formatBytes(item.TotalSize)}</td>
            <td>{statusNames[item.Status] || "Unknown"}</td>
            <td>
              <div class="progress-bar">
                <div
                  class="fill"
                  style="width: {item.percentage ||
                    (item.Status === 3 ? 100 : 0)}%"
                ></div>
              </div>
              {#if item.speed && item.Status === 1}
                <small>{formatBytes(item.speed)}/s</small>
              {/if}
            </td>
            <td class="actions">
              {#if item.Status === 1}
                <button onclick={() => handleStop(item.ID)}>Pause</button>
              {:else if item.Status === 2 || item.Status === 4}
                <button onclick={() => handleResume(item.ID)}>Resume</button>
              {/if}
              <button class="btn-del" onclick={() => handleDelete(item.ID)}
                >Delete</button
              >
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="empty">No downloads yet. Add a URL above.</td
            >
          </tr>
        {/each}
      </tbody>
    </table>
  </section>
</main>

<style>
  :global(body) {
    margin: 0;
    font-family:
      system-ui,
      -apple-system,
      BlinkMacSystemFont,
      "Segoe UI",
      Roboto,
      sans-serif;
    background: #1e1e2e;
    color: #cdd6f4;
  }
  .container {
    padding: 24px;
    max-width: 1100px;
    margin: 0 auto;
  }
  header h1 {
    margin-bottom: 24px;
    font-size: 24px;
  }
  .add-section {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
  }
  input {
    flex: 1;
    padding: 10px 14px;
    background: #313244;
    border: 1px solid #45475a;
    border-radius: 6px;
    color: white;
  }
  button {
    padding: 10px 18px;
    background: #89b4fa;
    border: none;
    border-radius: 6px;
    color: #11111b;
    font-weight: bold;
    cursor: pointer;
  }
  button:hover {
    background: #b4befe;
  }
  .btn-del {
    background: #f38ba8;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    background: #181825;
    border-radius: 8px;
    overflow: hidden;
  }
  th,
  td {
    padding: 12px 16px;
    text-align: left;
    border-bottom: 1px solid #313244;
  }
  th {
    background: #11111b;
  }
  .filename {
    max-width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .badge {
    background: #45475a;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
  }
  .progress-bar {
    width: 140px;
    height: 8px;
    background: #313244;
    border-radius: 4px;
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: #a6e3a1;
    transition: width 0.2s;
  }
  .actions {
    display: flex;
    gap: 8px;
  }
  .empty {
    text-align: center;
    padding: 32px;
    color: #6c7086;
  }
</style>
