import { Navigate, Route, Routes } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import { useAuth } from "./auth";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Account } from "./pages/Account";
import { Board } from "./pages/Board";
import { Categories } from "./pages/Categories";
import { Habits } from "./pages/Habits";
import { Timeline } from "./pages/Timeline";

export function App() {
  const { account, loading } = useAuth();

  if (loading) {
    return <div className="center muted">Loading…</div>;
  }

  return (
    <Routes>
      <Route path="/login" element={account ? <Navigate to="/" replace /> : <Login />} />
      <Route path="/register" element={account ? <Navigate to="/" replace /> : <Register />} />

      <Route element={account ? <AuthLayout /> : <Navigate to="/login" replace />}>
        <Route path="/" element={<Timeline />} />
        <Route path="/board" element={<Board />} />
        <Route path="/habits" element={<Habits />} />
        <Route path="/account" element={<Account />} />
        <Route path="/categories" element={<Categories />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
