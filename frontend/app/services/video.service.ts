import { apiClient } from "~/lib/apiClient";
import type { SignedUrlResponse } from "~/types/video.types";

export const videoService = {
  getSignedUrl(id: string, file: string): Promise<SignedUrlResponse> {
    return apiClient().get<SignedUrlResponse>(`/videos/${id}/stream/${file}/signed-url`);
  },
};
