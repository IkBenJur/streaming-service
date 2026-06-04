import { useMutation, useQueryClient } from "@tanstack/react-query";
import { videoService } from "~/services/video.service";

export const useCreateVideoMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ title, file }: { title: string; file: File }) => {
      const { id, "upload-url": uploadUrl } =
        await videoService.createVideoAndGetUploadUrl(title, file.name);
      await videoService.uploadFile(uploadUrl, file);
      await videoService.submitJob(id);
      return id;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["videos"] });
    },
  });
};
