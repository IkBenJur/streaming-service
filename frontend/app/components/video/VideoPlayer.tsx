import { useRef, useEffect } from "react";
import Hls, {
  type HlsConfig,
  type LoaderContext,
  type LoaderConfiguration,
  type LoaderCallbacks,
  type LoaderStats,
} from "hls.js";
import { videoService } from "~/services/video.service";
import { env } from "~/lib/env";

interface Props {
  videoId: string;
}

// Creates a custom hls.js loader that resolves every file (manifest + segments)
// through the backend's signed-URL endpoint before fetching from storage.
//
// We pass a URL shaped like /videos/:id/stream/index.m3u8 to hls.loadSource so
// that relative segment references in the manifest resolve to the same path
// prefix. The loader then extracts just the filename and calls the API.
function createSignedUrlLoader(videoId: string) {
  const BaseLoader = Hls.DefaultConfig.loader as unknown as new (
    config: HlsConfig
  ) => {
    context: LoaderContext;
    stats: LoaderStats;
    load(
      context: LoaderContext,
      config: LoaderConfiguration,
      callbacks: LoaderCallbacks<LoaderContext>
    ): void;
    abort(): void;
    destroy(): void;
    getStats(): LoaderStats;
  };

  return class SignedUrlLoader extends BaseLoader {
    private aborted = false;

    abort() {
      this.aborted = true;
      super.abort();
    }

    load(
      context: LoaderContext,
      loaderConfig: LoaderConfiguration,
      callbacks: LoaderCallbacks<LoaderContext>
    ) {
      const filename = new URL(context.url).pathname.split("/").pop()!;

      videoService
        .getSignedUrl(videoId, filename)
        .then(({ signed_url }) => {
          if (this.aborted) return;
          context.url = signed_url;
          super.load(context, loaderConfig, callbacks);
        })
        .catch((err: unknown) => {
          callbacks.onError(
            { code: 0, text: String(err) },
            context,
            null,
            {} as LoaderStats
          );
        });
    }
  };
}

export function VideoPlayer({ videoId }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    // The manifest URL is never fetched directly — the custom loader intercepts
    // it and replaces it with a signed URL. Its path shape just needs to match
    // the segment resolution base so hls.js resolves relative .ts references
    // to the same /videos/:id/stream/ prefix.
    const manifestUrl = `${env.VITE_API_URL}/videos/${videoId}/stream/playlist.m3u8`;

    if (Hls.isSupported()) {
      const hls = new Hls({ loader: createSignedUrlLoader(videoId) });
      hls.loadSource(manifestUrl);
      hls.attachMedia(video);
      return () => hls.destroy();
    }

    // Safari: native HLS support — segment requests go directly from the
    // browser to the S3 URL resolved from the manifest. Only the manifest
    // itself is signed; unsigned segment requests may fail if the bucket
    // requires per-object signing.
    if (video.canPlayType("application/vnd.apple.mpegurl")) {
      let cancelled = false;
      videoService.getSignedUrl(videoId, "playlist.m3u8").then(({ signed_url }) => {
        if (!cancelled) video.src = signed_url;
      });
      return () => {
        cancelled = true;
      };
    }
  }, [videoId]);

  return (
    <video
      ref={videoRef}
      controls
      className="w-full rounded-lg bg-black"
    />
  );
}
