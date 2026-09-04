import type { ReactNode } from "react";
import { cx } from "../components/cx";

export interface ScreenLayoutProps {
  /** Main column content — typically a <PageHeader> followed by the body. */
  children: ReactNode;
  /** Optional right-rail content (contextual widgets). */
  rail?: ReactNode;
  /** Accessible name for the rail region. */
  railLabel?: string;
  className?: string;
}

/**
 * The wrapper every screen renders inside the app shell's <main>. Provides the
 * main column + optional right rail (D3). The rail sits beside main at >= wide
 * and stacks below it otherwise (D4 — rail sheds first).
 */
export function ScreenLayout({ children, rail, railLabel = "Related information", className }: ScreenLayoutProps) {
  return (
    <div className={cx("screen", rail != null && "screen--has-rail", className)}>
      <div className="screen__main">{children}</div>
      {rail != null && (
        <aside className="screen__rail" aria-label={railLabel}>
          {rail}
        </aside>
      )}
    </div>
  );
}
