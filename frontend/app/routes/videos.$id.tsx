import { useParams } from "react-router";
import { VideoPlayer } from "~/components/video/VideoPlayer";

export function meta() {
  return [{ title: "Watch Video" }];
}

export default function VideoPage() {
  const { id } = useParams();
  if (!id) return null;

  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-background p-4">
      <div className="w-full max-w-4xl">
        <VideoPlayer videoId={id} />
      </div>
    </main>
  );
}
