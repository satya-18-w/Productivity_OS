import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { LeafMark } from "./LeafMark";
import { SidebarNavItem } from "./SidebarNavItem";
import { ThemeToggle } from "./ThemeToggle";
import { UserMenu } from "./UserMenu";
import { NAV_ITEMS } from "./navItems";
import type { ShellMode } from "./useShellState";

interface ContentProps {
  collapsed: boolean;
  onNavigate?: () => void;
}

function SidebarContent({ collapsed, onNavigate }: ContentProps) {
  return (
    <>
      <Link className="sidebar__brand" to="/" onClick={onNavigate} aria-label="Productivity OS — home">
        <span className="sidebar__brand-tile" aria-hidden="true">
          <LeafMark />
        </span>
        {!collapsed && <span className="sidebar__brand-text">Productivity OS</span>}
      </Link>

      <nav className="sidebar__nav" aria-label="Primary">
        {NAV_ITEMS.map((item) => (
          <SidebarNavItem key={item.to} item={item} collapsed={collapsed} onNavigate={onNavigate} />
        ))}
      </nav>

      <div className="sidebar__spacer" />

      <div className="sidebar__footer">
        <ThemeToggle compact={collapsed} />
        <UserMenu compact={collapsed} />
      </div>
    </>
  );
}

export function Sidebar({ mode }: { mode: Exclude<ShellMode, "drawer"> }) {
  return (
    <aside className="sidebar">
      <SidebarContent collapsed={mode === "collapsed"} />
    </aside>
  );
}

export function SidebarDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (open && !el.open) el.showModal();
    else if (!open && el.open) el.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      className="sidebar-drawer"
      aria-label="Navigation"
      onCancel={(e) => {
        e.preventDefault();
        onClose();
      }}
      onClose={onClose}
      onClick={(e) => {
        if (e.target === ref.current) onClose();
      }}
    >
      <div className="sidebar">
        <SidebarContent collapsed={false} onNavigate={onClose} />
      </div>
    </dialog>
  );
}
