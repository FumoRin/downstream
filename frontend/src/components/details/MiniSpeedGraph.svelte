<script lang="ts">
  import { onMount } from "svelte";
  import { PulseIcon } from "phosphor-svelte";
  import { formatSpeed } from "../../utils/formatters";
  import { speedTracker } from "../../utils/speedHistory.svelte";

  interface Props {
    speed?: number;
    status?: number;
    itemId?: string;
  }

  let { speed = 0, status = 0, itemId = "" }: Props = $props();

  let canvas: HTMLCanvasElement;
  let historyData = $derived(speedTracker.getHistory(itemId));

  // Determine colors & labels based on status
  const theme = $derived.by(() => {
    switch (status) {
      case 1: // Active
        return {
          stroke: "#4fd6be", // tokyo-teal
          gradientStart: "rgba(79, 214, 190, 0.25)",
          gradientEnd: "rgba(79, 214, 190, 0.0)",
          textClass: "text-tokyo-teal",
          label: formatSpeed(speed),
          badge: null,
        };
      case 2: // Paused
        return {
          stroke: "#ff9e64", // tokyo-orange
          gradientStart: "rgba(255, 158, 100, 0.20)",
          gradientEnd: "rgba(255, 158, 100, 0.0)",
          textClass: "text-tokyo-orange",
          label: "Paused",
          badge: "⏸ Paused",
          badgeClass:
            "text-tokyo-orange bg-tokyo-orange/10 border-tokyo-orange/20",
        };
      case 3: // Completed
        return {
          stroke: "#9ece6a", // tokyo-green
          gradientStart: "rgba(158, 206, 106, 0.20)",
          gradientEnd: "rgba(158, 206, 106, 0.0)",
          textClass: "text-tokyo-green",
          label: "Completed",
          badge: "✓ Finished",
          badgeClass:
            "text-tokyo-green bg-tokyo-green/10 border-tokyo-green/20",
        };
      case 4: // Failed
        return {
          stroke: "#f7768e", // tokyo-red
          gradientStart: "rgba(247, 118, 142, 0.20)",
          gradientEnd: "rgba(247, 118, 142, 0.0)",
          textClass: "text-tokyo-red",
          label: "Failed",
          badge: "✕ Failed",
          badgeClass: "text-tokyo-red bg-tokyo-red/10 border-tokyo-red/20",
        };
      default: // Queued / Scheduled
        return {
          stroke: "#bb9af7", // tokyo-purple
          gradientStart: "rgba(187, 154, 247, 0.20)",
          gradientEnd: "rgba(187, 154, 247, 0.0)",
          textClass: "text-tokyo-purple",
          label: "Queued",
          badge: "⏳ Queued",
          badgeClass:
            "text-tokyo-purple bg-tokyo-purple/10 border-tokyo-purple/20",
        };
    }
  });

  function drawChart() {
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const w = canvas.width;
    const h = canvas.height;
    ctx.clearRect(0, 0, w, h);

    const history = historyData.history;
    const max = Math.max(...history, 1024 * 1024); // at least 1MB scale
    const step = w / (history.length - 1);

    // Gradient fill using status theme
    const gradient = ctx.createLinearGradient(0, 0, 0, h);
    gradient.addColorStop(0, theme.gradientStart);
    gradient.addColorStop(1, theme.gradientEnd);

    ctx.beginPath();
    ctx.moveTo(0, h);

    for (let i = 0; i < history.length; i++) {
      const x = i * step;
      const y = h - (history[i] / max) * (h - 8);
      if (i === 0) ctx.lineTo(x, y);
      else ctx.lineTo(x, y);
    }

    ctx.lineTo(w, h);
    ctx.closePath();
    ctx.fillStyle = gradient;
    ctx.fill();

    // Solid stroke line using status theme
    ctx.beginPath();
    for (let i = 0; i < history.length; i++) {
      const x = i * step;
      const y = h - (history[i] / max) * (h - 8);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.strokeStyle = theme.stroke;
    ctx.lineWidth = 1.8;
    ctx.stroke();
  }

  $effect(() => {
    if (historyData || status) {
      drawChart();
    }
  });

  onMount(() => {
    drawChart();
  });
</script>

<div
  class="w-70 lg:w-[320px] h-42 border-l border-tokyo-borderSubtle pl-4 flex flex-col justify-between shrink-0 select-none"
>
  <!-- Header -->
  <div class="flex items-center justify-between text-xs">
    <div class="flex items-center gap-1.5 text-tokyo-textMuted">
      <PulseIcon size={14} class={theme.textClass} />
      <span class="text-[11px] font-medium uppercase tracking-wider"
        >Transfer Rate</span
      >
    </div>
    <span class="font-mono font-bold {theme.textClass}">
      {theme.label}
    </span>
  </div>

  <!-- Canvas Chart with Status Overlay -->
  <div
    class="relative h-[100px] w-full bg-tokyo-bgSurface/30 border border-tokyo-borderSubtle rounded overflow-hidden"
  >
    <canvas
      bind:this={canvas}
      width={300}
      height={100}
      class="w-full h-full block"
    ></canvas>

    {#if theme.badge}
      <div
        class="absolute inset-0 flex items-center justify-center bg-tokyo-bgDark/30 pointer-events-none"
      >
        <span
          class="px-2 py-0.5 rounded text-[10px] font-bold tracking-wider uppercase border shadow-sm backdrop-blur-[2px] {theme.badgeClass}"
        >
          {theme.badge}
        </span>
      </div>
    {/if}
  </div>

  <!-- Footer -->
  <div
    class="flex items-center justify-between text-[10px] font-mono text-tokyo-textMuted"
  >
    <span
      >Peak: <strong class="text-tokyo-textMain"
        >{formatSpeed(historyData.peak)}</strong
      ></span
    >
    <span>60s Rolling</span>
  </div>
</div>
