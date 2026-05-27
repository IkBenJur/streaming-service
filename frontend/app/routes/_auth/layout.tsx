import { redirect, Outlet } from "react-router";
import type { Route } from "./+types/layout";
import { getToken } from "~/lib/auth";

export async function clientLoader({ request }: Route.ClientLoaderArgs) {
  if (!getToken()) {
    const redirectTo = new URL(request.url).pathname;
    throw redirect(`/login?redirect=${redirectTo}`);
  }
  return null;
}

export default function AuthLayout() {
  return <Outlet />;
}
