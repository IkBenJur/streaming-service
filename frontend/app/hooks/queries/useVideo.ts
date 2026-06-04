import { queryOptions, useQuery } from "@tanstack/react-query";
import { videoService } from "~/services/video.service";
import { VIDEO_STATUS } from "~/types/video.types";

export const videoQueryOptions = (id: string) =>
  queryOptions({
    queryKey: ["videos", id],
    queryFn: () => videoService.getVideo(id),
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === VIDEO_STATUS.FINISHED || status === VIDEO_STATUS.FAILED) return false;
      return 2000;
    },
  });

export const useVideoQuery = (id: string) => useQuery(videoQueryOptions(id));
