export interface SignedUrlResponse {
  signed_url: string;
}

export interface Video {
  id: string;
  status: string;
  progress: number | null;
  created_at: string | null;
  updated_at: string | null;
  file_extension: "mp4" | "webm";
}
