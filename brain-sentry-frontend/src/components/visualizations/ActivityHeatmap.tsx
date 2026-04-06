import { useState, useEffect, useMemo } from "react";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { HeatmapChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  VisualMapComponent,
  CalendarComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import { Spinner } from "@/components/ui/spinner";
import { api } from "@/lib/api/client";

echarts.use([
  HeatmapChart,
  GridComponent,
  TooltipComponent,
  VisualMapComponent,
  CalendarComponent,
  CanvasRenderer,
]);

interface ActivityHeatmapProps {
  className?: string;
}

function getDateRange(): [string, string] {
  const end = new Date();
  const start = new Date();
  start.setFullYear(start.getFullYear() - 1);
  return [
    start.toISOString().split("T")[0],
    end.toISOString().split("T")[0],
  ];
}

function formatDate(date: Date): string {
  return date.toISOString().split("T")[0];
}

export function ActivityHeatmap({ className = "" }: ActivityHeatmapProps) {
  const [loading, setLoading] = useState(true);
  const [activityData, setActivityData] = useState<[string, number][]>([]);
  const [totalEvents, setTotalEvents] = useState(0);

  useEffect(() => {
    async function fetchActivity() {
      setLoading(true);
      try {
        // Fetch memories to build activity data from creation dates
        const [memoriesData, auditData] = await Promise.all([
          api.getMemories(0, 500).catch(() => ({ memories: [] })),
          api.getAuditLogs(500).catch(() => []),
        ]);

        // Count events per day
        const dayCounts: Record<string, number> = {};

        // Count memory creations
        for (const mem of memoriesData.memories || []) {
          if (mem.createdAt) {
            const day = new Date(mem.createdAt).toISOString().split("T")[0];
            dayCounts[day] = (dayCounts[day] || 0) + 1;
          }
        }

        // Count audit events
        const auditList = Array.isArray(auditData) ? auditData : auditData?.logs || [];
        for (const entry of auditList) {
          const ts = entry.timestamp || entry.createdAt;
          if (ts) {
            const day = new Date(ts).toISOString().split("T")[0];
            dayCounts[day] = (dayCounts[day] || 0) + 1;
          }
        }

        const data: [string, number][] = Object.entries(dayCounts).map(([day, count]) => [day, count]);
        setActivityData(data);
        setTotalEvents(data.reduce((sum, [, count]) => sum + count, 0));
      } catch (err) {
        console.error("Failed to load activity data:", err);
      } finally {
        setLoading(false);
      }
    }

    fetchActivity();
  }, []);

  const [rangeStart, rangeEnd] = useMemo(() => getDateRange(), []);

  const maxValue = useMemo(() => {
    if (activityData.length === 0) return 1;
    return Math.max(...activityData.map(([, v]) => v));
  }, [activityData]);

  const option = useMemo(
    () => ({
      tooltip: {
        formatter: (params: any) => {
          const date = params.value[0];
          const count = params.value[1];
          return `<strong>${date}</strong><br/>${count} event${count !== 1 ? "s" : ""}`;
        },
        backgroundColor: "#1f2937",
        borderColor: "#374151",
        textStyle: { color: "#e5e7eb" },
      },
      visualMap: {
        min: 0,
        max: Math.max(maxValue, 4),
        type: "piecewise",
        orient: "horizontal",
        left: "center",
        bottom: 0,
        pieces: [
          { min: 0, max: 0, label: "0", color: "#161b22" },
          { min: 1, max: Math.ceil(maxValue * 0.25) || 1, label: "Low", color: "#0e4429" },
          { min: Math.ceil(maxValue * 0.25) + 1, max: Math.ceil(maxValue * 0.5) || 2, label: "Med", color: "#006d32" },
          { min: Math.ceil(maxValue * 0.5) + 1, max: Math.ceil(maxValue * 0.75) || 3, label: "High", color: "#26a641" },
          { min: Math.ceil(maxValue * 0.75) + 1, max: maxValue || 4, label: "Max", color: "#39d353" },
        ],
        textStyle: { color: "#9ca3af", fontSize: 10 },
      },
      calendar: {
        range: [rangeStart, rangeEnd],
        cellSize: [13, 13],
        top: 30,
        left: 40,
        right: 10,
        itemStyle: {
          borderWidth: 2,
          borderColor: "#0d1117",
          color: "#161b22",
        },
        yearLabel: { show: false },
        monthLabel: {
          color: "#9ca3af",
          fontSize: 10,
          nameMap: "en",
        },
        dayLabel: {
          firstDay: 0,
          color: "#9ca3af",
          fontSize: 10,
          nameMap: ["", "Mon", "", "Wed", "", "Fri", ""],
        },
        splitLine: { show: false },
      },
      series: [
        {
          type: "heatmap",
          coordinateSystem: "calendar",
          data: activityData,
        },
      ],
    }),
    [activityData, rangeStart, rangeEnd, maxValue]
  );

  if (loading) {
    return (
      <div className={`flex items-center justify-center h-48 ${className}`}>
        <Spinner size="sm" />
      </div>
    );
  }

  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-medium text-muted-foreground">
          Activity ({totalEvents} events this year)
        </h3>
      </div>
      <ReactEChartsCore
        echarts={echarts}
        option={option}
        style={{ height: "180px", width: "100%" }}
        opts={{ renderer: "canvas" }}
      />
    </div>
  );
}
