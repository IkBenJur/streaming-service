import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";
import { videoService } from "~/services/video.service";

export const hlsSignedUrlQueryOptions = (id: string, file: string) =>
  queryOptions({
    queryKey: ["hlsSignedUrl", id, file],
    queryFn: () => videoService.getSignedUrl(id, file),
    staleTime: 50 * 60 * 1000,
  });

export const useHlsSignedUrlQuery = (id: string, file: string) =>
  useSuspenseQuery(hlsSignedUrlQueryOptions(id, file));
