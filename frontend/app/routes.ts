import { type RouteConfig, index, route, layout } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  layout("routes/_auth/layout.tsx", [
    // Protected routes go here
  ]),
] satisfies RouteConfig;
