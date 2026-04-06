import { useState, useEffect, useRef, useCallback } from "react";
import { Search, ZoomIn, ZoomOut, Maximize2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { api } from "@/lib/api/client";
import CytoscapeComponent from "react-cytoscapejs";
import type Cytoscape from "cytoscape";

const NODE_COLORS: Record<string, string> = {
  TECHNOLOGY: "#3b82f6",
  PERSON: "#10b981",
  PROJECT: "#f59e0b",
  CONCEPT: "#8b5cf6",
  LIBRARY: "#06b6d4",
  LANGUAGE: "#ec4899",
  TOOL: "#f97316",
  SERVICE: "#14b8a6",
  FILE: "#22c55e",
  FUNCTION: "#6366f1",
  DEFAULT: "#6b7280",
};

const NODE_SHAPES: Record<string, string> = {
  TECHNOLOGY: "round-rectangle",
  PERSON: "ellipse",
  PROJECT: "diamond",
  CONCEPT: "ellipse",
  LIBRARY: "hexagon",
  LANGUAGE: "round-rectangle",
  TOOL: "round-rectangle",
  SERVICE: "barrel",
  FILE: "rectangle",
  FUNCTION: "ellipse",
};

interface KnowledgeGraphProps {
  limit?: number;
  height?: string;
  className?: string;
}

export function KnowledgeGraph({ limit = 100, height = "600px", className = "" }: KnowledgeGraphProps) {
  const cyRef = useRef<Cytoscape.Core | null>(null);
  const [elements, setElements] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedNode, setSelectedNode] = useState<any | null>(null);
  const [stats, setStats] = useState({ nodes: 0, edges: 0 });

  const fetchGraph = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getKnowledgeGraph(limit);
      const nodes = (data.nodes || []).map((n: any) => ({
        data: {
          id: n.id || n.name,
          label: n.name,
          type: n.type || "DEFAULT",
          properties: n.properties || {},
          color: NODE_COLORS[n.type] || NODE_COLORS.DEFAULT,
          shape: NODE_SHAPES[n.type] || "ellipse",
        },
      }));

      const edges = (data.edges || data.relationships || []).map((e: any, i: number) => ({
        data: {
          id: e.id || `edge-${i}`,
          source: e.source || e.fromNodeId || e.sourceNodeId,
          target: e.target || e.toNodeId || e.targetNodeId,
          label: e.type || "",
          weight: e.weight || e.strength || 0.5,
        },
      }));

      setElements([...nodes, ...edges]);
      setStats({ nodes: nodes.length, edges: edges.length });
    } catch (err: any) {
      setError(err?.message || "Failed to load knowledge graph");
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    fetchGraph();
  }, [fetchGraph]);

  const handleCyInit = useCallback((cy: Cytoscape.Core) => {
    cyRef.current = cy;

    cy.on("tap", "node", (evt) => {
      const node = evt.target;
      setSelectedNode({
        id: node.id(),
        label: node.data("label"),
        type: node.data("type"),
        properties: node.data("properties"),
        degree: node.degree(false),
      });

      // Highlight connected
      cy.elements().removeClass("highlighted dimmed");
      const neighborhood = node.neighborhood().add(node);
      neighborhood.addClass("highlighted");
      cy.elements().not(neighborhood).addClass("dimmed");
    });

    cy.on("tap", (evt) => {
      if (evt.target === cy) {
        setSelectedNode(null);
        cy.elements().removeClass("highlighted dimmed");
      }
    });
  }, []);

  const handleSearch = useCallback(() => {
    if (!cyRef.current || !searchQuery) return;
    const cy = cyRef.current;
    cy.elements().removeClass("highlighted dimmed search-match");

    const matched = cy.nodes().filter((n) =>
      n.data("label").toLowerCase().includes(searchQuery.toLowerCase())
    );

    if (matched.length > 0) {
      matched.addClass("search-match highlighted");
      cy.elements().not(matched.neighborhood().add(matched)).addClass("dimmed");
      cy.animate({ fit: { eles: matched, padding: 50 } }, { duration: 500 });
    }
  }, [searchQuery]);

  const handleZoomIn = () => cyRef.current?.zoom(cyRef.current.zoom() * 1.3);
  const handleZoomOut = () => cyRef.current?.zoom(cyRef.current.zoom() / 1.3);
  const handleFit = () => cyRef.current?.fit(undefined, 50);

  const stylesheet: any[] = [
    {
      selector: "node",
      style: {
        label: "data(label)",
        "background-color": "data(color)",
        shape: "data(shape)" as any,
        width: "mapData(degree, 0, 20, 30, 80)",
        height: "mapData(degree, 0, 20, 30, 80)",
        "font-size": "10px",
        "text-wrap": "ellipsis",
        "text-max-width": "80px",
        "text-valign": "bottom",
        "text-margin-y": 5,
        color: "#e5e7eb",
        "text-outline-width": 2,
        "text-outline-color": "#1f2937",
        "border-width": 2,
        "border-color": "#374151",
      },
    },
    {
      selector: "edge",
      style: {
        width: "mapData(weight, 0, 1, 1, 4)",
        "line-color": "#4b5563",
        "target-arrow-color": "#4b5563",
        "target-arrow-shape": "triangle",
        "curve-style": "bezier",
        label: "data(label)",
        "font-size": "8px",
        color: "#9ca3af",
        "text-rotation": "autorotate",
        opacity: 0.6,
      },
    },
    {
      selector: ".highlighted",
      style: {
        opacity: 1,
        "border-width": 3,
        "border-color": "#f59e0b",
        "z-index": 10,
      },
    },
    {
      selector: ".dimmed",
      style: {
        opacity: 0.15,
      },
    },
    {
      selector: ".search-match",
      style: {
        "border-width": 4,
        "border-color": "#ef4444",
        "background-color": "#fbbf24",
      },
    },
    {
      selector: "edge.highlighted",
      style: {
        opacity: 1,
        width: 3,
        "line-color": "#f59e0b",
        "target-arrow-color": "#f59e0b",
      },
    },
  ];

  if (loading) {
    return (
      <div className={`flex items-center justify-center ${className}`} style={{ height }}>
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className={`flex flex-col items-center justify-center gap-4 ${className}`} style={{ height }}>
        <p className="text-sm text-destructive">{error}</p>
        <Button variant="outline" size="sm" onClick={fetchGraph}>
          <RefreshCw className="h-4 w-4 mr-2" /> Retry
        </Button>
      </div>
    );
  }

  if (elements.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center gap-2 text-muted-foreground ${className}`} style={{ height }}>
        <p className="text-sm">No graph data available</p>
        <p className="text-xs">Create memories to build the knowledge graph</p>
      </div>
    );
  }

  return (
    <div className={`relative ${className}`}>
      {/* Toolbar */}
      <div className="absolute top-3 left-3 z-10 flex items-center gap-2">
        <div className="flex items-center bg-background/90 backdrop-blur-sm border rounded-lg px-2 py-1 gap-1">
          <Search className="h-3.5 w-3.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search nodes..."
            className="bg-transparent text-sm border-none outline-none w-36"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          />
        </div>
        <div className="flex gap-1 bg-background/90 backdrop-blur-sm border rounded-lg p-0.5">
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleZoomIn}>
            <ZoomIn className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleZoomOut}>
            <ZoomOut className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleFit}>
            <Maximize2 className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={fetchGraph}>
            <RefreshCw className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="absolute top-3 right-3 z-10 bg-background/90 backdrop-blur-sm border rounded-lg px-3 py-1.5 text-xs text-muted-foreground">
        {stats.nodes} nodes / {stats.edges} edges
      </div>

      {/* Legend */}
      <div className="absolute bottom-3 left-3 z-10 bg-background/90 backdrop-blur-sm border rounded-lg p-2">
        <div className="grid grid-cols-2 gap-x-4 gap-y-1">
          {Object.entries(NODE_COLORS).filter(([k]) => k !== "DEFAULT").map(([type, color]) => (
            <div key={type} className="flex items-center gap-1.5">
              <div className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: color }} />
              <span className="text-[10px] text-muted-foreground capitalize">{type.toLowerCase()}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Selected Node Info */}
      {selectedNode && (
        <div className="absolute bottom-3 right-3 z-10 bg-background/90 backdrop-blur-sm border rounded-lg p-3 max-w-xs">
          <h4 className="font-semibold text-sm">{selectedNode.label}</h4>
          <div className="flex items-center gap-2 mt-1">
            <span
              className="text-[10px] px-1.5 py-0.5 rounded border"
              style={{ borderColor: NODE_COLORS[selectedNode.type] || "#6b7280", color: NODE_COLORS[selectedNode.type] }}
            >
              {selectedNode.type}
            </span>
            <span className="text-[10px] text-muted-foreground">{selectedNode.degree} connections</span>
          </div>
          {Object.keys(selectedNode.properties || {}).length > 0 && (
            <div className="mt-2 space-y-0.5">
              {Object.entries(selectedNode.properties).map(([k, v]) => (
                <div key={k} className="text-[10px]">
                  <span className="text-muted-foreground">{k}:</span> <span>{String(v)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Graph */}
      <CytoscapeComponent
        elements={elements}
        stylesheet={stylesheet}
        layout={{
          name: "cose",
          animate: true,
          animationDuration: 1000,
          nodeRepulsion: () => 8000,
          idealEdgeLength: () => 120,
          gravity: 0.25,
          numIter: 300,
        } as any}
        style={{ width: "100%", height }}
        cy={handleCyInit}
        className="rounded-lg border bg-gray-900"
      />
    </div>
  );
}
