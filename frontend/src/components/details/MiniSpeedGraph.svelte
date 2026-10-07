<script lang="ts">
  import { onMount } from "svelte";
  import { PulseIcon } from "phosphor-svelte";
  import { formatSpeed } from "../../utils/formatters";

  interface Props {
    speed?: number;
    status?: number;
    itemId?: string;
  }

  let { speed = 0, status = 0, itemId = "" }: Props = $props();

  let canvas: HTMLCanvasElement;
  let history: number[] = new Array(60).fill(0);
  let peakSpeed = $state(0);
  let trackedItemId = "";

  // Reset history and peak when the selected download item changes
  $effect(() => {
    if (itemId !== trackedItemId) {
      trackedItemId = itemId;
      history = new Array(60).fill(0);
      peakSpeed = 0;
      drawChart();
    }
  });

  function drawChart() {
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const w = canvas.width;
    const h = canvas.height;
    ctx.clearRect(0, 0, w, h);

    const max = Math.max(...history, 1024 * 1024); // at least 1MB scale
    const step = w / (history.length - 1);

    // Gradient fill
    const gradient = ctx.createLinearGradient(0, 0, 0, h);
    gradient.addColorStop(0, "rgba(79, 214, 190, 0.25)");
    gradient.addColorStop(1, "rgba(79, 214, 190, 0.0)");

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

    // Solid stroke line
    ctx.beginPath();
    for (let i = 0; i < history.length; i++) {
      const x = i * step;
      const y = h - (history[i] / max) * (h - 8);
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.strokeStyle = "#4fd6be"; // tokyo-teal
    ctx.lineWidth = 1.8;
    ctx.stroke();
  }

  onMount(() => {
    drawChart();

    // Sample speed once every second for true 60s rolling window
    const timer = setInterval(() => {
      const current = status === 1 ? speed : 0;
      if (current > peakSpeed) {
        peakSpeed = current;
      }
      history.shift();
      history.push(current);
      drawChart();
    }, 1000);

    return () => {
      clearInterval(timer);
    };
  });
</script>

<div
  class="w-70 lg:w-[320px] h-33.75 border-l border-tokyo-borderSubtle pl-4 flex flex-col justify-between shrink-0 select-none"
>
  <!-- Header -->
  <div class="flex items-center justify-between text-xs">
    <div class="flex items-center gap-1.5 text-tokyo-textMuted">
      <PulseIcon size={14} class="text-tokyo-teal" />
      <span class="text-[11px] font-medium uppercase tracking-wider"
        >Transfer Rate</span
      >
    </div>
    <span class="font-mono font-bold text-tokyo-teal">
      {status === 1 ? formatSpeed(speed) : "-"}
    </span>
  </div>

  <!-- Canvas Chart -->
  <div
    class="h-[75px] w-full bg-tokyo-bgSurface/30 border border-tokyo-borderSubtle rounded overflow-hidden"
  >
    <canvas bind:this={canvas} width={300} height={75} class="w-full h-full block"
    ></canvas>
  </div>

  <!-- Footer -->
  <div
    class="flex items-center justify-between text-[10px] font-mono text-tokyo-textMuted"
  >
    <span
      >Peak: <strong class="text-tokyo-textMain"
        >{formatSpeed(peakSpeed)}</strong
      ></span
    >
    <span>60s Rolling</span>
  </div>
</div>
