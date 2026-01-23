import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Eye, Edit, Trash } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Memory } from "@/lib/api";

interface MemoryCardProps {
  memory: Memory;
  onView?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}

const importanceStyles: Record<string, string> = {
  CRITICAL: "bg-red-500 text-white hover:bg-red-600",
  IMPORTANT: "bg-orange-500 text-white hover:bg-orange-600",
  MINOR: "bg-gray-500 text-white hover:bg-gray-600",
};

const categoryStyles: Record<string, string> = {
  // New universal categories
  INSIGHT: "bg-blue-500 text-white hover:bg-blue-600",
  DECISION: "bg-purple-500 text-white hover:bg-purple-600",
  WARNING: "bg-red-500 text-white hover:bg-red-600",
  KNOWLEDGE: "bg-indigo-500 text-white hover:bg-indigo-600",
  ACTION: "bg-green-500 text-white hover:bg-green-600",
  CONTEXT: "bg-cyan-500 text-white hover:bg-cyan-600",
  REFERENCE: "bg-orange-500 text-white hover:bg-orange-600",
  // Legacy categories (backward compatibility)
  PATTERN: "bg-blue-500 text-white hover:bg-blue-600",
  ANTIPATTERN: "bg-red-500 text-white hover:bg-red-600",
  DOMAIN: "bg-indigo-500 text-white hover:bg-indigo-600",
  BUG: "bg-yellow-500 text-white hover:bg-yellow-600",
  OPTIMIZATION: "bg-green-500 text-white hover:bg-green-600",
  INTEGRATION: "bg-cyan-500 text-white hover:bg-cyan-600",
};

export function MemoryCard({ memory, onView, onEdit, onDelete }: MemoryCardProps) {
  const importanceClass = importanceStyles[memory.importance] || importanceStyles.MINOR;
  const categoryClass = categoryStyles[memory.category] || "bg-gray-500 text-white";

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div className="space-y-1 flex-1 min-w-0">
            <CardTitle className="text-base leading-tight truncate">{memory.summary}</CardTitle>
            <CardDescription className="text-xs line-clamp-2">
              {memory.content}
            </CardDescription>
          </div>
          <div className="flex flex-col gap-1 shrink-0">
            <span className={cn("inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold", importanceClass)}>
              {memory.importance}
            </span>
            <span className={cn("inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold", categoryClass)}>
              {memory.category}
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {/* Tags */}
          {memory.tags && memory.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {memory.tags.map((tag) => (
                <span
                  key={tag}
                  className="inline-flex items-center rounded-md bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Metrics */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Eye className="h-3 w-3" />
              {memory.accessCount || 0}
            </span>
            <span>•</span>
            <span>{memory.injectionCount || 0} inj</span>
            <span>•</span>
            <span className="text-green-600">{memory.helpfulCount || 0} helpful</span>
          </div>

          <div className="text-xs text-muted-foreground">
            {memory.createdAt ? new Date(memory.createdAt).toLocaleDateString("pt-BR") : "-"}
          </div>

          {/* Actions */}
          <div className="flex items-center gap-1 pt-2 border-t">
            {onView && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 flex-1"
                onClick={() => onView(memory.id)}
              >
                <Eye className="h-3 w-3 mr-1" />
                Ver
              </Button>
            )}
            {onEdit && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 flex-1"
                onClick={() => onEdit(memory.id)}
              >
                <Edit className="h-3 w-3 mr-1" />
                Editar
              </Button>
            )}
            {onDelete && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 flex-1 text-destructive hover:text-destructive"
                onClick={() => onDelete(memory.id)}
              >
                <Trash className="h-3 w-3 mr-1" />
                Excluir
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
