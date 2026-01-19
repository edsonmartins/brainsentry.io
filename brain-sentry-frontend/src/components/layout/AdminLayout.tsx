import { Outlet, useNavigate, useLocation } from "react-router-dom";
import {
  Menu,
  X,
  FileText,
  Activity,
  LayoutDashboard,
  Search,
  Network,
  Shield,
  Settings,
  Users,
  Building2,
} from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { ThemeSelector } from "@/components/ui/theme-selector";
import { useAuth } from "@/contexts/AuthContext";

export function AdminLayout() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();

  const navigation = [
    {
      title: "Dashboard",
      href: "/dashboard",
      icon: LayoutDashboard,
      id: "dashboard",
    },
    {
      title: "Memórias",
      href: "/memories",
      icon: FileText,
      id: "memories",
    },
    {
      title: "Busca",
      href: "/search",
      icon: Search,
      id: "search",
    },
    {
      title: "Relacionamentos",
      href: "/relationships",
      icon: Network,
      id: "relationships",
    },
    {
      title: "Auditoria",
      href: "/audit",
      icon: Shield,
      id: "audit",
    },
    {
      title: "Usuários",
      href: "/users",
      icon: Users,
      id: "users",
    },
    {
      title: "Tenants",
      href: "/tenants",
      icon: Building2,
      id: "tenants",
    },
    {
      title: "Configurações",
      href: "/configuration",
      icon: Settings,
      id: "configuration",
    },
    {
      title: "Analytics",
      href: "/analytics",
      icon: Activity,
      id: "analytics",
    },
  ];

  const handleNavigation = (href: string) => {
    navigate(href);
  };

  const activePath = navigation.find((item) =>
    location.pathname.startsWith(item.href)
  )?.id || "dashboard";

  return (
    <div className="min-h-screen flex flex-col md:flex-row">
      {/* Top bar for mobile */}
      <div className="md:hidden flex items-center justify-between p-4 border-b bg-background">
        <div className="flex items-center gap-2">
          <div className="h-8 w-8 bg-primary rounded-md" />
          <h1 className="text-xl font-bold">Brain Sentry</h1>
        </div>
        <div className="flex items-center gap-2">
          <ThemeSelector />
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="p-2 rounded-md hover:bg-muted"
          >
            {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* Sidebar for desktop */}
      <aside className="hidden md:flex w-64 flex-col border-r bg-muted/40">
        <div className="p-6 border-b">
          <h2 className="text-lg font-bold">Brain Sentry</h2>
          <p className="text-xs text-muted-foreground">Admin Console</p>
        </div>
        <nav className="flex-1 p-4 space-y-1">
          {navigation.map((item) => {
            const isActive = activePath === item.id;
            return (
              <button
                key={item.id}
                onClick={() => handleNavigation(item.href)}
                className={cn(
                  "w-full flex items-center gap-3 px-4 py-2 rounded-md text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground hover:bg-primary/90"
                    : "text-muted-foreground hover:bg-muted/50 hover:text-accent-foreground"
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.title}
              </button>
            );
          })}
        </nav>
        <div className="p-4 border-t">
          <div className="flex items-center justify-between">
            <div className="text-xs text-muted-foreground">
              v1.0.0
            </div>
            <ThemeSelector />
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {new Date().toLocaleDateString('pt-BR')}
          </div>
          {user && (
            <div className="text-xs text-muted-foreground mt-1 truncate">
              {user.email}
            </div>
          )}
        </div>
      </aside>

      {/* Mobile menu */}
      {mobileMenuOpen && (
        <div className="md:hidden fixed inset-0 z-50 bg-background">
          <div className="flex flex-col h-full">
            <div className="flex items-center justify-between p-4 border-b">
              <h2 className="text-lg font-bold">Menu</h2>
              <button
                onClick={() => setMobileMenuOpen(false)}
                className="p-2 rounded-md hover:bg-muted"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <nav className="flex-1 p-4 space-y-1">
              {navigation.map((item) => (
                <button
                  key={item.id}
                  onClick={() => {
                    handleNavigation(item.href);
                    setMobileMenuOpen(false);
                  }}
                  className={cn(
                    "w-full flex items-center gap-3 px-4 py-3 rounded-md text-sm font-medium",
                    activePath === item.id
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:bg-muted/50"
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  {item.title}
                </button>
              ))}
            </nav>
          </div>
        </div>
      )}

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
