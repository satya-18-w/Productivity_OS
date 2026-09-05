import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { AppShell } from "./shell/AppShell";
import { ScreenLayout } from "./shell/ScreenLayout";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";
import { Account } from "./pages/Account";
import { TimelineScreen } from "./features/timeline";
import { TasksScreen } from "./features/tasks";
import { BoardScreen } from "./features/board";
import { HabitsScreen } from "./features/habits";
import { GoalsScreen } from "./features/goals";
import { CategoriesScreen } from "./features/categories";
import { ReportsScreen } from "./features/reports";
import { DailyReviewScreen, WeeklyReviewScreen } from "./features/reviews";
import { ExportScreen } from "./features/export";

/**
 * Routes (D10). Authenticated screens render inside <AppShell>.
 * Built: Timeline, Tasks, Board, Habits, Goals, Categories, Reports, Daily +
 * Weekly Review, Account, Auth, Export. Reports + both reviews run on
 * mock/placeholder data and Export is provisional JSON — backends pending
 * (see docs/left.md). No /dashboard, /notes, /calendar, /timeline/week|month
 * (design-system.md §6.4).
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
        <Route path="/timeline" element={<TimelineScreen />} />
        <Route path="/tasks" element={<TasksScreen />} />
        <Route path="/board" element={<BoardScreen />} />
        <Route path="/habits" element={<HabitsScreen />} />
        <Route path="/goals" element={<GoalsScreen />} />
        <Route path="/categories" element={<CategoriesScreen />} />
        <Route path="/reports" element={<ReportsScreen />} />
        <Route path="/reviews/daily" element={<DailyReviewScreen />} />
        <Route path="/reviews/weekly" element={<WeeklyReviewScreen />} />
        <Route path="/account" element={<ScreenLayout><Account /></ScreenLayout>} />
        <Route path="/export" element={<ExportScreen />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
