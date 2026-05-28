import { Suspense } from "react";
import { Link } from "react-router";
import { useVideosQuery } from "~/hooks/queries/useVideos";
import { Skeleton } from "~/components/ui/skeleton";
import type { Video } from "~/types/video.types";

function IconPlay({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M8 5.5v13a1 1 0 0 0 1.5.87l11-6.5a1 1 0 0 0 0-1.74l-11-6.5A1 1 0 0 0 8 5.5z" />
    </svg>
  );
}

function VideoCard({ video, index, total }: { video: Video; index: number; total: number }) {
  return (
    <Link to={`/videos/${video.id}`} className="block no-underline group">
      <div className="relative w-full aspect-[16/10] rounded-[14px] overflow-hidden border border-border bg-[oklch(0.19_0.005_250)]">
        <span className="absolute top-2.5 left-2.5 sm:top-3.5 sm:left-3.5 font-mono text-[9px] sm:text-[10px] tracking-[0.16em] uppercase text-muted-foreground">
          {String(index + 1).padStart(2, "0")} / {String(total).padStart(2, "0")}
        </span>
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 size-8 sm:size-11 lg:size-[52px] rounded-full bg-foreground text-background grid place-items-center shadow-[0_8px_28px_oklch(0_0_0/0.45)] transition-transform duration-[160ms] scale-100 group-hover:scale-[1.08]">
          <IconPlay className="size-3.5 sm:size-4 lg:size-[18px]" />
        </div>
      </div>
      <div className="mt-3.5 flex items-baseline justify-between gap-4 px-0.5">
        <span
          className="font-mono text-base sm:text-[22px] leading-[1.1] tracking-tight font-medium text-foreground truncate"
          style={{ fontFeatureSettings: '"ss01", "tnum"' }}
        >
          {video.id}
        </span>
        <span className="font-mono text-[10px] sm:text-[11px] tracking-[0.04em] text-muted-foreground/50 shrink-0">
          title · {String(index + 1).padStart(2, "0")}
        </span>
      </div>
    </Link>
  );
}

function SkeletonCard() {
  return (
    <div>
      <Skeleton className="w-full aspect-[16/10] rounded-[14px]" />
      <div className="mt-3.5 px-0.5">
        <Skeleton className="h-[22px] w-3/5" />
      </div>
    </div>
  );
}

function VideoGrid() {
  const { data: videos } = useVideosQuery();
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-7 gap-y-14 pt-2">
      {videos.map((v, i) => (
        <VideoCard key={v.id} video={v} index={i} total={videos.length} />
      ))}
    </div>
  );
}

export function meta() {
  return [{ title: "Videos" }];
}

export default function VideosPage() {
  return (
    <div className="min-h-screen bg-[oklch(0.145_0.004_250)] text-foreground antialiased">
      <div className="max-w-[1400px] mx-auto px-9 pt-7 pb-36">
        <Suspense
          fallback={
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-7 gap-y-14 pt-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <SkeletonCard key={i} />
              ))}
            </div>
          }
        >
          <VideoGrid />
        </Suspense>
      </div>
    </div>
  );
}
