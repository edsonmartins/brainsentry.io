import { useState } from "react";
import {
  Settings,
  Save,
  RotateCcw,
  Bell,
  Shield,
  Database,
  Palette,
  Search,
  Sparkles,
  ChevronRight,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/filter";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

interface ConfigSection {
  id: string;
  title: string;
  description: string;
  icon: React.ElementType;
}

interface ConfigOption {
  key: string;
  label: string;
  type: "text" | "number" | "boolean" | "select";
  value: string | number | boolean;
  options?: { value: string; label: string }[];
  description?: string;
}

const CONFIG_SECTIONS: ConfigSection[] = [
  {
    id: "general",
    title: "Geral",
    description: "Configurações gerais do sistema",
    icon: Settings,
  },
  {
    id: "notifications",
    title: "Notificações",
    description: "Preferências de alertas e notificações",
    icon: Bell,
  },
  {
    id: "security",
    title: "Segurança",
    description: "Configurações de segurança e acesso",
    icon: Shield,
  },
  {
    id: "embeddings",
    title: "Embeddings",
    description: "Configurações do modelo de embeddings",
    icon: Sparkles,
  },
  {
    id: "database",
    title: "Banco de Dados",
    description: "Configurações de persistência",
    icon: Database,
  },
];

const DEFAULT_CONFIGS: Record<string, ConfigOption[]> = {
  general: [
    {
      key: "appName",
      label: "Nome da Aplicação",
      type: "text",
      value: "Brain Sentry",
      description: "Nome exibido na interface",
    },
    {
      key: "sessionTimeout",
      label: "Timeout de Sessão (minutos)",
      type: "number",
      value: 30,
      description: "Tempo limite de inatividade",
    },
    {
      key: "defaultPageSize",
      label: "Itens por Página",
      type: "select",
      value: 20,
      options: [
        { value: "10", label: "10" },
        { value: "20", label: "20" },
        { value: "50", label: "50" },
        { value: "100", label: "100" },
      ],
    },
  ],
  notifications: [
    {
      key: "enabled",
      label: "Habilitar Notificações",
      type: "boolean",
      value: true,
    },
    {
      key: "emailAlerts",
      label: "Alertas por Email",
      type: "boolean",
      value: false,
    },
    {
      key: "errorNotifications",
      label: "Notificar Erros",
      type: "boolean",
      value: true,
    },
  ],
  security: [
    {
      key: "maxLoginAttempts",
      label: "Tentativas Máximas de Login",
      type: "number",
      value: 5,
    },
    {
      key: "passwordMinLength",
      label: "Comprimento Mínimo de Senha",
      type: "number",
      value: 8,
    },
    {
      key: "requireMFA",
      label: "Exigir MFA",
      type: "boolean",
      value: false,
    },
  ],
  embeddings: [
    {
      key: "model",
      label: "Modelo de Embeddings",
      type: "select",
      value: "all-MiniLM-L6-v2",
      options: [
        { value: "all-MiniLM-L6-v2", label: "all-MiniLM-L6-v2 (Rápido)" },
        { value: "all-mpnet-base-v2", label: "all-mpnet-base-v2 (Preciso)" },
      ],
    },
    {
      key: "dimension",
      label: "Dimensão",
      type: "number",
      value: 384,
      description: "Dimensão do vetor de embeddings",
    },
  ],
  database: [
    {
      key: "connectionTimeout",
      label: "Timeout de Conexão (segundos)",
      type: "number",
      value: 30,
    },
    {
      key: "poolSize",
      label: "Tamanho do Pool",
      type: "number",
      value: 10,
    },
    {
      key: "enableCache",
      label: "Habilitar Cache",
      type: "boolean",
      value: true,
    },
  ],
};

export function ConfigurationPage() {
  const { user } = useAuth();
  const { toast } = useToast();
  const [activeSection, setActiveSection] = useState("general");
  const [configs, setConfigs] = useState<Record<string, ConfigOption[]>>(DEFAULT_CONFIGS);
  const [isSaving, setIsSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  const currentSection = CONFIG_SECTIONS.find((s) => s.id === activeSection);
  const currentOptions = configs[activeSection] || [];

  const handleValueChange = (key: string, value: string | number | boolean) => {
    setConfigs((prev) => ({
      ...prev,
      [activeSection]: prev[activeSection].map((opt) =>
        opt.key === key ? { ...opt, value } : opt
      ),
    }));
    setHasChanges(true);
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));

      toast({
        title: "Configurações salvas",
        description: "As configurações foram atualizadas com sucesso.",
        variant: "success",
      });

      setHasChanges(false);
    } catch {
      toast({
        title: "Erro",
        description: "Não foi possível salvar as configurações.",
        variant: "error",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    setConfigs(DEFAULT_CONFIGS);
    setHasChanges(false);
    toast({
      title: "Configurações resetadas",
      description: "As configurações foram restauradas aos valores padrão.",
      variant: "info",
    });
  };

  const renderConfigInput = (option: ConfigOption) => {
    const baseId = `config-${activeSection}-${option.key}`;

    switch (option.type) {
      case "boolean":
        return (
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              id={baseId}
              type="checkbox"
              checked={option.value as boolean}
              onChange={(e) => handleValueChange(option.key, e.target.checked)}
              className="w-5 h-5 rounded border-gray-300 text-primary focus:ring-primary"
            />
            <span className="text-sm">
              {option.description || option.label}
            </span>
          </label>
        );

      case "select":
        return (
          <div className="space-y-1">
            <label htmlFor={baseId} className="text-sm font-medium">
              {option.label}
            </label>
            <select
              id={baseId}
              value={option.value as string}
              onChange={(e) => handleValueChange(option.key, e.target.value)}
              className="w-full h-9 rounded-md border border-input bg-background px-3 py-1"
            >
              {option.options?.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            {option.description && (
              <p className="text-xs text-muted-foreground">{option.description}</p>
            )}
          </div>
        );

      case "number":
        return (
          <div className="space-y-1">
            <label htmlFor={baseId} className="text-sm font-medium">
              {option.label}
            </label>
            <Input
              id={baseId}
              type="number"
              value={option.value as number}
              onChange={(e) => handleValueChange(option.key, Number(e.target.value))}
              className="w-full"
            />
            {option.description && (
              <p className="text-xs text-muted-foreground">{option.description}</p>
            )}
          </div>
        );

      default:
        return (
          <div className="space-y-1">
            <label htmlFor={baseId} className="text-sm font-medium">
              {option.label}
            </label>
            <Input
              id={baseId}
              type="text"
              value={option.value as string}
              onChange={(e) => handleValueChange(option.key, e.target.value)}
              className="w-full"
            />
            {option.description && (
              <p className="text-xs text-muted-foreground">{option.description}</p>
            )}
          </div>
        );
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white -mx-0">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Settings className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">Configurações</h1>
                <p className="text-xs text-white/80">
                  Gerencie as configurações do sistema
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              {hasChanges && (
                <Button variant="outline" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={handleReset} disabled={isSaving}>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Resetar
                </Button>
              )}
              <Button className="bg-white text-brain-primary hover:bg-white/90" onClick={handleSave} disabled={!hasChanges || isSaving}>
                {isSaving ? (
                  <Spinner size="sm" />
                ) : (
                  <Save className="h-4 w-4 mr-2" />
                )}
                Salvar Alterações
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="flex gap-6">
          {/* Sidebar Navigation */}
          <aside className="w-64 flex-shrink-0">
            <nav className="space-y-1">
              {CONFIG_SECTIONS.map((section) => {
                const Icon = section.icon;
                const isActive = activeSection === section.id;

                return (
                  <button
                    key={section.id}
                    onClick={() => setActiveSection(section.id)}
                    className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-left transition-colors ${
                      isActive
                        ? "bg-gradient-to-r from-brain-primary to-brain-accent text-white shadow-md"
                        : "hover:bg-accent"
                    }`}
                  >
                    <Icon className="h-5 w-5" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">{section.title}</p>
                      <p className={`text-xs ${isActive ? "text-white/70" : "text-muted-foreground"}`}>
                        {section.description}
                      </p>
                    </div>
                    {isActive && <ChevronRight className="h-4 w-4" />}
                  </button>
                );
              })}
            </nav>
          </aside>

          {/* Configuration Panel */}
          <div className="flex-1">
            <Card>
              <CardHeader>
                <div className="flex items-center gap-3">
                  {currentSection && (
                    <div className="p-2 bg-gradient-to-br from-brain-primary to-brain-accent rounded-lg">
                      <currentSection.icon className="h-5 w-5 text-white" />
                    </div>
                  )}
                  <div>
                    <CardTitle>{currentSection?.title}</CardTitle>
                    <CardDescription>{currentSection?.description}</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  {currentOptions.map((option) => (
                    <div key={option.key} className="pb-4 border-b last:border-0">
                      {renderConfigInput(option)}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Info Card */}
            <Card className="mt-6">
              <CardHeader>
                <CardTitle className="text-base">Informações do Sistema</CardTitle>
              </CardHeader>
              <CardContent>
                <dl className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <dt className="text-muted-foreground">Versão</dt>
                    <dd className="font-medium">1.0.0</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Ambiente</dt>
                    <dd className="font-medium">
                      {import.meta.env.MODE === "production" ? "Produção" : "Desenvolvimento"}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">API URL</dt>
                    <dd className="font-mono text-xs">{API_URL}</dd>
                  </div>
                  <div>
                    <dt className="text-muted-foreground">Tenant</dt>
                    <dd className="font-medium">{user?.tenantId || "default"}</dd>
                  </div>
                </dl>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}
