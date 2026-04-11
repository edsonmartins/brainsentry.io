import { useState } from "react";
import { Wand2, Send, MessageSquare, Brain } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/filter";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api";

interface InterceptResult {
  enhanced: boolean;
  originalPrompt: string;
  enhancedPrompt: string;
  contextInjected: string;
  memoriesUsed: Array<{ id: string; summary: string }>;
  notesUsed: Array<{ id: string; title: string }>;
  latencyMs: number;
  reasoning: string;
  confidence: number;
  tokensInjected: number;
  llmCalls: number;
}

interface NLQueryResult {
  question: string;
  cypher: string;
  results: any[];
  attempts: number;
}

export default function PlaygroundPage() {
  const { toast } = useToast();

  // Intercept state
  const [prompt, setPrompt] = useState("");
  const [interceptResult, setInterceptResult] = useState<InterceptResult | null>(null);
  const [interceptLoading, setInterceptLoading] = useState(false);

  // NL Query state
  const [nlQuestion, setNLQuestion] = useState("");
  const [nlResult, setNLResult] = useState<NLQueryResult | null>(null);
  const [nlLoading, setNLLoading] = useState(false);

  const handleIntercept = async () => {
    if (!prompt.trim()) return;
    setInterceptLoading(true);
    try {
      const resp = await api.axiosInstance.post<InterceptResult>("/v1/intercept", {
        prompt,
        sessionId: "playground-" + Date.now(),
      });
      setInterceptResult(resp.data);
    } catch (err: any) {
      toast({ title: "Erro", description: err.message, variant: "error" });
    } finally {
      setInterceptLoading(false);
    }
  };

  const handleNLQuery = async () => {
    if (!nlQuestion.trim()) return;
    setNLLoading(true);
    try {
      const resp = await api.axiosInstance.post<NLQueryResult>("/v1/graph/nl-query", {
        question: nlQuestion,
      });
      setNLResult(resp.data);
    } catch (err: any) {
      toast({ title: "Erro", description: err.message, variant: "error" });
    } finally {
      setNLLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <Wand2 className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">Playground</h1>
              <p className="text-xs text-white/80">
                Teste interception e consultas ao grafo
              </p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 space-y-8">
        {/* Prompt Interception */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Send className="h-5 w-5" />
              Interceptação de Prompt
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder="Digite um prompt para interceptar com memórias..."
                className="flex-1 min-h-[100px] p-3 border rounded-md bg-background resize-y text-sm"
              />
            </div>
            <Button onClick={handleIntercept} disabled={interceptLoading || !prompt.trim()}>
              {interceptLoading ? <Spinner size="sm" /> : <Send className="h-4 w-4 mr-2" />}
              Interceptar
            </Button>

            {interceptResult && (
              <div className="space-y-4 mt-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="p-3 bg-muted rounded-md">
                    <p className="text-xs text-muted-foreground">Confiança</p>
                    <p className="text-lg font-bold">{(interceptResult.confidence * 100).toFixed(0)}%</p>
                  </div>
                  <div className="p-3 bg-muted rounded-md">
                    <p className="text-xs text-muted-foreground">Tokens Injetados</p>
                    <p className="text-lg font-bold">{interceptResult.tokensInjected}</p>
                  </div>
                  <div className="p-3 bg-muted rounded-md">
                    <p className="text-xs text-muted-foreground">Latência</p>
                    <p className="text-lg font-bold">{interceptResult.latencyMs}ms</p>
                  </div>
                </div>

                {interceptResult.enhanced && interceptResult.enhancedPrompt && (
                  <div>
                    <h4 className="text-sm font-medium mb-1">Prompt Enhanced</h4>
                    <pre className="p-3 bg-muted rounded-md text-sm whitespace-pre-wrap overflow-x-auto">
                      {interceptResult.enhancedPrompt}
                    </pre>
                  </div>
                )}

                {interceptResult.contextInjected && (
                  <div>
                    <h4 className="text-sm font-medium mb-1">Contexto Injetado</h4>
                    <pre className="p-3 bg-muted rounded-md text-sm whitespace-pre-wrap overflow-x-auto max-h-64 overflow-y-auto">
                      {interceptResult.contextInjected}
                    </pre>
                  </div>
                )}

                {interceptResult.memoriesUsed.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-1">Memórias Usadas ({interceptResult.memoriesUsed.length})</h4>
                    <ul className="space-y-1">
                      {interceptResult.memoriesUsed.map((m) => (
                        <li key={m.id} className="text-sm text-muted-foreground p-2 bg-muted rounded">
                          <span className="font-mono text-xs">{m.id.slice(0, 8)}</span> — {m.summary}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {interceptResult.reasoning && (
                  <div>
                    <h4 className="text-sm font-medium mb-1">Raciocínio</h4>
                    <p className="text-sm text-muted-foreground">{interceptResult.reasoning}</p>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        {/* NL Graph Query */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MessageSquare className="h-5 w-5" />
              Consulta ao Grafo (Linguagem Natural)
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <Input
                value={nlQuestion}
                onChange={(e) => setNLQuestion(e.target.value)}
                placeholder="Quais memórias estão relacionadas a autenticação?"
                className="flex-1"
                onKeyDown={(e) => e.key === "Enter" && handleNLQuery()}
              />
              <Button onClick={handleNLQuery} disabled={nlLoading || !nlQuestion.trim()}>
                {nlLoading ? <Spinner size="sm" /> : <Brain className="h-4 w-4 mr-2" />}
                Perguntar
              </Button>
            </div>

            {nlResult && (
              <div className="space-y-3 mt-4">
                <div>
                  <h4 className="text-sm font-medium mb-1">Cypher Gerado</h4>
                  <pre className="p-3 bg-muted rounded-md text-sm font-mono overflow-x-auto">
                    {nlResult.cypher}
                  </pre>
                </div>
                <div className="text-xs text-muted-foreground">
                  Tentativas: {nlResult.attempts}
                </div>
                {nlResult.results.length > 0 ? (
                  <div>
                    <h4 className="text-sm font-medium mb-1">Resultados ({nlResult.results.length})</h4>
                    <pre className="p-3 bg-muted rounded-md text-sm whitespace-pre-wrap overflow-x-auto max-h-64 overflow-y-auto">
                      {JSON.stringify(nlResult.results, null, 2)}
                    </pre>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">Nenhum resultado encontrado.</p>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
