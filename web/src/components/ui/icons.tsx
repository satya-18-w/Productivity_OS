/**
 * Minimal inline icon set for the primitives. Kept tiny and local — no icon
 * library (dependency discipline). Feature work can introduce a proper set as
 * its own decision.
 */
import { type SVGProps } from "react";

const base: SVGProps<SVGSVGElement> = {
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 2.25,
  strokeLinecap: "round",
  strokeLinejoin: "round",
  "aria-hidden": true,
  focusable: false,
};

export const CheckIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p}>
    <path d="M20 6 9 17l-5-5" />
  </svg>
);

export const ChevronDownIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <path d="m6 9 6 6 6-6" />
  </svg>
);

export const CloseIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p}>
    <path d="M18 6 6 18M6 6l12 12" />
  </svg>
);

export const AlertIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <circle cx="12" cy="12" r="9" />
    <path d="M12 8v5M12 16.5h.01" />
  </svg>
);

export const InboxIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.75}>
    <path d="M4 13h4l2 3h4l2-3h4" />
    <path d="M5 13 7 5h10l2 8v5a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2z" />
  </svg>
);

export const ChevronRightIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <path d="m9 6 6 6-6 6" />
  </svg>
);

export const MenuIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <path d="M4 6h16M4 12h16M4 18h16" />
  </svg>
);

export const LogOutIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" />
  </svg>
);

/* --- theme --- */
export const SunIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
  </svg>
);

export const MoonIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
  </svg>
);

export const MonitorIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={2}>
    <rect x="3" y="4" width="18" height="12" rx="2" />
    <path d="M8 20h8M12 16v4" />
  </svg>
);

/* --- primary nav --- */
export const TimelineIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <rect x="3" y="4" width="18" height="17" rx="2" />
    <path d="M3 9h18M8 2v4M16 2v4" />
    <circle cx="12" cy="15" r="3" />
    <path d="M12 13.5V15l1 1" />
  </svg>
);

export const TasksIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <rect x="3" y="3" width="18" height="18" rx="3" />
    <path d="m8 12 3 3 5-6" />
  </svg>
);

export const BoardIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <rect x="3" y="3" width="18" height="18" rx="2" />
    <path d="M9 3v18M15 3v18" />
  </svg>
);

export const HabitsIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M17 3l3 3-3 3" />
    <path d="M4 11V9a4 4 0 0 1 4-4h12" />
    <path d="M7 21l-3-3 3-3" />
    <path d="M20 13v2a4 4 0 0 1-4 4H4" />
  </svg>
);

export const GoalsIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <circle cx="12" cy="12" r="9" />
    <circle cx="12" cy="12" r="5" />
    <circle cx="12" cy="12" r="1.5" />
  </svg>
);

export const TagIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M12 2H5a3 3 0 0 0-3 3v7l9.3 9.3a2.4 2.4 0 0 0 3.4 0l6.6-6.6a2.4 2.4 0 0 0 0-3.4L12 2z" />
    <circle cx="7.5" cy="7.5" r="1.5" />
  </svg>
);

export const ReportsIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M3 3v18h18" />
    <path d="M7 15v-3M12 15V9M17 15V6" />
  </svg>
);

export const ReviewsIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M4 5a2 2 0 0 1 2-2h12v18H6a2 2 0 0 1-2-2z" />
    <path d="M4 5h2M4 9h2M4 13h2M4 17h2M10 8h5M10 12h5" />
  </svg>
);

export const UserIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <circle cx="12" cy="8" r="4" />
    <path d="M4 21c0-4 3.6-6 8-6s8 2 8 6" />
  </svg>
);

export const DownloadIcon = (p: SVGProps<SVGSVGElement>) => (
  <svg {...base} {...p} strokeWidth={1.9}>
    <path d="M12 3v12M7 10l5 5 5-5M5 21h14" />
  </svg>
);
