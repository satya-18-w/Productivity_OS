import type { ComponentType, SVGProps } from "react";
import {
  TimelineIcon,
  TasksIcon,
  BoardIcon,
  HabitsIcon,
  GoalsIcon,
  TagIcon,
  ReportsIcon,
  ReviewsIcon,
} from "../components/ui/icons";

export interface NavItem {
  label: string;
  to: string;
  /** `end` matches the route exactly (used for index-ish routes). */
  end?: boolean;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
}

/**
 * Primary sidebar navigation (D10). V1 only — no Dashboard / Notes / Calendar /
 * Analytics / Spaces (design-system.md §6.4). Account / Export / Log out live in
 * the sidebar footer, not here.
 */
export const NAV_ITEMS: NavItem[] = [
  { label: "Timeline", to: "/timeline", icon: TimelineIcon },
  { label: "Tasks", to: "/tasks", icon: TasksIcon },
  { label: "Board", to: "/board", icon: BoardIcon },
  { label: "Habits", to: "/habits", icon: HabitsIcon },
  { label: "Goals", to: "/goals", icon: GoalsIcon },
  { label: "Categories", to: "/categories", icon: TagIcon },
  { label: "Reports", to: "/reports", icon: ReportsIcon },
  { label: "Reviews", to: "/reviews/daily", icon: ReviewsIcon },
];
