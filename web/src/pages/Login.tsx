import { useId, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { useAuth } from "../auth";
import { Card } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Field } from "../components/ui/Field";
import { Input } from "../components/ui/Input";
import { LeafMark } from "../shell/LeafMark";

/**
 * Login (`v1.md §1`). Calm centered card, no shell (app-shell.md route table).
 * Visual language: `overall.png` panel 12. One brand primary, tokens only.
 */
export function Login() {
  const { refresh } = useAuth();
  const navigate = useNavigate();
  const formId = useId();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.login(email, password);
      await refresh();
      navigate("/", { replace: true });
    } catch (err) {
      setError(
        err instanceof ApiError && err.status === 429
          ? "Too many attempts. Please wait a few minutes."
          : "Incorrect email or password.",
      );
      setBusy(false);
    }
  }

  return (
    <div className="center">
      <Card style={{ width: "100%", maxWidth: 384 }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--sp-2)" }}>
          <span
            style={{
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              width: 22,
              height: 22,
              borderRadius: "var(--radius-sm)",
              background: "var(--brand)",
              color: "var(--on-brand)",
            }}
            aria-hidden="true"
          >
            <LeafMark />
          </span>
          <span style={{ fontWeight: "var(--fw-semibold)", fontSize: "var(--fs-small)" }}>
            Productivity OS
          </span>
        </div>
        <div>
          <h1>Welcome back</h1>
          <p className="secondary">Log in to plan your day.</p>
        </div>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <form id={formId} onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: "var(--sp-4)" }}>
          <Field label="Email" htmlFor={`${formId}-email`} required>
            <Input
              id={`${formId}-email`}
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </Field>
          <Field label="Password" htmlFor={`${formId}-password`} required>
            <Input
              id={`${formId}-password`}
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </Field>
          <Button type="submit" block loading={busy}>
            {busy ? "Logging in…" : "Log in"}
          </Button>
        </form>
        <p className="muted">
          No account? <Link to="/register">Create one</Link>
        </p>
      </Card>
    </div>
  );
}
