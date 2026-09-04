import { type ReactNode } from "react";
import { cx } from "../cx";

export interface StatCardProps {
  label: ReactNode;
  value: ReactNode;
  sublabel?: ReactNode;
  icon?: ReactNode;
  tint?: "none" | "success" | "info" | "goal" | "warning";
  className?: string;
}

/**
 * KPI card (design-system.md §4.6). Number + label (+ optional icon / tint).
 * Ratified exclusions: NO period-over-period delta, NO trend sparkline
 * (requirements §13 — "no comparison between two ranges", "no trend lines").
 */
export function StatCard({ label, value, sublabel, icon, tint = "none", className }: StatCardProps) {
  return (
    <div className={cx("ui-stat", tint !== "none" && `ui-stat--tint-${tint}`, className)}>
      <div className="ui-stat__top">
        {icon && <span className="ui-stat__icon">{icon}</span>}
        <span className="ui-stat__value">{value}</span>
      </div>
      <div>
        <div className="ui-stat__label">{label}</div>
        {sublabel && <div className="ui-stat__sublabel">{sublabel}</div>}
      </div>
    </div>
  );
}
