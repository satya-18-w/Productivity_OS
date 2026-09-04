import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api";
import { useAuth } from "../auth";
import { Avatar } from "../components/ui/Avatar";
import { UserIcon, DownloadIcon, LogOutIcon } from "../components/ui/icons";

export interface UserMenuProps {
  /** Collapsed sidebar → avatar only, no name. */
  compact?: boolean;
}

export function UserMenu({ compact }: UserMenuProps) {
  const { account, setAccount } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);

  const name = account?.email ?? "Account";

  useEffect(() => {
    if (!open) return;
    function onDocClick(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setOpen(false);
        triggerRef.current?.focus();
      }
    }
    document.addEventListener("mousedown", onDocClick);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDocClick);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  async function logout() {
    setOpen(false);
    try {
      await api.logout();
    } finally {
      setAccount(null);
      navigate("/login", { replace: true });
    }
  }

  return (
    <div className="user-menu" ref={rootRef}>
      <button
        ref={triggerRef}
        type="button"
        className="user-menu__trigger"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={compact ? name : undefined}
        onClick={() => setOpen((v) => !v)}
      >
        <Avatar name={name} size="sm" decorative={!compact} />
        {!compact && <span className="user-menu__name">{name}</span>}
      </button>

      {open && (
        <div className="user-menu__list" role="menu" aria-label="Account menu">
          <Link className="user-menu__item" role="menuitem" to="/account" onClick={() => setOpen(false)}>
            <UserIcon /> Account
          </Link>
          <Link className="user-menu__item" role="menuitem" to="/export" onClick={() => setOpen(false)}>
            <DownloadIcon /> Export data
          </Link>
          <button className="user-menu__item" role="menuitem" type="button" onClick={logout}>
            <LogOutIcon /> Log out
          </button>
        </div>
      )}
    </div>
  );
}
