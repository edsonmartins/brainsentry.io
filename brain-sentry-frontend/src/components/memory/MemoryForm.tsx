import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X } from "lucide-react";
import type { MemoryCategory, ImportanceLevel } from "@/lib/api";

export interface MemoryFormData {
  content: string;
  summary: string;
  category: MemoryCategory;
  importance: ImportanceLevel;
  tags: string[];
}

interface MemoryFormProps {
  initialData?: Partial<MemoryFormData>;
  onSubmit: (data: MemoryFormData) => void;
  onCancel?: () => void;
  isSubmitting?: boolean;
  title?: string;
  inline?: boolean; // Se true, não renderiza o Card wrapper
}

const categoryDefs: { value: MemoryCategory; color: string }[] = [
  { value: "INSIGHT", color: "bg-blue-500" },
  { value: "DECISION", color: "bg-purple-500" },
  { value: "WARNING", color: "bg-red-500" },
  { value: "KNOWLEDGE", color: "bg-indigo-500" },
  { value: "ACTION", color: "bg-green-500" },
  { value: "CONTEXT", color: "bg-cyan-500" },
  { value: "REFERENCE", color: "bg-orange-500" },
];

const importanceDefs: { value: ImportanceLevel; color: string }[] = [
  { value: "CRITICAL", color: "bg-red-500" },
  { value: "IMPORTANT", color: "bg-orange-500" },
  { value: "MINOR", color: "bg-gray-500" },
];

export function MemoryForm({
  initialData,
  onSubmit,
  onCancel,
  isSubmitting = false,
  title,
  inline = false,
}: MemoryFormProps) {
  const { t } = useTranslation();
  const categories = categoryDefs.map((c) => ({
    ...c,
    label: t(`memory.categories.${c.value}`),
  }));
  const importanceLevels = importanceDefs.map((l) => ({
    ...l,
    label: t(`memory.importanceLevels.${l.value}`),
  }));
  const [formData, setFormData] = useState<MemoryFormData>({
    content: initialData?.content || "",
    summary: initialData?.summary || "",
    category: initialData?.category || "INSIGHT",
    importance: initialData?.importance || "IMPORTANT",
    tags: initialData?.tags || [],
  });

  const [tagInput, setTagInput] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  const addTag = () => {
    const trimmed = tagInput.trim().toLowerCase();
    if (trimmed && !formData.tags.includes(trimmed)) {
      setFormData((prev) => ({ ...prev, tags: [...prev.tags, trimmed] }));
      setTagInput("");
    }
  };

  const removeTag = (tag: string) => {
    setFormData((prev) => ({ ...prev, tags: prev.tags.filter((t) => t !== tag) }));
  };

  const handleTagKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      addTag();
    }
  };

  const formContent = (
    <form onSubmit={handleSubmit} className="space-y-6">
          {/* Content */}
          <div className="space-y-2">
            <label htmlFor="content" className="text-sm font-medium">
              {t("memory.content")} <span className="text-destructive">*</span>
            </label>
            <textarea
              id="content"
              required
              rows={4}
              value={formData.content}
              onChange={(e) => setFormData((prev) => ({ ...prev, content: e.target.value }))}
              placeholder={t("memory.contentPlaceholder")}
              className="w-full px-3 py-2 text-sm bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring resize-none"
            />
          </div>

          {/* Summary */}
          <div className="space-y-2">
            <label htmlFor="summary" className="text-sm font-medium">
              {t("memory.summary")} <span className="text-destructive">*</span>
            </label>
            <input
              id="summary"
              type="text"
              required
              value={formData.summary}
              onChange={(e) => setFormData((prev) => ({ ...prev, summary: e.target.value }))}
              placeholder={t("memory.summaryPlaceholder")}
              className="w-full px-3 py-2 text-sm bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          {/* Category */}
          <div className="space-y-2">
            <label className="text-sm font-medium">{t("memory.category")}</label>
            <div className="flex flex-wrap gap-2">
              {categories.map((cat) => (
                <button
                  key={cat.value}
                  type="button"
                  onClick={() => setFormData((prev) => ({ ...prev, category: cat.value as any }))}
                  className={`
                    px-3 py-1.5 text-sm font-medium rounded-full transition-colors
                    ${formData.category === cat.value
                      ? `${cat.color} text-white`
                      : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
                    }
                  `}
                >
                  {cat.label}
                </button>
              ))}
            </div>
          </div>

          {/* Importance */}
          <div className="space-y-2">
            <label className="text-sm font-medium">{t("memory.importance")}</label>
            <div className="flex flex-wrap gap-2">
              {importanceLevels.map((level) => (
                <button
                  key={level.value}
                  type="button"
                  onClick={() => setFormData((prev) => ({ ...prev, importance: level.value as any }))}
                  className={`
                    px-3 py-1.5 text-sm font-medium rounded-full transition-colors
                    ${formData.importance === level.value
                      ? `${level.color} text-white`
                      : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
                    }
                  `}
                >
                  {level.label}
                </button>
              ))}
            </div>
          </div>

          {/* Tags */}
          <div className="space-y-2">
            <label htmlFor="tags" className="text-sm font-medium">
              {t("common.tags")}
            </label>
            <div className="flex gap-2">
              <input
                id="tags"
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
                placeholder={t("memory.tagsPlaceholder")}
                className="flex-1 px-3 py-2 text-sm bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <Button type="button" size="sm" variant="secondary" onClick={addTag}>
                {t("common.add")}
              </Button>
            </div>
            {formData.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-2">
                {formData.tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center gap-1 rounded-full bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => removeTag(tag)}
                      className="hover:text-destructive"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-3 pt-4 border-t">
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
                {t("common.cancel")}
              </Button>
            )}
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("common.saving") : t("common.save")}
            </Button>
          </div>
        </form>
  );

  // Se inline, retorna apenas o form sem o Card wrapper
  if (inline) {
    return formContent;
  }

  // Caso contrário, retorna com o Card wrapper
  return (
    <Card className="max-w-2xl mx-auto">
      {(title || title === undefined) && (
        <CardHeader>
          <CardTitle>{title ?? t("memory.new")}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        {formContent}
      </CardContent>
    </Card>
  );
}
