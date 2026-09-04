import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api, ApiError } from "../api";
import { useAuth } from "../auth";

export function Login() {
  const { refresh } = useAuth();
  const navigate = useNavigate();
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
      <form className="card" onSubmit={submit}>
        <div className="auth-brand"><span className="brand">Productivity OS</span></div>
        <div>
          <h1>Welcome back</h1>
          <p className="secondary">Log in to plan your day.</p>
        </div>
        {error && <p className="error" role="alert">{error}</p>}
        <label>
          Email
          <input type="email" autoComplete="email" required value={email}
            onChange={(e) => setEmail(e.target.value)} />
        </label>
        <label>
          Password
          <input type="password" autoComplete="current-password" required value={password}
            onChange={(e) => setPassword(e.target.value)} />
        </label>
        <button type="submit" disabled={busy}>{busy ? "Logging in…" : "Log in"}</button>
        <p className="muted">No account? <Link to="/register">Create one</Link></p>
      </form>
    </div>
  );
}
