import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { AppShell } from "./shell/AppShell";
import { ScreenLayout } from "./shell/ScreenLayout";
import { Placeholder } from "./shell/Placeholder";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Account } from "./pages/Account";
import { Board } from "./pages/Board";
import { Categories } from "./pages/Categories";
import { Goals } from "./pages/Goals";
import { Habits } from "./pages/Habits";
import { Timeline } from "./pages/Timeline";

/**
 * Routes (D10). Authenticated screens render inside <AppShell>. Screens not yet
 * rebuilt for the new design system render either their existing page wrapped in
 * <ScreenLayout> (Timeline / Board / Habits / Goals / Categories / Account) or a
 * <Placeholder> (Tasks / Reports / Reviews / Export). No /dashboard, /notes,
 * /calendar, /timeline/week|month (design-system.md §6.4).
 */
export function App() {
  const { account, loading } = useAuth();

  if (loading) {
    return (
      <div className="center">
        <span className="brand" style={{ opacity: 0.6 }}>Productivity OS</span>
      </div>
    );
  }

  return (
    <Routes>
      <Route path="/login" element={account ? <Navigate to="/" replace /> : <Login />} />
      <Route path="/register" element={account ? <Navigate to="/" replace /> : <Register />} />

      <Route element={account ? <AppShell /> : <Navigate to="/login" replace />}>
        <Route path="/" element={<Navigate to="/timeline" replace />} />
        <Route path="/timeline" element={<ScreenLayout><Timeline /></ScreenLayout>} />
        <Route path="/tasks" element={<Placeholder name="Tasks" phase={4} />} />
        <Route path="/board" element={<ScreenLayout><Board /></ScreenLayout>} />
        <Route path="/habits" element={<ScreenLayout><Habits /></ScreenLayout>} />
        <Route path="/goals" element={<ScreenLayout><Goals /></ScreenLayout>} />
        <Route path="/categories" element={<ScreenLayout><Categories /></ScreenLayout>} />
        <Route path="/reports" element={<Placeholder name="Reports" phase={9} />} />
        <Route path="/reviews/daily" element={<Placeholder name="Daily review" phase={10} />} />
        <Route path="/reviews/weekly" element={<Placeholder name="Weekly review" phase={11} />} />
        <Route path="/account" element={<ScreenLayout><Account /></ScreenLayout>} />
        <Route path="/export" element={<Placeholder name="Data export" phase={14} />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
