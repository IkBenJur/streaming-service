import { useParams, Link } from "react-router";
import { useVideoQuery } from "~/hooks/queries/useVideo";
import { VIDEO_STATUS } from "~/types/video.types";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Skeleton } from "~/components/ui/skeleton";

const STATUS_LABEL: Record<string, string> = {
  [VIDEO_STATUS.PENDING]: "Pending",
  [VIDEO_STATUS.PROCESSING]: "Processing",
  [VIDEO_STATUS.FINISHED]: "Finished",
  [VIDEO_STATUS.FAILED]: "Failed",
};

const STEPS = [
  { status: VIDEO_STATUS.PENDING, label: "Uploaded" },
  { status: VIDEO_STATUS.PROCESSING, label: "Processing" },
  { status: VIDEO_STATUS.FINISHED, label: "Finished" },
];

function stepState(stepStatus: string, currentStatus: string): "done" | "active" | "idle" {
  const order: string[] = [VIDEO_STATUS.PENDING, VIDEO_STATUS.PROCESSING, VIDEO_STATUS.FINISHED];
  const stepIdx = order.indexOf(stepStatus);
  const currentIdx = order.indexOf(currentStatus);
  if (currentStatus === VIDEO_STATUS.FAILED) return stepIdx === 0 ? "done" : "idle";
  if (currentIdx > stepIdx) return "done";
  if (currentIdx === stepIdx) return stepStatus === VIDEO_STATUS.FINISHED ? "done" : "active";
  return "idle";
}

export function meta() {
  return [{ title: "Manage Video" }];
}

export default function ManageVideoPage() {
  const { id } = useParams();
  const { data: video, isLoading } = useVideoQuery(id!);

  const isFailed = video?.status === VIDEO_STATUS.FAILED;
  const isFinished = video?.status === VIDEO_STATUS.FINISHED;
  const progress = video?.progress ?? 0;

  return (
    <div className="min-h-screen bg-[oklch(0.145_0.004_250)] text-foreground antialiased flex items-center justify-center px-4 py-8">
      <Card className="w-full max-w-2xl">
        <CardHeader className="px-5 pt-6 pb-4 sm:px-8 sm:pt-8 flex-row items-center justify-between">
          <CardTitle className="text-xl sm:text-2xl truncate pr-4">
            {isLoading ? <Skeleton className="h-7 w-48" /> : (video?.id ?? id)}
          </CardTitle>
          {video && (
            <Badge
              variant={isFailed ? "destructive" : isFinished ? "outline" : "default"}
              className={isFinished ? "border-green-500 text-green-400" : undefined}
            >
              {STATUS_LABEL[video.status] ?? video.status}
            </Badge>
          )}
        </CardHeader>

        <CardContent className="px-5 sm:px-8 pb-8 flex flex-col gap-8">
          {/* Steps */}
          <div className="flex items-center gap-0">
            {STEPS.map((step, i) => {
              const state = video ? stepState(step.status, video.status) : "idle";
              const isLast = i === STEPS.length - 1;
              return (
                <div key={step.status} className="flex items-center flex-1 last:flex-none">
                  <div className="flex flex-col items-center gap-2">
                    <div
                      className={[
                        "size-8 rounded-full border-2 flex items-center justify-center transition-colors",
                        state === "done" ? "border-green-500 bg-green-500/20" :
                        state === "active" ? "border-primary bg-primary/20" :
                        "border-muted-foreground/30 bg-transparent",
                      ].join(" ")}
                    >
                      {state === "done" ? (
                        <svg className="size-4 text-green-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      ) : state === "active" ? (
                        <div className="size-2.5 rounded-full bg-primary animate-pulse" />
                      ) : (
                        <div className="size-2.5 rounded-full bg-muted-foreground/30" />
                      )}
                    </div>
                    <span className={[
                      "text-xs font-medium",
                      state === "idle" ? "text-muted-foreground/50" : "text-foreground",
                    ].join(" ")}>
                      {step.label}
                    </span>
                  </div>
                  {!isLast && (
                    <div className={[
                      "flex-1 h-0.5 mb-5 mx-2 transition-colors",
                      state === "done" ? "bg-green-500/50" : "bg-muted-foreground/20",
                    ].join(" ")} />
                  )}
                </div>
              );
            })}
          </div>

          {/* Progress bar */}
          {video?.status === VIDEO_STATUS.PROCESSING && (
            <div className="flex flex-col gap-2">
              <div className="flex justify-between text-sm text-muted-foreground">
                <span>Transcoding</span>
                <span>{progress}%</span>
              </div>
              <div className="h-2 rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full rounded-full bg-primary transition-all duration-500"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}

          {isFailed && (
            <p className="text-sm text-destructive">
              Transcoding failed. Please try uploading the video again.
            </p>
          )}

          {isFinished && (
            <Button asChild size="lg" className="self-start">
              <Link to={`/videos/${id}`}>Watch video</Link>
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
