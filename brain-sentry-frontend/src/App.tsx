import { Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import { ToastProvider, ToastProvider as ToastProviderComp } from "./components/ui/toast";
import { ErrorBoundary } from "./components/ui/error-boundary";
import { AdminLayout } from "./components/layout/AdminLayout";
import { LoginPage } from "./pages/LoginPage";
import { DashboardPage } from "./pages/DashboardPage";
import { SearchPage } from "./pages/SearchPage";
import { RelationshipsPage } from "./pages/RelationshipsPage";
import { AuditPage } from "./pages/AuditPage";
import { ConfigurationPage } from "./pages/ConfigurationPage";
import { UsersPage } from "./pages/UsersPage";
import { TenantsPage } from "./pages/TenantsPage";
import MemoryAdminPage from "./pages/MemoryAdminPage";
import AnalyticsAdminPage from "./pages/AnalyticsAdminPage";
import { LandingPage } from "./landing/pages/LandingPage";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="mt-4 text-muted-foreground">Carregando...</p>
        </div>
      </div>
    );
  }

  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <ToastProviderComp>
          <AuthProvider>
            <Routes>
              {/* Landing Page - Public */}
              <Route path="/" element={<LandingPage />} />

              {/* Login Page */}
              <Route path="/login" element={<LoginPage />} />

              {/* Protected Routes - App */}
              <Route
                path="/app"
                element={
                  <ProtectedRoute>
                    <AdminLayout />
                  </ProtectedRoute>
                }
              >
                <Route index element={<Navigate to="/app/dashboard" replace />} />
                <Route path="dashboard" element={<DashboardPage />} />
                <Route path="memories" element={<MemoryAdminPage />} />
                <Route path="search" element={<SearchPage />} />
                <Route path="relationships" element={<RelationshipsPage />} />
                <Route path="audit" element={<AuditPage />} />
                <Route path="configuration" element={<ConfigurationPage />} />
                <Route path="users" element={<UsersPage />} />
                <Route path="tenants" element={<TenantsPage />} />
                <Route path="analytics" element={<AnalyticsAdminPage />} />
              </Route>

              {/* Legacy redirect - /dashboard -> /app/dashboard */}
              <Route path="/dashboard" element={<Navigate to="/app/dashboard" replace />} />

              {/* Catch all - redirect to landing */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </AuthProvider>
        </ToastProviderComp>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
