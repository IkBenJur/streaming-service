import { type RouteConfig, index, route, layout } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("videos/:id", "routes/videos.$id.tsx"),
  layout("routes/_auth/layout.tsx", [
    // Protected routes go here
  ]),
] satisfies RouteConfig;
