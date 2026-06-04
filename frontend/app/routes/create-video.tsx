import { useRef } from "react";
import { useForm } from "react-hook-form";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "~/components/ui/card";

function IconUpload({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="17 8 12 3 7 8" />
      <line x1="12" y1="3" x2="12" y2="15" />
    </svg>
  );
}

interface CreateVideoForm {
  title: string;
  file: FileList;
}

export function meta() {
  return [{ title: "Create Video" }];
}

export default function CreateVideoPage() {
  const { register, handleSubmit, watch, formState: { errors } } = useForm<CreateVideoForm>();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const fileRegistration = register("file", { required: "File is required", validate: (files) => files?.length > 0 || "File is required" });

  const selectedFile = watch("file")?.[0];

  function onSubmit(data: CreateVideoForm) {
    console.log(data);
  }

  return (
    <div className="min-h-screen bg-[oklch(0.145_0.004_250)] text-foreground antialiased flex items-center justify-center px-4 py-8">
      <Card className="w-full max-w-2xl">
        <CardHeader className="px-5 pt-6 pb-4 sm:px-8 sm:pt-8">
          <CardTitle className="text-xl sm:text-2xl">Create Video</CardTitle>
        </CardHeader>
        <CardContent className="px-5 sm:px-8">
          <form id="create-video-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-6">
            <div className="flex flex-col gap-1.5 sm:flex-row sm:items-start sm:gap-5">
              <Label htmlFor="title" className="text-base sm:w-14 sm:shrink-0 sm:mt-2.5">Title</Label>
              <div className="flex flex-col gap-2 flex-1">
                <Input
                  id="title"
                  className="h-11 text-base"
                  aria-invalid={!!errors.title}
                  {...register("title", { required: "Title is required" })}
                />
                {errors.title && (
                  <span className="text-sm text-destructive">{errors.title.message}</span>
                )}
              </div>
            </div>
            <div className="flex flex-col gap-1.5 sm:flex-row sm:items-start sm:gap-5">
              <Label htmlFor="file" className="text-base sm:w-14 sm:shrink-0 sm:mt-2.5">File</Label>
              <div className="flex flex-col gap-2">
                <input
                  type="file"
                  accept=".mp4,.webm"
                  className="hidden"
                  {...fileRegistration}
                  ref={(e) => {
                    fileRegistration.ref(e);
                    fileInputRef.current = e;
                  }}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="lg"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <IconUpload className="size-4" />
                  Browse files
                </Button>
                {selectedFile ? (
                  <span className="text-sm text-muted-foreground truncate max-w-xs">
                    {selectedFile.name}
                  </span>
                ) : errors.file && (
                  <span className="text-sm text-destructive">{errors.file.message}</span>
                )}
              </div>
            </div>
          </form>
        </CardContent>
        <CardFooter className="px-5 pb-6 sm:px-8 sm:pb-8">
          <Button type="submit" form="create-video-form" size="lg" className="ml-auto">Upload</Button>
        </CardFooter>
      </Card>
    </div>
  );
}
