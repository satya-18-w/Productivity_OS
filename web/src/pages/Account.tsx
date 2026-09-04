import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api, ApiError, type FieldErrors } from "../api";
import { useAuth } from "../auth";

export function Account() {
  const { account, refresh, setAccount } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="stack">
      <section className="card">
        <h2>Account</h2>
        <dl>
          <dt>Email</dt>
          <dd>{account?.email}</dd>
          <dt>Timezone</dt>
          <dd>{account?.timezone}</dd>
        </dl>
      </section>

      <TimezoneForm current={account?.timezone ?? ""} onSaved={refresh} />
      <PasswordForm
        onChanged={() => {
          setAccount(null);
          navigate("/login", { replace: true });
        }}
      />
    </div>
  );
}

function TimezoneForm({ current, onSaved }: { current: string; onSaved: () => Promise<void> }) {
  const [timezone, setTimezone] = useState(current);
  const [status, setStatus] = useState<"" | "saved" | "error">("");
  const [fieldError, setFieldError] = useState("");
  const [busy, setBusy] = useState(false);

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
    <form className="card" onSubmit={submit}>
      <h2>Change timezone</h2>
      <label>
        IANA timezone
        <input type="text" required value={timezone} onChange={(e) => setTimezone(e.target.value)} />
        {fieldError && <span className="field-error">{fieldError}</span>}
      </label>
      <button type="submit" disabled={busy}>Save</button>
      {status === "saved" && <span className="ok">Saved.</span>}
      {status === "error" && <span className="field-error">Could not save.</span>}
    </form>
  );
}

function PasswordForm({ onChanged }: { onChanged: () => void }) {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [fields, setFields] = useState<FieldErrors>({});
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setError("");
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
    <form className="card" onSubmit={submit}>
      <h2>Change password</h2>
      {error && <p className="error" role="alert">{error}</p>}
      <label>
        Current password
        <input
          type="password"
          autoComplete="current-password"
          required
          value={current}
          onChange={(e) => setCurrent(e.target.value)}
        />
      </label>
      <label>
        New password
        <input
          type="password"
          autoComplete="new-password"
          required
          minLength={6}
          value={next}
          onChange={(e) => setNext(e.target.value)}
        />
        {fields.new_password ? (
          <span className="field-error">{fields.new_password}</span>
        ) : (
          <span className="hint">
            At least 6 characters, with a lowercase letter, an uppercase letter, and a special
            character. You will be signed out.
          </span>
        )}
      </label>
      <button type="submit" disabled={busy}>Change password</button>
    </form>
  );
}
