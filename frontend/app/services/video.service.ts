import { apiClient } from "~/lib/apiClient";
import type { CreateVideoResponse, SignedUrlResponse, Video } from "~/types/video.types";

export const videoService = {
  listVideos(): Promise<Video[]> {
    return apiClient().get<Video[]>("/videos");
  },

  getSignedUrl(id: string, file: string): Promise<SignedUrlResponse> {
    return apiClient().get<SignedUrlResponse>(`/videos/${id}/stream/${file}/signed-url`);
  },

  createVideoAndGetUploadUrl(title: string, fileName: string): Promise<CreateVideoResponse> {
    return apiClient().post<CreateVideoResponse>("/videos/create-and-get-upload-url", {
      title,
      file_name: fileName,
    });
  },

  async uploadFile(uploadUrl: string, file: File): Promise<void> {
    const res = await fetch(uploadUrl, { method: "PUT", body: file });
    if (!res.ok) {
      throw new Error(`Upload failed: HTTP ${res.status}`);
    }
  },

  submitJob(id: string): Promise<void> {
    return apiClient().post<void>(`/videos/${id}/process`);
  },
};
