import { NavLink } from "react-router-dom";
import { cx } from "../components/cx";
import { Tooltip } from "../components/ui/Tooltip";
import type { NavItem } from "./navItems";

export interface SidebarNavItemProps {
  item: NavItem;
  /** Icon-only mode — label is kept for assistive tech, shown as a tooltip. */
  collapsed?: boolean;
  /** Called on navigation (used to close the mobile drawer). */
  onNavigate?: () => void;
}

export function SidebarNavItem({ item, collapsed, onNavigate }: SidebarNavItemProps) {
  const { to, end, label, icon: Icon } = item;

  const link = (
    <NavLink
      to={to}
      end={end}
      onClick={onNavigate}
      className={({ isActive }) => cx("sidebar__item", isActive && "active")}
    >
      <Icon />
      <span className={collapsed ? "ui-visually-hidden" : "sidebar__label"}>{label}</span>
    </NavLink>
  );

  return collapsed ? <Tooltip content={label}>{link}</Tooltip> : link;
}
