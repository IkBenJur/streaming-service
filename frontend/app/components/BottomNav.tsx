import { NavLink } from "react-router";
import { Home, List } from "lucide-react";

export function BottomNav() {
  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 w-[28%]">
      <nav className="flex rounded-2xl border border-border bg-background/90 backdrop-blur-md shadow-lg shadow-black/30 overflow-hidden">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `flex flex-1 flex-col items-center gap-1 py-3 text-xs transition-colors ${
              isActive ? "text-foreground" : "text-muted-foreground"
            }`
          }
        >
          <Home size={20} />
          Home
        </NavLink>
        <NavLink
          to="/videos"
          className={({ isActive }) =>
            `flex flex-1 flex-col items-center gap-1 py-3 text-xs transition-colors ${
              isActive ? "text-foreground" : "text-muted-foreground"
            }`
          }
        >
          <List size={20} />
          Videos
        </NavLink>
      </nav>
    </div>
  );
}
