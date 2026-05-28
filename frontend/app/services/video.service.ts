import { apiClient } from "~/lib/apiClient";
import type { SignedUrlResponse, Video } from "~/types/video.types";

export const videoService = {
  listVideos(): Promise<Video[]> {
    return apiClient().get<Video[]>("/videos");
  },

  getSignedUrl(id: string, file: string): Promise<SignedUrlResponse> {
    return apiClient().get<SignedUrlResponse>(`/videos/${id}/stream/${file}/signed-url`);
  },
};
