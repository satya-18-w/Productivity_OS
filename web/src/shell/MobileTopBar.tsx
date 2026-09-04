import { IconButton } from "../components/ui/IconButton";
import { MenuIcon } from "../components/ui/icons";
import { LeafMark } from "./LeafMark";

export function MobileTopBar({ onOpenMenu }: { onOpenMenu: () => void }) {
  return (
    <header className="mobile-topbar">
      <IconButton label="Open navigation" onClick={onOpenMenu}>
        <MenuIcon />
      </IconButton>
      <span className="mobile-topbar__brand">
        <span className="sidebar__brand-tile" aria-hidden="true">
          <LeafMark />
        </span>
        Productivity OS
      </span>
    </header>
  );
}
