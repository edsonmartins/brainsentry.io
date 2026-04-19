import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { MemoryForm, type MemoryFormData } from "./MemoryForm";
import { MemoryInsights } from "./MemoryInsights";
import { api, type Memory, type MemoryCategory, type ImportanceLevel, type CreateMemoryRequest, type UpdateMemoryRequest } from "@/lib/api";

interface MemoryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  memory?: Memory;
  onSuccess?: () => void;
}

export function MemoryDialog({ open, onOpenChange, memory, onSuccess }: MemoryDialogProps) {
  const { t } = useTranslation();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!memory;
  const title = isEditing ? t("memory.edit") : t("memory.new");

  const initialData: Partial<MemoryFormData> = memory
    ? {
        content: memory.content,
        summary: memory.summary,
        category: memory.category as MemoryCategory,
        importance: memory.importance as ImportanceLevel,
        tags: memory.tags || [],
      }
    : {
        category: "INSIGHT",
        importance: "IMPORTANT",
        tags: [],
      };

  const handleSubmit = async (data: MemoryFormData) => {
    setIsSubmitting(true);
    setError(null);

    try {
      // tenantId é adicionado automaticamente pelo interceptor
      if (isEditing && memory) {
        // Editar memória existente
        const updateData: UpdateMemoryRequest = data;
        await api.updateMemory(memory.id, updateData);
      } else {
        // Criar nova memória
        const createData: CreateMemoryRequest = data;
        await api.createMemory(createData);
      }

      // Fechar o modal e chamar onSuccess
      onOpenChange(false);
      onSuccess?.();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t("memory.saveError");
      setError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    if (!isSubmitting) {
      onOpenChange(false);
      setError(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl" onClose={handleClose}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>

        {error && (
          <div className="mx-6 mt-4 p-3 bg-destructive/10 text-destructive rounded-md text-sm">
            {error}
          </div>
        )}

        <div className="p-6 pt-0">
          <MemoryForm
            initialData={initialData}
            onSubmit={handleSubmit}
            isSubmitting={isSubmitting}
            inline
          />

          {/* Insights panel (only when editing — we need an existing ID) */}
          {isEditing && memory && (
            <div className="mt-6 pt-4 border-t">
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                {t("memory.insights")}
              </h4>
              <MemoryInsights memoryId={memory.id} />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
