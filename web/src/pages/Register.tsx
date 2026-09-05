import { useId, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError, browserTimezone, type FieldErrors } from "../api";
import { useAuth } from "../auth";
import { Card } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Field } from "../components/ui/Field";
import { Input } from "../components/ui/Input";
import { LeafMark } from "../shell/LeafMark";
import { TimezoneSelect } from "../features/reviews/TimezoneSelect";

const PASSWORD_HINT =
  "At least 6 characters, with a lowercase letter, an uppercase letter, and a special character.";

/**
 * Register (`v1.md §1` + Q4/Q6). Calm centered card, no shell (app-shell.md
 * route table). Visual language: `overall.png` panel 12. Timezone defaults to
 * the browser-detected IANA name (Q4, fallback `UTC`); password policy hint is
 * Q6, enforced server-side with field errors mapped below.
 */
export function Register() {
  const { refresh } = useAuth();
  const navigate = useNavigate();
  const formId = useId();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [timezone, setTimezone] = useState(browserTimezone());
  const [fields, setFields] = useState<FieldErrors>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setFields({});
    setBusy(true);
    try {
      await api.register(email, password, timezone);
      await refresh();
      navigate("/", { replace: true });
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") {
        setFields(err.fields ?? {});
      } else if (err instanceof ApiError && err.code === "EMAIL_ALREADY_REGISTERED") {
        setError("An account with this email already exists.");
      } else if (err instanceof ApiError && err.status === 429) {
        setError("Too many attempts. Please wait a few minutes.");
      } else {
        setError("Could not create the account. Please try again.");
      }
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
          <h1>Create your account</h1>
          <p className="secondary">Plan, track, and review where your time goes.</p>
        </div>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <form id={formId} onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: "var(--sp-4)" }}>
          <Field label="Email" htmlFor={`${formId}-email`} required error={fields.email}>
            <Input
              id={`${formId}-email`}
              type="email"
              autoComplete="email"
              required
              value={email}
              invalid={!!fields.email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </Field>
          <Field
            label="Password"
            htmlFor={`${formId}-password`}
            required
            hint={PASSWORD_HINT}
            error={fields.password}
          >
            <Input
              id={`${formId}-password`}
              type="password"
              autoComplete="new-password"
              required
              minLength={6}
              value={password}
              invalid={!!fields.password}
              aria-describedby={fields.password ? undefined : `${formId}-password-hint`}
              onChange={(e) => setPassword(e.target.value)}
            />
          </Field>
          <Field label="Timezone" htmlFor={`${formId}-timezone`} required error={fields.timezone}>
            <TimezoneSelect
              id={`${formId}-timezone`}
              required
              value={timezone}
              onChange={setTimezone}
              invalid={!!fields.timezone}
            />
          </Field>
          <Button type="submit" block loading={busy}>
            {busy ? "Creating…" : "Create account"}
          </Button>
        </form>
        <p className="muted">
          Already have one? <Link to="/login">Log in</Link>
        </p>
      </Card>
    </div>
  );
}
