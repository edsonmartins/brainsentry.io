import { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import { ThemeProvider } from "./contexts/ThemeContext";
import { ToastProvider as ToastProviderComp } from "./components/ui/toast";
import { ErrorBoundary } from "./components/ui/error-boundary";
import { AdminLayout } from "./components/layout/AdminLayout";

const LoginPage = lazy(() => import("./pages/LoginPage").then(({ LoginPage }) => ({ default: LoginPage })));
const DashboardPage = lazy(() => import("./pages/DashboardPage").then(({ DashboardPage }) => ({ default: DashboardPage })));
const SearchPage = lazy(() => import("./pages/SearchPage").then(({ SearchPage }) => ({ default: SearchPage })));
const RelationshipsPage = lazy(() => import("./pages/RelationshipsPage").then(({ RelationshipsPage }) => ({ default: RelationshipsPage })));
const AuditPage = lazy(() => import("./pages/AuditPage").then(({ AuditPage }) => ({ default: AuditPage })));
const ConfigurationPage = lazy(() => import("./pages/ConfigurationPage").then(({ ConfigurationPage }) => ({ default: ConfigurationPage })));
const UsersPage = lazy(() => import("./pages/UsersPage").then(({ UsersPage }) => ({ default: UsersPage })));
const TenantsPage = lazy(() => import("./pages/TenantsPage").then(({ TenantsPage }) => ({ default: TenantsPage })));
const MemoryAdminPage = lazy(() => import("./pages/MemoryAdminPage"));
const AnalyticsAdminPage = lazy(() => import("./pages/AnalyticsAdminPage"));
const ProfilePage = lazy(() => import("./pages/ProfilePage"));
const PlaygroundPage = lazy(() => import("./pages/PlaygroundPage"));
const ConnectorsPage = lazy(() => import("./pages/ConnectorsPage"));
const NotesPage = lazy(() => import("./pages/NotesPage"));
const TasksPage = lazy(() => import("./pages/TasksPage"));
const TimelinePage = lazy(() => import("./pages/TimelinePage"));
const ConsolePage = lazy(() => import("./pages/ConsolePage"));
const AgentTracesPage = lazy(() => import("./pages/AgentTracesPage"));
const ExtractionLabPage = lazy(() => import("./pages/ExtractionLabPage"));
const OntologyPage = lazy(() => import("./pages/OntologyPage"));
const SessionCachePage = lazy(() => import("./pages/SessionCachePage"));
const ActionsPage = lazy(() => import("./pages/ActionsPage"));
const MeshPage = lazy(() => import("./pages/MeshPage"));
const BatchSearchPage = lazy(() => import("./pages/BatchSearchPage"));
const DecisionsPage = lazy(() => import("./pages/DecisionsPage"));
const PoliciesPage = lazy(() => import("./pages/PoliciesPage"));
const EventsPage = lazy(() => import("./pages/EventsPage"));
const ReasoningPage = lazy(() => import("./pages/ReasoningPage"));
const ProvenancePage = lazy(() => import("./pages/ProvenancePage"));
const GraphGlobalPage = lazy(() => import("./pages/GraphGlobalPage"));
const GraphEgoPage = lazy(() => import("./pages/GraphEgoPage"));
const GraphTimelinePage = lazy(() => import("./pages/GraphTimelinePage"));
const DiagnosticsPage = lazy(() => import("./pages/DiagnosticsPage"));
const ModelsPage = lazy(() => import("./pages/ModelsPage"));
const LandingPage = lazy(() => import("./landing/pages/LandingPage").then(({ LandingPage }) => ({ default: LandingPage })));

function RouteLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center" aria-busy="true">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" aria-hidden="true" />
    </div>
  );
}

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
            <Suspense fallback={<RouteLoading />}>
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
                <Route path="profile" element={<ProfilePage />} />
                <Route path="playground" element={<PlaygroundPage />} />
                <Route path="connectors" element={<ConnectorsPage />} />
                <Route path="notes" element={<NotesPage />} />
                <Route path="tasks" element={<TasksPage />} />
                <Route path="timeline" element={<TimelinePage />} />
                <Route path="console" element={<ConsolePage />} />
                <Route path="traces" element={<AgentTracesPage />} />
                <Route path="extraction" element={<ExtractionLabPage />} />
                <Route path="ontology" element={<OntologyPage />} />
                <Route path="session-cache" element={<SessionCachePage />} />
                <Route path="actions" element={<ActionsPage />} />
                <Route path="mesh" element={<MeshPage />} />
                <Route path="batch-search" element={<BatchSearchPage />} />
                <Route path="decisions" element={<DecisionsPage />} />
                <Route path="policies" element={<PoliciesPage />} />
                <Route path="events" element={<EventsPage />} />
                <Route path="reasoning" element={<ReasoningPage />} />
                <Route path="provenance" element={<ProvenancePage />} />
                <Route path="graph/global" element={<GraphGlobalPage />} />
                <Route path="graph/ego" element={<GraphEgoPage />} />
                <Route path="graph/timeline" element={<GraphTimelinePage />} />
                <Route path="diagnostics" element={<DiagnosticsPage />} />
                <Route path="models" element={<ModelsPage />} />
              </Route>

              {/* Legacy redirect - /dashboard -> /app/dashboard */}
              <Route path="/dashboard" element={<Navigate to="/app/dashboard" replace />} />

              {/* Catch all - redirect to landing */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
            </Suspense>
          </AuthProvider>
        </ToastProviderComp>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
