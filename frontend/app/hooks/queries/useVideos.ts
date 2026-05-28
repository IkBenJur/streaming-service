import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";
import { videoService } from "~/services/video.service";

export const videosQueryOptions = () =>
  queryOptions({
    queryKey: ["videos"],
    queryFn: () => videoService.listVideos(),
  });

export const useVideosQuery = () => useSuspenseQuery(videosQueryOptions());
