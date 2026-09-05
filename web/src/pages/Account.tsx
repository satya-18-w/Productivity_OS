import { useEffect, useId, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api, ApiError, type FieldErrors } from "../api";
import { useAuth } from "../auth";
import { PageHeader } from "../components/layout/PageHeader";
import { Card } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Field } from "../components/ui/Field";
import { Input } from "../components/ui/Input";
import { TimezoneSelect } from "../features/reviews/TimezoneSelect";

const formStack = { display: "flex", flexDirection: "column", gap: "var(--sp-4)" } as const;

/**
 * Account (`v1.md §1` + Q4/Q6). Rendered inside `ScreenLayout` by `App.tsx` —
 * this component adds the `PageHeader` + section cards, not another layout.
 * Email (display), timezone `Select` (Q4), password change with confirm (Q6),
 * log out. No profile fields beyond email/password/timezone (§1 boundary).
 */
export function Account() {
  const { account, refresh, setAccount } = useAuth();
  const navigate = useNavigate();
  const [loggingOut, setLoggingOut] = useState(false);

  async function logout() {
    setLoggingOut(true);
    try {
      await api.logout();
    } finally {
      setAccount(null);
      navigate("/login", { replace: true });
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Account"
        title="Account"
        subtitle="Your email, timezone, password, and sign-out."
      />

      <Card title="Account" headingLevel={2}>
        <dl>
          <dt>Email</dt>
          <dd>{account?.email}</dd>
          <dt>Timezone</dt>
          <dd>{account?.timezone}</dd>
        </dl>
      </Card>

      <TimezoneForm current={account?.timezone ?? ""} onSaved={refresh} />
      <PasswordForm
        onChanged={() => {
          setAccount(null);
          navigate("/login", { replace: true });
        }}
      />

      <Card title="Log out" headingLevel={2}>
        <p className="muted">Sign out of Productivity OS on this device.</p>
        <div>
          <Button variant="secondary" onClick={logout} loading={loggingOut}>
            Log out
          </Button>
        </div>
      </Card>
    </>
  );
}

function TimezoneForm({ current, onSaved }: { current: string; onSaved: () => Promise<void> }) {
  const formId = useId();
  const [timezone, setTimezone] = useState(current);
  const [status, setStatus] = useState<"" | "saved" | "error">("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setTimezone(current);
  }, [current]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setStatus("");
    setFieldError("");
    setBusy(true);
    try {
      await api.setTimezone(timezone);
      await onSaved();
      setStatus("saved");
    } catch (err) {
      if (err instanceof ApiError && err.fields?.timezone) setFieldError(err.fields.timezone);
      else setStatus("error");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card title="Change timezone" headingLevel={2}>
      <form id={formId} onSubmit={submit} style={formStack}>
        <Field label="IANA timezone" htmlFor={`${formId}-timezone`} error={fieldError || undefined}>
          <TimezoneSelect
            id={`${formId}-timezone`}
            required
            value={timezone}
            onChange={setTimezone}
            invalid={!!fieldError}
          />
        </Field>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: "var(--sp-3)" }}>
          {status === "saved" && (
            <span className="ok" role="status">
              Saved.
            </span>
          )}
          {status === "error" && (
            <span className="field-error" role="alert">
              Could not save.
            </span>
          )}
          <Button type="submit" loading={busy}>
            Save
          </Button>
        </div>
      </form>
    </Card>
  );
}

const PASSWORD_HINT =
  "At least 6 characters, with a lowercase letter, an uppercase letter, and a special character. You will be signed out.";

function PasswordForm({ onChanged }: { onChanged: () => void }) {
  const formId = useId();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [fields, setFields] = useState<FieldErrors>({});
  const [confirmError, setConfirmError] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setConfirmError("");
    setError("");
    if (next !== confirm) {
      setConfirmError("Passwords do not match.");
      return;
    }
    setBusy(true);
    try {
      await api.changePassword(current, next);
      onChanged();
    } catch (err) {
      if (err instanceof ApiError && err.code === "VALIDATION_ERROR") setFields(err.fields ?? {});
      else if (err instanceof ApiError && err.status === 401) setError("Current password is incorrect.");
      else setError("Could not change the password.");
      setBusy(false);
    }
  }

  return (
    <Card title="Change password" headingLevel={2}>
      <form id={formId} onSubmit={submit} style={formStack}>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <Field label="Current password" htmlFor={`${formId}-current`} required>
          <Input
            id={`${formId}-current`}
            type="password"
            autoComplete="current-password"
            required
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
          />
        </Field>
        <Field
          label="New password"
          htmlFor={`${formId}-new`}
          required
          hint={PASSWORD_HINT}
          error={fields.new_password}
        >
          <Input
            id={`${formId}-new`}
            type="password"
            autoComplete="new-password"
            required
            minLength={6}
            value={next}
            invalid={!!fields.new_password}
            aria-describedby={fields.new_password ? undefined : `${formId}-new-hint`}
            onChange={(e) => setNext(e.target.value)}
          />
        </Field>
        <Field label="Confirm new password" htmlFor={`${formId}-confirm`} required error={confirmError || undefined}>
          <Input
            id={`${formId}-confirm`}
            type="password"
            autoComplete="new-password"
            required
            value={confirm}
            invalid={!!confirmError}
            onChange={(e) => setConfirm(e.target.value)}
          />
        </Field>
        <div style={{ display: "flex", justifyContent: "flex-end" }}>
          <Button type="submit" loading={busy}>
            Change password
          </Button>
        </div>
      </form>
    </Card>
  );
}
