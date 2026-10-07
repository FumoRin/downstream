<script lang="ts">
  import {
    FolderIcon,
    DownloadIcon,
    ClockIcon,
    PauseIcon,
    CheckIcon,
    WarningIcon,
    PlusIcon,
    GearIcon,
  } from "phosphor-svelte";
  import type { FilterState } from "../../types/download";
  interface Props {
    activeFilter?: FilterState;
    counts?: {
      all: number;
      active: number;
      scheduled: number;
      paused: number;
      completed: number;
      failed: number;
      categories: Record<string, number>;
    };
    onSelectFilter?: (filter: FilterState) => void;
    onOpenModal?: () => void;
  }

  let {
    activeFilter = { type: "status", value: "all" },
    counts = {
      all: 0,
      active: 0,
      scheduled: 0,
      paused: 0,
      completed: 0,
      failed: 0,
      categories: {},
    },
    onSelectFilter,
    onOpenModal,
  }: Props = $props();

  const statusItem = [
    { label: "All Downloads", value: "all", icon: FolderIcon, color: "" },
    {
      label: "Active",
      value: "1",
      icon: DownloadIcon,
      color: "text-tokyo-teal",
    },
    {
      label: "Paused",
      value: "2",
      icon: PauseIcon,
      color: "text-tokyo-orange",
    },
    {
      label: "Completed",
      value: "3",
      icon: CheckIcon,
      color: "text-tokyo-green",
    },
    { label: "Failed", value: "4", icon: WarningIcon, color: "text-tokyo-red" },
    {
      label: "Scheduled",
      value: "5",
      icon: ClockIcon,
      color: "text-tokyo-purple",
    },
  ];

  const categoryItems = [
    { label: "Video", dotColor: "bg-tokyo-orange" },
    { label: "Documents", dotColor: "bg-tokyo-blue" },
    { label: "Compressed", dotColor: "bg-tokyo-purple" },
    { label: "Programs", dotColor: "bg-tokyo-teal" },
    { label: "Audio", dotColor: "bg-tokyo-yellow" },
  ];

  function isActiveFilter(type: "status" | "category", value: string) {
    return (
      activeFilter.type === type &&
      activeFilter.value.toLowerCase() === value.toLowerCase()
    );
  }

  function getStatusCount(val: string) {
    switch (val) {
      case "all":
        return counts.all;
      case "1":
        return counts.active;
      case "2":
        return counts.paused;
      case "3":
        return counts.completed;
      case "4":
        return counts.failed;
      case "5":
        return counts.scheduled;
    }
  }
</script>

<aside
  class="w-55 bg-tokyo-bgDark border-r border-tokyo-borderSubtle flex flex-col shrink-0 h-full p-3 select-none"
>
  <button
    type="button"
    onclick={onOpenModal}
    title="Add a download"
    class="w-full h-9.5 bg-tokyo-teal text-tokyo-bgDark font-bold text-xs uppercase tracking-wide rounded flex items-center justify-center gap-2 shadow-sm hover:brightness-110 active:scale-[0.99] transition-all mb-4"
  >
    <PlusIcon size="16" stroke-width={2.5} />
    <span>Add Download</span>
  </button>

  <div class="flex-1 overflow-y-auto space-y-4 pr-1">
    <div>
      <div
        class="text-[10px] font-bold uppercase tracking-wide text-tokyo-textMuted px-2 mb-1.5"
      >
        Status
      </div>
      <div class="space-y-0.5">
        {#each statusItem as item}
          {@const active = isActiveFilter("status", item.value)}
          {@const IconComponent = item.icon}
          <button
            type="button"
            onclick={() =>
              onSelectFilter?.({ type: "status", value: item.value })}
            title={item.label}
            class="w-full h-8 px-2 rounded flex items-center justify-between text-xs transition-colors {active
              ? 'bg-tokyo-teal/10 text-tokyo-teal font-medium'
              : 'text-tokyo-textSub hover:bg-tokyo-bgSurface/50 hover:text-tokyo-textMain'}"
          >
            <div class="flex items-center gap-2">
              <IconComponent
                size="14"
                class={active ? "text-tokyo-teal" : item.color}
              />
              <span>{item.label}</span>
            </div>
            <span class="font-mono text-[11px] text-tokyo-textMuted"
              >{getStatusCount(item.value)}</span
            >
          </button>
        {/each}
      </div>
    </div>

    <div>
      <div
        class="text-[10px] font-bold uppercase tracking-wider text-tokyo-textMuted px-2 mb-1.5"
      >
        Categories
      </div>
      <div class="space-y-0.5">
        {#each categoryItems as cat}
          {@const active = isActiveFilter("category", cat.label)}
          <button
            type="button"
            onclick={() =>
              onSelectFilter?.({ type: "category", value: cat.label })}
            class="w-full h-8 px-2 rounded flex items-center justify-between text-xs transition-colors {active
              ? 'bg-tokyo-teal/10 text-tokyo-teal font-medium'
              : 'text-tokyo-textSub hover:bg-tokyo-bgSurface/50 hover:text-tokyo-textMain'}"
            title={cat.label}
          >
            <div class="flex items-center gap-2.5">
              <span class="w-2 h-2 rounded-full {cat.dotColor}"></span>
              <span>{cat.label}</span>
            </div>
            <span class="font-mono text-[11px] text-tokyo-textMuted">
              {counts.categories[cat.label] || 0}
            </span>
          </button>
        {/each}
      </div>
    </div>
  </div>

  <div class="pt-2 border-t border-tokyo-borderSubtle">
    <button
      type="button"
      class="w-full h-8 px-2 rounded flex items-center gap-2 text-xs text-tokyo-textMuted hover:text-tokyo-textMain hover:bg-tokyo-bgSurface/50 transition-colors"
      title="settings"
    >
      <GearIcon size={14} />
      <span>Settings</span>
    </button>
  </div>
</aside>
