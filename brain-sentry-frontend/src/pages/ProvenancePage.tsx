import { useState } from "react";
import { useTranslation } from "react-i18next";
import { FileCode2, Download, Eye, Info } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api/client";

type Format = "turtle" | "jsonld";

export default function ProvenancePage() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [loadingFormat, setLoadingFormat] = useState<Format | null>(null);
  const [previewFormat, setPreviewFormat] = useState<Format | null>(null);
  const [preview, setPreview] = useState<string | null>(null);

  const download = async (format: Format) => {
    setLoadingFormat(format);
    try {
      const data = await api.exportProvenance(format);
      let blob: Blob;
      let filename: string;
      if (format === "turtle") {
        blob = new Blob([typeof data === "string" ? data : String(data)], {
          type: "text/turtle;charset=utf-8",
        });
        filename = "brainsentry-provenance.ttl";
      } else {
        const jsonStr =
          typeof data === "string" ? data : JSON.stringify(data, null, 2);
        blob = new Blob([jsonStr], { type: "application/ld+json;charset=utf-8" });
        filename = "brainsentry-provenance.jsonld";
      }
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast({ title: "Download iniciado", description: filename, variant: "success" });
    } catch (err: any) {
      toast({ title: "Erro no export", description: err?.message, variant: "error" });
    } finally {
      setLoadingFormat(null);
    }
  };

  const runPreview = async (format: Format) => {
    setPreviewFormat(format);
    setPreview(null);
    try {
      const data = await api.exportProvenance(format);
      if (format === "turtle") {
        setPreview(typeof data === "string" ? data : String(data));
      } else {
        setPreview(
          typeof data === "string"
            ? tryFormatJson(data)
            : JSON.stringify(data, null, 2)
        );
      }
    } catch (err: any) {
      toast({ title: "Erro na pré-visualização", description: err?.message, variant: "error" });
      setPreview(null);
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center gap-3">
            <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
              <FileCode2 className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-base font-bold leading-tight">{t("nav.provenance")}</h1>
              <p className="text-xs text-white/80">Export W3C PROV-O para compliance</p>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6 max-w-5xl space-y-6">
        <Card className="border-dashed">
          <CardContent className="p-4 flex items-start gap-2">
            <Info className="h-4 w-4 text-blue-500 mt-0.5" />
            <div className="text-xs">
              <p className="font-medium">Proveniência W3C PROV-O</p>
              <p className="text-muted-foreground">
                Esta exportação mapeia o audit log e as decisões do BrainSentry sobre o vocabulário{" "}
                <a
                  href="https://www.w3.org/TR/prov-o/"
                  target="_blank"
                  rel="noreferrer"
                  className="underline text-brain-primary"
                >
                  W3C PROV-O
                </a>
                . Use o arquivo resultante para auditoria, compliance regulatório e integração com
                sistemas de linhagem de dados.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Baixar export</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            <Button
              size="sm"
              onClick={() => download("turtle")}
              disabled={loadingFormat === "turtle"}
              className="bg-gradient-to-r from-brain-primary to-brain-accent text-white"
            >
              {loadingFormat === "turtle" ? <Spinner size="sm" /> : <Download className="h-4 w-4 mr-1" />}
              Baixar Turtle (.ttl)
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => download("jsonld")}
              disabled={loadingFormat === "jsonld"}
            >
              {loadingFormat === "jsonld" ? <Spinner size="sm" /> : <Download className="h-4 w-4 mr-1" />}
              Baixar JSON-LD
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm flex items-center gap-2">
              <Eye className="h-4 w-4" /> Pré-visualizar
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => runPreview("turtle")}>
                Turtle
              </Button>
              <Button size="sm" variant="outline" onClick={() => runPreview("jsonld")}>
                JSON-LD
              </Button>
            </div>

            {previewFormat && preview === null && (
              <div className="flex justify-center py-6"><Spinner size="md" /></div>
            )}

            {preview !== null && (
              <div>
                <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
                  {previewFormat}
                </p>
                <pre className="bg-muted p-3 rounded text-[11px] overflow-auto max-h-[480px] whitespace-pre-wrap break-words">
                  {preview}
                </pre>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

function tryFormatJson(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}
