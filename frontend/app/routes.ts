import { type RouteConfig, index, route, layout } from "@react-router/dev/routes";

export default [
  route(".well-known/appspecific/com.chrome.devtools.json", "routes/devtools-probe.ts"),
  index("routes/home.tsx"),
  route("videos", "routes/videos.tsx"),
  route("videos/manage/:id", "routes/videos.manage.$id.tsx"),
  route("watch/video/:id", "routes/watch-video.$id.tsx"),
  route("create-video", "routes/create-video.tsx"),
  layout("routes/_auth/layout.tsx", [
    // Protected routes go here
  ]),
] satisfies RouteConfig;
