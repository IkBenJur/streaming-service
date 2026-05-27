import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Streaming Service" }];
}

export default function Home() {
  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <h1 className="text-2xl font-semibold text-foreground">Streaming Service</h1>
    </main>
  );
}
