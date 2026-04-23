import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  CalendarClock, RefreshCw, Save, Trash2, Sparkles, Users,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import {
  api,
  type EventRecord,
  type RecordEventRequest,
  type EventParticipant,
} from "@/lib/api/client";

type TabKey = "register" | "extract";

export default function EventsPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();

  const [tab, setTab] = useState<TabKey>("register");
  const [events, setEvents] = useState<EventRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [form, setForm] = useState<{
    eventType: string;
    title: string;
    description: string;
    occurredAt: string;
    participants: string;
  }>({
    eventType: "",
    title: "",
    description: "",
    occurredAt: new Date().toISOString().slice(0, 16),
    participants: "",
  });

  const [extractContent, setExtractContent] = useState("");
  const [extractedEvents, setExtractedEvents] = useState<EventRecord[] | null>(null);
  const [extracting, setExtracting] = useState(false);

  const [filter, setFilter] = useState({ eventType: "", entityId: "", from: "", to: "" });

  const loadEvents = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | undefined> = { limit: "100" } as any;
      if (filter.eventType) params.eventType = filter.eventType;
      if (filter.entityId) params.entityId = filter.entityId;
      if (filter.from) params.from = new Date(filter.from).toISOString();
      if (filter.to) params.to = new Date(filter.to).toISOString();
      const res = await api.listEvents(params as any);
      setEvents(res.events || []);
    } catch (err: any) {
      toast({ title: "Falha ao carregar eventos", description: err?.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  }, [toast, filter]);

  useEffect(() => {
    loadEvents();
  }, [loadEvents]);

  const submitRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.eventType.trim() || !form.title.trim()) {
      toast({ title: "Tipo e título obrigatórios", variant: "warning" });
      return;
    }
    const participants: EventParticipant[] = form.participants
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean)
      .map((entityId) => ({ entityId }));

    const req: RecordEventRequest = {
      eventType: form.eventType,
      title: form.title,
      description: form.description,
      occurredAt: new Date(form.occurredAt).toISOString(),
      participants,
    };

    setSaving(true);
    try {
      await api.recordEvent(req);
      toast({ title: "Evento registrado", variant: "success" });
      setForm({
        eventType: "",
        title: "",
        description: "",
        occurredAt: new Date().toISOString().slice(0, 16),
        participants: "",
      });
      await loadEvents();
    } catch (err: any) {
      toast({ title: "Erro ao registrar", description: err?.message, variant: "error" });
    } finally {
      setSaving(false);
    }
  };

  const runExtract = async () => {
    if (!extractContent.trim()) return;
    setExtracting(true);
    try {
      const res = await api.extractEvents(extractContent);
      setExtractedEvents(res.events || []);
      toast({ title: `${res.count ?? 0} eventos extraídos`, variant: "success" });
      await loadEvents();
    } catch (err: any) {
      toast({ title: "Erro ao extrair", description: err?.message, variant: "error" });
    } finally {
      setExtracting(false);
    }
  };

  const remove = async (id: string) => {
    if (!confirm("Remover evento?")) return;
    try {
      await api.deleteEvent(id);
      toast({ title: "Evento removido", variant: "success" });
      await loadEvents();
    } catch (err: any) {
      toast({ title: "Erro ao remover", description: err?.message, variant: "error" });
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <CalendarClock className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("nav.events")}</h1>
                <p className="text-xs text-white/80">Eventos bi-temporais e extração</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="bg-white/20 border-white/30 text-white hover:bg-white/30"
              onClick={loadEvents}
              disabled={loading}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 space-y-6">
        <div className="inline-flex rounded-lg bg-muted p-0.5">
          {(
            [
              { k: "register", label: "Registrar" },
              { k: "extract", label: "Extrair" },
            ] as const
          ).map((it) => (
            <button
              key={it.k}
              onClick={() => setTab(it.k)}
              className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                tab === it.k
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {it.label}
            </button>
          ))}
        </div>

        {tab === "register" ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Registrar evento</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={submitRegister} className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                <Field label="Tipo">
                  <input
                    value={form.eventType}
                    onChange={(e) => setForm((f) => ({ ...f, eventType: e.target.value }))}
                    className="w-full bg-transparent border rounded px-2 py-1"
                    placeholder="meeting, deploy, incident..."
                  />
                </Field>
                <Field label="Título">
                  <input
                    value={form.title}
                    onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                    className="w-full bg-transparent border rounded px-2 py-1"
                  />
                </Field>
                <div className="md:col-span-2">
                  <Field label="Descrição">
                    <textarea
                      value={form.description}
                      onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                      rows={3}
                      className="w-full bg-transparent border rounded px-2 py-1"
                    />
                  </Field>
                </div>
                <Field label="Ocorrido em">
                  <input
                    type="datetime-local"
                    value={form.occurredAt}
                    onChange={(e) => setForm((f) => ({ ...f, occurredAt: e.target.value }))}
                    className="w-full bg-transparent border rounded px-2 py-1"
                  />
                </Field>
                <Field label="Participantes (entityIds separados por vírgula)">
                  <input
                    value={form.participants}
                    onChange={(e) => setForm((f) => ({ ...f, participants: e.target.value }))}
                    className="w-full bg-transparent border rounded px-2 py-1 font-mono text-xs"
                    placeholder="uuid1, uuid2"
                  />
                </Field>
                <div className="md:col-span-2">
                  <Button
                    type="submit"
                    size="sm"
                    disabled={saving}
                    className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
                  >
                    {saving ? <Spinner size="sm" /> : <Save className="h-4 w-4 mr-1" />}
                    Registrar
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm flex items-center gap-2">
                <Sparkles className="h-4 w-4" /> Extrair eventos de texto
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <textarea
                value={extractContent}
                onChange={(e) => setExtractContent(e.target.value)}
                rows={6}
                className="w-full bg-transparent border rounded px-2 py-1"
                placeholder="Cole texto com descrição de eventos (reuniões, deploys, incidentes)..."
              />
              <Button size="sm" onClick={runExtract} disabled={extracting || !extractContent.trim()}>
                {extracting ? <Spinner size="sm" /> : <Sparkles className="h-4 w-4 mr-1" />}
                Extrair
              </Button>
              {extractedEvents && (
                <div>
                  <p className="text-xs text-muted-foreground mb-2">
                    Eventos extraídos: {extractedEvents.length}
                  </p>
                  <ul className="space-y-2">
                    {extractedEvents.map((ev) => (
                      <li key={ev.id} className="border rounded p-2 text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-medium">{ev.title}</span>
                          <span className="px-1.5 py-0.5 rounded bg-muted text-[10px]">{ev.eventType}</span>
                        </div>
                        <p className="text-muted-foreground mt-0.5">{ev.description}</p>
                        <p className="text-[10px] text-muted-foreground mt-1">
                          {new Date(ev.occurredAt).toLocaleString(i18n.language)}
                        </p>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Filtrar</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-2 text-sm">
              <input
                value={filter.eventType}
                onChange={(e) => setFilter((f) => ({ ...f, eventType: e.target.value }))}
                placeholder="tipo"
                className="bg-transparent border rounded px-2 py-1"
              />
              <input
                value={filter.entityId}
                onChange={(e) => setFilter((f) => ({ ...f, entityId: e.target.value }))}
                placeholder="entityId"
                className="bg-transparent border rounded px-2 py-1 font-mono text-xs"
              />
              <input
                type="datetime-local"
                value={filter.from}
                onChange={(e) => setFilter((f) => ({ ...f, from: e.target.value }))}
                className="bg-transparent border rounded px-2 py-1"
              />
              <input
                type="datetime-local"
                value={filter.to}
                onChange={(e) => setFilter((f) => ({ ...f, to: e.target.value }))}
                className="bg-transparent border rounded px-2 py-1"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Eventos ({events.length})</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {loading ? (
              <div className="flex justify-center py-10"><Spinner size="lg" /></div>
            ) : events.length === 0 ? (
              <EmptyState icon={CalendarClock} title="Nenhum evento" description="Registre ou extraia eventos acima." />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead className="border-b bg-muted/40">
                    <tr>
                      <th className="text-left p-2">Tipo</th>
                      <th className="text-left p-2">Título</th>
                      <th className="text-left p-2">Ocorrido em</th>
                      <th className="text-left p-2">Participantes</th>
                      <th className="text-left p-2"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {events.map((ev) => (
                      <tr key={ev.id} className="border-b hover:bg-muted/30">
                        <td className="p-2 font-mono">{ev.eventType}</td>
                        <td className="p-2 max-w-[280px] truncate" title={ev.title}>{ev.title}</td>
                        <td className="p-2 text-muted-foreground">
                          {new Date(ev.occurredAt).toLocaleString(i18n.language)}
                        </td>
                        <td className="p-2">
                          <span className="inline-flex items-center gap-1">
                            <Users className="h-3 w-3" />
                            {ev.participants?.length ?? 0}
                          </span>
                        </td>
                        <td className="p-2">
                          <Button size="sm" variant="ghost" className="h-7 w-7 p-0 text-destructive" onClick={() => remove(ev.id)}>
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</span>
      <div className="mt-0.5">{children}</div>
    </label>
  );
}
