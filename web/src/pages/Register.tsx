import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError, browserTimezone, type FieldErrors } from "../api";
import { useAuth } from "../auth";

export function Register() {
  const { refresh } = useAuth();
  const navigate = useNavigate();
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
      <form className="card" onSubmit={submit}>
        <div className="auth-brand"><span className="brand">Productivity OS</span></div>
        <div>
          <h1>Create your account</h1>
          <p className="secondary">Plan, track, and review where your time goes.</p>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
        <label>
          Email
          <input type="email" autoComplete="email" required value={email}
            onChange={(e) => setEmail(e.target.value)} />
          {fields.email && <span className="field-error">{fields.email}</span>}
        </label>
        <label>
          Password
          <input type="password" autoComplete="new-password" required minLength={6} value={password}
            onChange={(e) => setPassword(e.target.value)} />
          {fields.password
            ? <span className="field-error">{fields.password}</span>
            : <span className="hint">At least 6 characters, with a lowercase letter, an uppercase letter, and a special character.</span>}
        </label>
        <label>
          Timezone
          <input type="text" required value={timezone}
            onChange={(e) => setTimezone(e.target.value)} />
          {fields.timezone && <span className="field-error">{fields.timezone}</span>}
        </label>
        <button type="submit" disabled={busy}>{busy ? "Creating…" : "Create account"}</button>
        <p className="muted">Already have one? <Link to="/login">Log in</Link></p>
      </form>
    </div>
  );
}
