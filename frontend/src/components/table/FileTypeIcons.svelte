<script lang="ts">
  import {
    FilmSlateIcon,
    FileTextIcon,
    ArchiveIcon,
    DiscIcon,
    MusicNoteIcon,
    FileIcon,
  } from "phosphor-svelte";

  interface Props {
    category?: string;
    size?: number;
    wrapperClass?: string;
  }

  let { category = "", size = 16, wrapperClass = "" }: Props = $props();

  const iconConfig = $derived.by(() => {
    switch (category.toLowerCase()) {
      case "video":
        return {
          icon: FilmSlateIcon,
          color: "text-tokyo-orange bg-tokyo-orange/10",
        };
      case "documents":
      case "document":
        return {
          icon: FileTextIcon,
          color: "text-tokyo-blue bg-tokyo-blue/10",
        };
      case "compressed":
        return {
          icon: ArchiveIcon,
          color: "text-tokyo-purple bg-tokyo-purple/10",
        };
      case "programs":
        return { icon: DiscIcon, color: "text-tokyo-teal bg-tokyo-teal/10" };
      case "audio":
        return {
          icon: MusicNoteIcon,
          color: "text-tokyo-yellow bg-tokyo-yellow/10",
        };
      default:
        return { icon: FileIcon, color: "text-tokyo-cyan bg-tokyo-cyan/10" };
    }
  });
</script>

<div
  class={wrapperClass ? `${wrapperClass} ${iconConfig.color}` : `w-6 h-6 rounded flex items-center justify-center shrink-0 ${iconConfig.color}`}
>
  <iconConfig.icon {size} />
</div>
