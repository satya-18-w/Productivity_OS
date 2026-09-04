import type { SVGProps } from "react";

/** The Productivity OS brand glyph — a leaf. Currentcolor fill. */
export function LeafMark(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width={16} height={16} fill="currentColor" aria-hidden="true" {...props}>
      <path d="M20 3c-8 0-14 4.5-14 12 0 1.4.3 2.7.8 3.9L4 22l1.4-1.4C7 22 9 22.5 11 22.5 18 22.5 21 15 21 7c0-1.5-.4-2.9-1-4zM8.5 16.5C10 12 13.5 8.8 18 7c-3.2 3-5.8 6.8-7.5 11z" />
    </svg>
  );
}
