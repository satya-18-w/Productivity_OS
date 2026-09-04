import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Shell } from "./pages/Shell";

export function App() {
  const { account, loading } = useAuth();

  if (loading) {
    return <div className="center muted">Loading…</div>;
  }

  return (
    <Routes>
      <Route path="/login" element={account ? <Navigate to="/" replace /> : <Login />} />
      <Route path="/register" element={account ? <Navigate to="/" replace /> : <Register />} />
      <Route path="/" element={account ? <Shell /> : <Navigate to="/login" replace />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
