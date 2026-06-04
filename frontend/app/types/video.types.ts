export const VIDEO_STATUS = {
  PENDING: "00000000-0000-0000-0000-000000000001",
  PROCESSING: "00000000-0000-0000-0000-000000000002",
  FINISHED: "00000000-0000-0000-0000-000000000003",
  FAILED: "00000000-0000-0000-0000-000000000004",
} as const;

export interface SignedUrlResponse {
  signed_url: string;
}

export interface CreateVideoResponse {
  id: string;
  "upload-url": string;
}

export interface Video {
  id: string;
  status: string;
  progress: number | null;
  created_at: string | null;
  updated_at: string | null;
  file_extension: "mp4" | "webm";
}
