import { useState } from "react";
import { Brain, LogIn, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Input } from "@/components/ui/filter";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/contexts/AuthContext";
import { useToast } from "@/components/ui/toast";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export function LoginPage() {
  const { login } = useAuth();
  const { toast } = useToast();
  const navigate = (path: string) => {
    window.location.href = path;
  };

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      await login(email, password);
      toast({
        title: "Login realizado",
        description: "Bem-vindo de volta!",
        variant: "success",
      });
      navigate("/dashboard");
    } catch (err) {
      const errorMessage = (err as Error).message || "Credenciais inválidas";
      setError(errorMessage);
      toast({
        title: "Erro no login",
        description: errorMessage,
        variant: "error",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDemoLogin = async () => {
    setEmail("demo@example.com");
    setPassword("demo123");
    setError("");
    setIsLoading(true);

    try {
      // Create or get demo user
      const response = await fetch(`${API_URL}/v1/auth/demo`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) {
        throw new Error("Demo login not available");
      }

      const data = await response.json();
      localStorage.setItem("brain_sentry_token", data.token);
      localStorage.setItem("brain_sentry_user", JSON.stringify(data.user));

      toast({
        title: "Demo login realizado",
        description: "Você está usando a conta de demonstração.",
        variant: "success",
      });
      navigate("/dashboard");
    } catch (err) {
      setError("Modo demo não disponível");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo/Brand */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center p-3 bg-primary/10 rounded-full mb-4">
            <Brain className="h-10 w-10 text-primary" />
          </div>
          <h1 className="text-3xl font-bold">Brain Sentry</h1>
          <p className="text-muted-foreground">
            Sistema de Memória para Desenvolvedores
          </p>
        </div>

        {/* Login Card */}
        <Card>
          <CardHeader>
            <CardTitle>Entrar</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Error Message */}
              {error && (
                <div className="flex items-center gap-2 p-3 bg-destructive/10 text-destructive rounded-md">
                  <AlertCircle className="h-4 w-4" />
                  <span className="text-sm">{error}</span>
                </div>
              )}

              {/* Email */}
              <div className="space-y-2">
                <label htmlFor="email" className="text-sm font-medium">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="seu@email.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              {/* Password */}
              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium">
                  Senha
                </label>
                <Input
                  id="password"
                  type="password"
                  placeholder="•••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              {/* Forgot Password */}
              <div className="text-right">
                <a
                  href="#"
                  className="text-sm text-primary hover:underline"
                  onClick={(e) => {
                    e.preventDefault();
                    toast({
                      title: "Recuperação de senha",
                      description: "Entre em contato com o administrador.",
                      variant: "info",
                    });
                  }}
                >
                  Esqueceu sua senha?
                </a>
              </div>

              {/* Submit Button */}
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Spinner size="sm" label="Entrando..." />
                  </>
                ) : (
                  <>
                    <LogIn className="h-4 w-4 mr-2" />
                    Entrar
                  </>
                )}
              </Button>

              {/* Divider */}
              <div className="relative my-6">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-background px-2 text-muted-foreground">
                    Ou
                  </span>
                </div>
              </div>

              {/* Demo Login */}
              <Button
                type="button"
                variant="outline"
                className="w-full"
                onClick={handleDemoLogin}
                disabled={isLoading}
              >
                Entrar com conta demo
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Footer */}
        <p className="text-center text-sm text-muted-foreground mt-6">
          Não tem uma conta?{" "}
          <a
            href="#"
            className="text-primary hover:underline"
            onClick={(e) => {
              e.preventDefault();
              toast({
                title: "Registro",
                description: "Entre em contato com o administrador para criar uma conta.",
                variant: "info",
              });
            }}
          >
            Solicite acesso
          </a>
        </p>

        {/* Tenant Info */}
        <div className="mt-8 p-4 bg-muted/20 rounded-lg">
          <p className="text-xs text-muted-foreground text-center">
            Para ambientes de desenvolvimento, use: demo@example.com / demo123
          </p>
        </div>
      </div>
    </div>
  );
}
