<script lang="ts">
  import AppHeader from "./components/header/AppHeader.svelte";
  import Sidebar from "./components/sidebar/Sidebar.svelte";
  import DownloadTable from "./components/table/DownloadTable.svelte";
  import DetailsPane from "./components/details/DetailsPane.svelte";
  import StatusBar from "./components/statusbar/StatusBar.svelte";
</script>

<div class="flex flex-col h-screen w-screen overflow-hidden bg-tokyo-bgDark text-tokyo-textMain select-none">
  <AppHeader />

  <div class="flex flex-1 min-h-0 overflow-hidden">
    <Sidebar />
    <main class="flex flex-1 flex-col min-w-0 bg-tokyo-bgBase overflow-hidden">
      <DownloadTable />

      <DetailsPane />
    </main>
  </div>

  <StatusBar />
</div>
