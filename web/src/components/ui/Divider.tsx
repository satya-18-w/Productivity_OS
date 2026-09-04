import { cx } from "../cx";

export interface DividerProps {
  orientation?: "horizontal" | "vertical";
  className?: string;
}

export function Divider({ orientation = "horizontal", className }: DividerProps) {
  return (
    <hr
      className={cx("ui-divider", orientation === "vertical" && "ui-divider--vertical", className)}
      aria-orientation={orientation}
    />
  );
}
