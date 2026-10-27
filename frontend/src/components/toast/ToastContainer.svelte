<script lang="ts">
  import { fly } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { toastManager, type ToastType } from "../../utils/toast.svelte";
  import {
    CheckCircleIcon,
    InfoIcon,
    WarningCircleIcon,
    XCircleIcon,
    XIcon,
  } from "phosphor-svelte";

  function getToastStyle(type: ToastType) {
    switch (type) {
      case "success":
        return {
          icon: CheckCircleIcon,
          border: "border-tokyo-teal/50",
          iconColor: "text-tokyo-teal",
          glow: "shadow-[0_4px_16px_rgba(79,214,190,0.12)]",
        };
      case "warning":
        return {
          icon: WarningCircleIcon,
          border: "border-tokyo-orange/50",
          iconColor: "text-tokyo-orange",
          glow: "shadow-[0_4px_16px_rgba(255,158,100,0.12)]",
        };
      case "error":
        return {
          icon: XCircleIcon,
          border: "border-tokyo-red/50",
          iconColor: "text-tokyo-red",
          glow: "shadow-[0_4px_16px_rgba(247,118,142,0.15)]",
        };
      default:
        return {
          icon: InfoIcon,
          border: "border-tokyo-cyan/50",
          iconColor: "text-tokyo-cyan",
          glow: "shadow-[0_4px_16px_rgba(125,207,255,0.12)]",
        };
    }
  }
</script>

<div
  class="fixed bottom-9 right-4 z-50 flex flex-col gap-2 max-w-sm pointer-events-none select-none"
  aria-live="polite"
>
  {#each toastManager.toasts as toast (toast.id)}
    {@const style = getToastStyle(toast.type)}
    {@const Icon = style.icon}
    <div
      in:fly={{ y: 16, duration: 220 }}
      out:fly={{ x: 40, duration: 180 }}
      animate:flip={{ duration: 200 }}
      class="pointer-events-auto flex items-center gap-3 px-3.5 py-2.5 rounded-lg bg-tokyo-bgDark border text-xs text-tokyo-textMain {style.border} {style.glow}"
    >
      <Icon size={16} class="shrink-0 {style.iconColor}" />
      <span class="flex-1 font-medium leading-tight">
        {toast.message}
      </span>
      <button
        type="button"
        onclick={() => toastManager.dismiss(toast.id)}
        class="text-tokyo-textMuted hover:text-tokyo-textMain p-0.5 rounded transition-colors shrink-0"
        title="Dismiss"
      >
        <XIcon size={14} />
      </button>
    </div>
  {/each}
</div>
