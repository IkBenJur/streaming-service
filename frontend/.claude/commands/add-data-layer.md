Add a new service, query hook, or mutation hook. The user will describe the feature or API operation as $ARGUMENTS (e.g. "fetch all videos", "create a video", "delete a video by id").

The data layer follows a strict three-tier separation:

```
Route component
  └── hooks/queries/<domain>.ts  or  hooks/mutations/use<Action>.ts
        └── services/<domain>.service.ts
              └── lib/apiClient.ts
```

Never call `apiClient()` directly from a component or route. Never put fetch logic in a hook file. Keep each layer to its single responsibility.

---

## When to add what

**Add a service function** when a new API endpoint needs to be called. One file per domain (e.g. `video.service.ts` covers all `/videos/...` endpoints). Add a new function to an existing service file if the domain already exists.

**Add a query hook** when a component needs to read data from the server (GET). Use `useSuspenseQuery` — not `useQuery` — so that loading state is handled by React Suspense boundaries, not inline conditionals.

**Add a mutation hook** when a component needs to write data (POST/PUT/PATCH/DELETE). One file per mutation in `hooks/mutations/`. Include cache invalidation and navigation in `onSuccess`.

---

## Step 1 — Identify the domain and determine what to create

From $ARGUMENTS, identify:
- The resource domain (e.g. "video", "user", "playlist")
- Whether this is a **read** (→ query) or **write** (→ mutation) operation
- The HTTP method and endpoint path

Read `app/lib/env.ts` to confirm the base URL env var name. Read `app/lib/apiClient.ts` to understand the available methods (`.get`, `.post`, `.put`, `.patch`, `.delete`).

Check whether a service file for this domain already exists in `app/services/`. If it does, add to it; do not create a duplicate.

---

## Step 2 — Create or update the service file

File: `app/services/<domain>.service.ts`

```typescript
import { apiClient } from "~/lib/apiClient";
import type { Video, CreateVideoDto } from "~/types/video.types";

export const videoService = {
  getAll: (): Promise<Video[]> =>
    apiClient().get("/videos"),

  getById: (id: string): Promise<Video> =>
    apiClient().get(`/videos/${id}`),

  create: (dto: CreateVideoDto): Promise<Video> =>
    apiClient().post("/videos", dto),

  delete: (id: string): Promise<void> =>
    apiClient().delete(`/videos/${id}`),
};
```

Rules:
- Export a single plain object (not a class) named `<domain>Service` (camelCase).
- Every method returns a `Promise<T>` — no hooks, no side effects.
- Import types from `~/types/<domain>.types.ts`. If the types file does not exist, create it (see Step 2b).

### Step 2b — Create or update the types file (if needed)

File: `app/types/<domain>.types.ts`

```typescript
export interface Video {
  id: string;
  title: string;
  createdAt: string;
}

export type CreateVideoDto = Pick<Video, "title">;
```

Rules:
- One file per domain.
- Use plain `interface` for API response shapes.
- Use `type` aliases (usually `Pick` or `Omit`) for request DTOs.
- Infer form types from the Zod schema in the component, not here — only put server-contract types in this file.

---

## Step 3a — Create the query hook (for read operations)

File: `app/hooks/queries/use<Domain>.ts`

```typescript
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";
import { videoService } from "~/services/video.service";
import type { Video } from "~/types/video.types";

export const allVideosQueryOptions = () =>
  queryOptions({
    queryKey: ["videos"],
    queryFn: () => videoService.getAll(),
  });

export const videoQueryOptions = (id: string) =>
  queryOptions({
    queryKey: ["videos", id],
    queryFn: () => videoService.getById(id),
  });

export const useGetAllVideosQuery = () =>
  useSuspenseQuery(allVideosQueryOptions());

export const useGetVideoQuery = (id: string) =>
  useSuspenseQuery(videoQueryOptions(id));
```

Rules:
- Always export `queryOptions` factories separately from the hooks. Loaders use the factory directly; components use the hook.
- Query key convention: `["<domain>"]` for collections, `["<domain>", id]` for single items.
- Use `useSuspenseQuery`, not `useQuery`.
- Group all queries for a domain into one file.

**Prefetch in the route loader** so data is ready before render:

```typescript
// In the route file (app/routes/_auth/videos.tsx):
import { queryClient } from "~/lib/queryClient";
import { allVideosQueryOptions } from "~/hooks/queries/useVideos";

export async function clientLoader() {
  await queryClient.ensureQueryData(allVideosQueryOptions());
  return null;
}
```

---

## Step 3b — Create the mutation hook (for write operations)

File: `app/hooks/mutations/use<Action><Domain>.ts`
Examples: `useCreateVideo.ts`, `useDeleteVideo.ts`, `useUpdateVideo.ts`

```typescript
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { videoService } from "~/services/video.service";
import type { CreateVideoDto } from "~/types/video.types";

export const useCreateVideoMutation = () => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (dto: CreateVideoDto) => videoService.create(dto),
    onSuccess: (newVideo) => {
      queryClient.invalidateQueries({ queryKey: ["videos"] });
      navigate(`/video/${newVideo.id}`);
    },
  });
};
```

Rules:
- One file per mutation.
- Always invalidate the relevant query cache in `onSuccess`. Prefer `invalidateQueries` over manual `setQueryData` unless you need optimistic updates.
- If the mutation should navigate after success, do it in `onSuccess` — not in the component.
- For delete mutations, invalidate the collection and remove the individual item from the cache:

```typescript
onSuccess: (_, id) => {
  queryClient.invalidateQueries({ queryKey: ["videos"] });
  queryClient.removeQueries({ queryKey: ["videos", id] });
  navigate("/videos");
},
```

**Optimistic update pattern** (use only when immediate UI feedback is required):

```typescript
onMutate: async (updated) => {
  const queryKey = ["videos", updated.id];
  await queryClient.cancelQueries({ queryKey });
  const previous = queryClient.getQueryData(queryKey);
  queryClient.setQueryData(queryKey, { ...previous, ...updated });
  return { previous };
},
onError: (_err, _vars, context) => {
  queryClient.setQueryData(["videos", _vars.id], context?.previous);
},
onSettled: (_data, _err, vars) => {
  queryClient.invalidateQueries({ queryKey: ["videos", vars.id] });
},
```

---

## Step 4 — Wire it into the route (if the user asks to)

If $ARGUMENTS mentions a route or page, also update the relevant route file:

1. Add the `clientLoader` that calls `queryClient.ensureQueryData(...)` (queries only).
2. In the page component, call the hook and render the data.
3. Update `app/routes.ts` if a new route file was created.

Do not add a `clientLoader` for mutation-only routes — loaders are for prefetching reads.

---

## Checklist before finishing

- [ ] Service file created or updated in `app/services/`
- [ ] Types file created or updated in `app/types/`
- [ ] Query hook in `app/hooks/queries/` (if read operation)
- [ ] Mutation hook in `app/hooks/mutations/` (if write operation)
- [ ] `queryOptions` factory exported separately from the hook (queries only)
- [ ] Cache invalidation in `onSuccess` (mutations only)
- [ ] No fetch calls outside `app/services/`
- [ ] No `useQuery` used — only `useSuspenseQuery`
