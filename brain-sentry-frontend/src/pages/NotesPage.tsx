import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { StickyNote, RefreshCw, Eye, Lightbulb } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api";

interface Note {
  id: string;
  title: string;
  content: string;
  noteType: string;
  sessionId: string;
  createdAt: string;
}

interface HindsightNote {
  id: string;
  sessionId: string;
  content: string;
  impact: string;
  createdAt: string;
}

export default function NotesPage() {
  const { t, i18n } = useTranslation();
  const { toast } = useToast();
  const dateLocale = i18n.language === "en" ? "en-US" : "pt-BR";
  const [tab, setTab] = useState<"notes" | "hindsight">("notes");
  const [notes, setNotes] = useState<Note[]>([]);
  const [hindsight, setHindsight] = useState<HindsightNote[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    setLoading(true);
    try {
      if (tab === "notes") {
        const resp = await api.axiosInstance.get<Note[]>("/v1/notes");
        setNotes(Array.isArray(resp.data) ? resp.data : []);
      } else {
        const resp = await api.axiosInstance.get<HindsightNote[]>("/v1/notes/hindsight");
        setHindsight(Array.isArray(resp.data) ? resp.data : []);
      }
    } catch (err: any) {
      toast({ title: t("notes.error"), description: err.message, variant: "error" });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [tab]);

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <StickyNote className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">{t("notes.title")}</h1>
                <p className="text-xs text-white/80">
                  {t("notes.subtitle")}
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="bg-white/20 text-white border-white/30 hover:bg-white/30" onClick={fetchData}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          <Button
            variant={tab === "notes" ? "default" : "outline"}
            size="sm"
            onClick={() => setTab("notes")}
          >
            <Eye className="h-4 w-4 mr-2" />
            {t("notes.tabs.notes")}
          </Button>
          <Button
            variant={tab === "hindsight" ? "default" : "outline"}
            size="sm"
            onClick={() => setTab("hindsight")}
          >
            <Lightbulb className="h-4 w-4 mr-2" />
            {t("notes.tabs.hindsight")}
          </Button>
        </div>

        {loading && (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        )}

        {!loading && tab === "notes" && (
          <div className="space-y-4">
            {notes.length === 0 ? (
              <Card>
                <CardContent className="p-12 text-center">
                  <StickyNote className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
                  <h3 className="text-lg font-semibold mb-2">{t("notes.empty")}</h3>
                  <p className="text-muted-foreground">
                    {t("notes.emptyDesc")}
                  </p>
                </CardContent>
              </Card>
            ) : (
              notes.map((note) => (
                <Card key={note.id}>
                  <CardHeader>
                    <CardTitle className="text-sm">{note.title}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground whitespace-pre-wrap">{note.content}</p>
                    <div className="flex gap-4 mt-3 text-xs text-muted-foreground">
                      <span>{t("notes.type", { type: note.noteType })}</span>
                      <span>{t("notes.session", { id: note.sessionId?.slice(0, 8) || "-" })}</span>
                      <span>{new Date(note.createdAt).toLocaleString(dateLocale)}</span>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        )}

        {!loading && tab === "hindsight" && (
          <div className="space-y-4">
            {hindsight.length === 0 ? (
              <Card>
                <CardContent className="p-12 text-center">
                  <Lightbulb className="h-16 w-16 mx-auto mb-4 text-muted-foreground opacity-50" />
                  <h3 className="text-lg font-semibold mb-2">{t("notes.emptyHindsight")}</h3>
                  <p className="text-muted-foreground">
                    {t("notes.emptyHindsightDesc")}
                  </p>
                </CardContent>
              </Card>
            ) : (
              hindsight.map((note) => (
                <Card key={note.id}>
                  <CardContent className="p-6">
                    <p className="text-sm whitespace-pre-wrap">{note.content}</p>
                    {note.impact && (
                      <div className="mt-2 p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded">
                        <p className="text-sm text-yellow-700 dark:text-yellow-300">
                          {t("notes.impact", { impact: note.impact })}
                        </p>
                      </div>
                    )}
                    <div className="flex gap-4 mt-3 text-xs text-muted-foreground">
                      <span>{t("notes.session", { id: note.sessionId?.slice(0, 8) || "-" })}</span>
                      <span>{new Date(note.createdAt).toLocaleString(dateLocale)}</span>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        )}
      </main>
    </div>
  );
}
