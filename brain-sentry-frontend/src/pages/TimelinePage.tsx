import { useState, useEffect, useMemo } from "react";
import {
  Clock, Brain, Search, AlertTriangle, Zap, FileText, Settings,
  ChevronDown, Filter, RefreshCw,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { TypeChips } from "@/components/ui/TypeChips";
import { ImportanceBar } from "@/components/ui/StrengthBar";
import { EmptyState } from "@/components/ui/EmptyState";
import { useToast } from "@/components/ui/toast";
import { api } from "@/lib/api/client";

interface TimelineEvent {
  id: string;
  content: string;
  summary: string;
  category: string;
  importance: string;
  createdAt: string;
  tags: string[];
  memoryType?: string;
}

const CATEGORY_COLORS: Record<string, string> = {
  INSIGHT: "#3b82f6", WARNING: "#ef4444", KNOWLEDGE: "#10b981",
  ACTION: "#f59e0b", CONTEXT: "#8b5cf6", REFERENCE: "#06b6d4",
  DECISION: "#3b82f6", PATTERN: "#10b981", ANTIPATTERN: "#ef4444",
  BUG: "#dc2626", OPTIMIZATION: "#8b5cf6", INTEGRATION: "#06b6d4",
  DOMAIN: "#f59e0b",
};

const CATEGORY_ICONS: Record<string, typeof Brain> = {
  INSIGHT: Zap, WARNING: AlertTriangle, KNOWLEDGE: Brain,
  ACTION: Settings, CONTEXT: FileText, REFERENCE: Search,
  DECISION: Zap, BUG: AlertTriangle,
};

export default function TimelinePage() {
  const { toast } = useToast();
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [minImportance, setMinImportance] = useState<string | null>(null);
  const [visibleCount, setVisibleCount] = useState(30);

  useEffect(() => {
    async function fetchEvents() {
      setLoading(true);
      try {
        const data = await api.getMemories(0, 200);
        const sorted = (data.memories || []).sort(
          (a: TimelineEvent, b: TimelineEvent) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
        setEvents(sorted);
      } catch (err) {
        console.error("Failed to load timeline:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchEvents();
  }, []);

  // Category chip items
  const categoryChips = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const e of events) {
      counts[e.category] = (counts[e.category] || 0) + 1;
    }
    return Object.entries(counts)
      .map(([label, count]) => ({ label, count, color: CATEGORY_COLORS[label] }))
      .sort((a, b) => b.count - a.count);
  }, [events]);

  // Filtered events
  const filtered = useMemo(() => {
    return events.filter((e) => {
      if (selectedCategory && e.category !== selectedCategory) return false;
      if (minImportance) {
        const order = ["MINOR", "IMPORTANT", "CRITICAL"];
        if (order.indexOf(e.importance) < order.indexOf(minImportance)) return false;
      }
      return true;
    });
  }, [events, selectedCategory, minImportance]);

  const visible = filtered.slice(0, visibleCount);

  // Group by date
  const grouped = useMemo(() => {
    const groups: Record<string, TimelineEvent[]> = {};
    for (const e of visible) {
      const date = new Date(e.createdAt).toLocaleDateString("en-US", {
        weekday: "short", year: "numeric", month: "short", day: "numeric",
      });
      if (!groups[date]) groups[date] = [];
      groups[date].push(e);
    }
    return Object.entries(groups);
  }, [visible]);

  const handleRefresh = () => {
    setLoading(true);
    api.getMemories(0, 200).then((data) => {
      setEvents(
        (data.memories || []).sort(
          (a: TimelineEvent, b: TimelineEvent) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        )
      );
      setLoading(false);
      toast({ title: "Timeline updated", variant: "info" });
    });
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-gradient-to-r from-brain-primary to-brain-accent text-white">
        <div className="px-4 py-[14px]">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-1.5 bg-white/20 rounded-lg backdrop-blur-sm">
                <Clock className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-base font-bold leading-tight">Timeline</h1>
                <p className="text-xs text-white/80">{events.length} events</p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="bg-white/20 border-white/30 text-white hover:bg-white/30" onClick={handleRefresh}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        {/* Filters */}
        <div className="mb-6 space-y-3">
          <TypeChips items={categoryChips} selected={selectedCategory} onSelect={setSelectedCategory} />
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">Min Importance:</span>
            {["MINOR", "IMPORTANT", "CRITICAL"].map((level) => (
              <button
                key={level}
                onClick={() => setMinImportance(level === minImportance ? null : level)}
                className={`px-2 py-0.5 text-[10px] rounded-full border transition-colors uppercase tracking-wider ${
                  minImportance === level
                    ? "bg-foreground text-background"
                    : "text-muted-foreground border-border hover:border-foreground/50"
                }`}
              >
                {level}
              </button>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">
            Showing {visible.length} of {filtered.length} events
          </p>
        </div>

        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : events.length === 0 ? (
          <EmptyState
            icon={Clock}
            title="No events yet"
            description="Create memories to see them appear in the timeline"
            action={{ label: "Create Memory", onClick: () => window.location.href = "/app/memories" }}
          />
        ) : (
          <>
            {/* Timeline */}
            <div className="relative">
              {/* Center line */}
              <div className="absolute left-1/2 top-0 bottom-0 w-px bg-border -translate-x-1/2 hidden md:block" />

              {grouped.map(([date, dateEvents]) => (
                <div key={date}>
                  {/* Date header */}
                  <div className="flex justify-center mb-4 mt-6 first:mt-0">
                    <span className="px-3 py-1 text-xs font-medium bg-muted rounded-full border text-muted-foreground relative z-10">
                      {date}
                    </span>
                  </div>

                  {/* Events */}
                  {dateEvents.map((event, idx) => {
                    const isLeft = idx % 2 === 0;
                    const color = CATEGORY_COLORS[event.category] || "#6b7280";
                    const Icon = CATEGORY_ICONS[event.category] || Brain;

                    return (
                      <div key={event.id} className={`flex items-start mb-4 ${isLeft ? "md:flex-row" : "md:flex-row-reverse"} flex-col md:gap-0 gap-2`}>
                        {/* Content */}
                        <div className={`w-full md:w-[calc(50%-24px)] ${isLeft ? "md:pr-4 md:text-right" : "md:pl-4"}`}>
                          <Card className="border-l-[3px] hover:shadow-md transition-shadow" style={{ borderLeftColor: color }}>
                            <CardContent className="p-3">
                              <div className={`flex items-center gap-2 mb-1.5 ${isLeft ? "md:justify-end" : ""}`}>
                                <Icon className="h-3.5 w-3.5 flex-shrink-0" style={{ color }} />
                                <span className="text-[10px] uppercase tracking-wider font-medium" style={{ color }}>
                                  {event.category}
                                </span>
                                <span className="text-[10px] text-muted-foreground">
                                  {new Date(event.createdAt).toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })}
                                </span>
                              </div>

                              <p className={`text-sm leading-relaxed ${isLeft ? "md:text-right" : ""}`}>
                                {event.summary || (event.content.length > 120 ? event.content.slice(0, 117) + "..." : event.content)}
                              </p>

                              <div className={`flex items-center gap-2 mt-2 ${isLeft ? "md:justify-end" : ""}`}>
                                <ImportanceBar importance={event.importance} />
                                {event.tags?.slice(0, 3).map((tag) => (
                                  <span key={tag} className="text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                                    {tag}
                                  </span>
                                ))}
                              </div>
                            </CardContent>
                          </Card>
                        </div>

                        {/* Center dot */}
                        <div className="hidden md:flex items-center justify-center w-12 relative z-10">
                          <div
                            className="h-3 w-3 rounded-full border-2 border-background"
                            style={{ backgroundColor: color }}
                          />
                        </div>

                        {/* Spacer */}
                        <div className="hidden md:block w-[calc(50%-24px)]" />
                      </div>
                    );
                  })}
                </div>
              ))}

              {/* Load More */}
              {visibleCount < filtered.length && (
                <div className="flex justify-center mt-6">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setVisibleCount((c) => c + 30)}
                    className="gap-2"
                  >
                    <ChevronDown className="h-4 w-4" />
                    Load more ({filtered.length - visibleCount} remaining)
                  </Button>
                </div>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
