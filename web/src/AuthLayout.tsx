import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { api } from "./api";
import { useAuth } from "./auth";

export function AuthLayout() {
  const { setAccount } = useAuth();
  const navigate = useNavigate();

  async function logout() {
    try {
      await api.logout();
    } finally {
      setAccount(null);
      navigate("/login", { replace: true });
    }
  }

  return (
    <div className="app">
      <nav className="nav">
        <span className="brand">Productivity OS</span>
        <div className="nav-links">
          <NavLink to="/" end>Timeline</NavLink>
          <NavLink to="/board">Board</NavLink>
          <NavLink to="/habits">Habits</NavLink>
          <NavLink to="/categories">Categories</NavLink>
          <NavLink to="/account">Account</NavLink>
        </div>
        <button className="link" onClick={logout}>Log out</button>
      </nav>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
