# Brain Sentry - Graph Visualization with Cytoscape.js

**Version:** 1.0  
**Date:** January 2025  
**Library:** Cytoscape.js 3.28+  

---

## Table of Contents

1. [Why Cytoscape.js](#why-cytoscapejs)
2. [Setup & Installation](#setup--installation)
3. [Memory Graph Component](#memory-graph-component)
4. [Layout Algorithms](#layout-algorithms)
5. [Interactive Features](#interactive-features)
6. [Styling & Themes](#styling--themes)
7. [Performance Optimization](#performance-optimization)
8. [Integration Examples](#integration-examples)

---

## 1. Why Cytoscape.js

### 1.1 Comparison with Alternatives

```
┌────────────────────────────────────────────────────────┐
│  Feature           │ React Flow │ Cytoscape │ D3.js   │
├────────────────────────────────────────────────────────┤
│ Max Nodes          │    ~500    │  10,000+  │ ~1,000  │
│ Layout Algorithms  │   Basic    │ Advanced  │ Custom  │
│ Network Analysis   │     ❌     │    ✅     │   ⚠️    │
│ React Integration  │     ✅     │    ✅     │   ⚠️    │
│ Performance        │   Medium   │    High   │  Medium │
│ Learning Curve     │    Easy    │  Medium   │   Hard  │
│ Built for Graphs   │     ⚠️     │    ✅     │   ⚠️    │
└────────────────────────────────────────────────────────┘
```

### 1.2 Perfect for Brain Sentry

**Brain Sentry é um "grafo de conhecimento":**
- Memórias = Nós (nodes)
- Relacionamentos = Arestas (edges)
- USED_WITH, CONFLICTS_WITH, SUPERSEDES = Tipos de arestas
- Precisa de layouts inteligentes
- Análise de clusters (patterns relacionados)

**Cytoscape.js foi feito para isso:**
- Usado em bioinformática (redes de proteínas)
- Análise de redes sociais
- Knowledge graphs
- Network topology

---

## 2. Setup & Installation

### 2.1 Install Dependencies

```bash
# Core library
npm install cytoscape

# React wrapper
npm install react-cytoscapejs

# Layout extensions
npm install cytoscape-cola        # Force-directed
npm install cytoscape-dagre       # Hierarchical
npm install cytoscape-cose-bilkent # Advanced force-directed

# Additional utilities
npm install cytoscape-context-menus
npm install cytoscape-panzoom
```

### 2.2 Package.json

```json
{
  "dependencies": {
    "cytoscape": "^3.28.1",
    "react-cytoscapejs": "^2.0.0",
    "cytoscape-cola": "^2.5.1",
    "cytoscape-dagre": "^2.5.0",
    "cytoscape-cose-bilkent": "^4.1.0",
    "cytoscape-context-menus": "^4.1.0",
    "cytoscape-panzoom": "^2.5.3"
  }
}
```

### 2.3 TypeScript Types

```typescript
// types/cytoscape.d.ts
import cytoscape from 'cytoscape';

declare module 'cytoscape' {
  interface Core {
    panzoom: (options?: any) => void;
  }
}

export interface MemoryNode {
  data: {
    id: string;
    label: string;
    category: MemoryCategory;
    importance: ImportanceLevel;
    summary: string;
    accessCount: number;
    helpfulnessRate: number;
  };
  classes?: string[];
}

export interface MemoryEdge {
  data: {
    id: string;
    source: string;
    target: string;
    type: RelationshipType;
    frequency?: number;
    strength?: number;
  };
  classes?: string[];
}

export type CytoscapeData = {
  nodes: MemoryNode[];
  edges: MemoryEdge[];
};
```

---

## 3. Memory Graph Component

### 3.1 Base Component

```tsx
// components/graph/MemoryGraph.tsx
'use client';

import { useEffect, useRef, useState } from 'react';
import cytoscape, { Core, ElementDefinition } from 'cytoscape';
import cola from 'cytoscape-cola';
import dagre from 'cytoscape-dagre';
import coseBilkent from 'cytoscape-cose-bilkent';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { 
  ZoomIn, ZoomOut, Maximize2, RefreshCw, Download 
} from 'lucide-react';

// Register layouts
cytoscape.use(cola);
cytoscape.use(dagre);
cytoscape.use(coseBilkent);

interface MemoryGraphProps {
  memoryId?: string;  // Center node (optional)
  depth?: number;     // How many levels to show
  onNodeClick?: (nodeId: string) => void;
  onEdgeClick?: (edgeId: string) => void;
  layout?: 'cola' | 'dagre' | 'cose-bilkent' | 'circle';
}

export function MemoryGraph({
  memoryId,
  depth = 2,
  onNodeClick,
  onEdgeClick,
  layout = 'cose-bilkent'
}: MemoryGraphProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // Fetch graph data
  const { data: graphData, isLoading } = useMemoryGraph(memoryId, depth);

  useEffect(() => {
    if (!containerRef.current || !graphData) return;

    // Initialize Cytoscape
    const cy = cytoscape({
      container: containerRef.current,
      elements: [...graphData.nodes, ...graphData.edges],
      style: getCytoscapeStyle(),
      layout: getLayoutConfig(layout),
      minZoom: 0.1,
      maxZoom: 3,
      wheelSensitivity: 0.2,
    });

    cyRef.current = cy;

    // Event listeners
    cy.on('tap', 'node', (event) => {
      const node = event.target;
      const nodeId = node.id();
      setSelectedNode(nodeId);
      onNodeClick?.(nodeId);
      
      // Highlight connected
      highlightConnected(cy, nodeId);
    });

    cy.on('tap', 'edge', (event) => {
      const edge = event.target;
      onEdgeClick?.(edge.id());
    });

    // Double click to expand
    cy.on('dbltap', 'node', (event) => {
      const nodeId = event.target.id();
      expandNode(nodeId);
    });

    setLoading(false);

    return () => {
      cy.destroy();
    };
  }, [graphData, layout]);

  // Toolbar actions
  const handleZoomIn = () => cyRef.current?.zoom(cyRef.current.zoom() * 1.2);
  const handleZoomOut = () => cyRef.current?.zoom(cyRef.current.zoom() * 0.8);
  const handleFit = () => cyRef.current?.fit();
  const handleReset = () => cyRef.current?.layout(getLayoutConfig(layout)).run();
  const handleExport = () => exportGraph(cyRef.current);

  if (isLoading || loading) {
    return (
      <Card className="h-[600px] flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-2" />
          <p>Loading graph...</p>
        </div>
      </Card>
    );
  }

  return (
    <Card className="relative h-[600px]">
      {/* Toolbar */}
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Button size="sm" variant="outline" onClick={handleZoomIn}>
          <ZoomIn className="h-4 w-4" />
        </Button>
        <Button size="sm" variant="outline" onClick={handleZoomOut}>
          <ZoomOut className="h-4 w-4" />
        </Button>
        <Button size="sm" variant="outline" onClick={handleFit}>
          <Maximize2 className="h-4 w-4" />
        </Button>
        <Button size="sm" variant="outline" onClick={handleReset}>
          <RefreshCw className="h-4 w-4" />
        </Button>
        <Button size="sm" variant="outline" onClick={handleExport}>
          <Download className="h-4 w-4" />
        </Button>
      </div>

      {/* Graph container */}
      <div ref={containerRef} className="w-full h-full" />

      {/* Selected node info */}
      {selectedNode && (
        <div className="absolute bottom-4 left-4 z-10">
          <Card className="p-4 max-w-xs">
            <NodeInfo nodeId={selectedNode} />
          </Card>
        </div>
      )}
    </Card>
  );
}
```

### 3.2 Cytoscape Styling

```typescript
// lib/graph/styles.ts
import { Stylesheet } from 'cytoscape';

export function getCytoscapeStyle(): Stylesheet[] {
  return [
    // ===== NODES =====
    {
      selector: 'node',
      style: {
        'label': 'data(label)',
        'text-valign': 'center',
        'text-halign': 'center',
        'background-color': '#3b82f6',
        'color': '#fff',
        'font-size': '12px',
        'font-weight': 'bold',
        'width': 'label',
        'height': 'label',
        'padding': '10px',
        'shape': 'roundrectangle',
        'text-wrap': 'wrap',
        'text-max-width': '100px',
        'border-width': 2,
        'border-color': '#1e40af',
      },
    },

    // Center node (if specified)
    {
      selector: 'node.center',
      style: {
        'background-color': '#8b5cf6',
        'border-color': '#6d28d9',
        'border-width': 4,
        'width': 80,
        'height': 80,
        'font-size': '14px',
        'z-index': 100,
      },
    },

    // By importance
    {
      selector: 'node.critical',
      style: {
        'background-color': '#ef4444',
        'border-color': '#b91c1c',
      },
    },
    {
      selector: 'node.important',
      style: {
        'background-color': '#f59e0b',
        'border-color': '#d97706',
      },
    },
    {
      selector: 'node.minor',
      style: {
        'background-color': '#10b981',
        'border-color': '#059669',
      },
    },

    // By category
    {
      selector: 'node.decision',
      style: {
        'shape': 'diamond',
      },
    },
    {
      selector: 'node.pattern',
      style: {
        'shape': 'hexagon',
      },
    },
    {
      selector: 'node.antipattern',
      style: {
        'shape': 'triangle',
        'background-color': '#dc2626',
      },
    },

    // Hover
    {
      selector: 'node:hover',
      style: {
        'border-width': 4,
        'border-color': '#fbbf24',
      },
    },

    // Selected
    {
      selector: 'node:selected',
      style: {
        'border-width': 5,
        'border-color': '#fbbf24',
        'background-color': '#8b5cf6',
      },
    },

    // ===== EDGES =====
    {
      selector: 'edge',
      style: {
        'width': 2,
        'line-color': '#94a3b8',
        'target-arrow-color': '#94a3b8',
        'target-arrow-shape': 'triangle',
        'curve-style': 'bezier',
        'label': 'data(type)',
        'font-size': '10px',
        'text-rotation': 'autorotate',
        'text-margin-y': -10,
      },
    },

    // By relationship type
    {
      selector: 'edge.used-with',
      style: {
        'line-color': '#10b981',
        'target-arrow-color': '#10b981',
        'width': 'data(frequency)',
        'line-style': 'solid',
      },
    },
    {
      selector: 'edge.conflicts-with',
      style: {
        'line-color': '#ef4444',
        'target-arrow-color': '#ef4444',
        'line-style': 'dashed',
        'width': 3,
      },
    },
    {
      selector: 'edge.supersedes',
      style: {
        'line-color': '#8b5cf6',
        'target-arrow-color': '#8b5cf6',
        'line-style': 'dotted',
      },
    },
    {
      selector: 'edge.related-to',
      style: {
        'line-color': '#6b7280',
        'target-arrow-color': '#6b7280',
        'line-style': 'solid',
        'width': 1,
      },
    },

    // Highlighted (when node selected)
    {
      selector: 'edge.highlighted',
      style: {
        'width': 4,
        'line-color': '#fbbf24',
        'target-arrow-color': '#fbbf24',
        'z-index': 999,
      },
    },

    // Dimmed (when others highlighted)
    {
      selector: 'node.dimmed',
      style: {
        'opacity': 0.3,
      },
    },
    {
      selector: 'edge.dimmed',
      style: {
        'opacity': 0.2,
      },
    },
  ];
}
```

---

## 4. Layout Algorithms

### 4.1 Layout Configurations

```typescript
// lib/graph/layouts.ts
import { LayoutOptions } from 'cytoscape';

export function getLayoutConfig(
  layout: 'cola' | 'dagre' | 'cose-bilkent' | 'circle'
): LayoutOptions {
  const configs: Record<string, LayoutOptions> = {
    // Force-directed (best for general graphs)
    cola: {
      name: 'cola',
      animate: true,
      refresh: 1,
      maxSimulationTime: 4000,
      ungrabifyWhileSimulating: false,
      fit: true,
      padding: 30,
      nodeDimensionsIncludeLabels: true,
      randomize: false,
      avoidOverlap: true,
      handleDisconnected: true,
      convergenceThreshold: 0.01,
      nodeSpacing: 50,
      flow: undefined,
      alignment: undefined,
      gapInequalities: undefined,
    },

    // Hierarchical (best for supersedes relationships)
    dagre: {
      name: 'dagre',
      nodeSep: 50,
      edgeSep: 10,
      rankSep: 100,
      rankDir: 'TB', // Top to bottom
      ranker: 'network-simplex',
      minLen: (edge: any) => {
        return edge.data('type') === 'SUPERSEDES' ? 2 : 1;
      },
      edgeWeight: (edge: any) => {
        return edge.data('frequency') || 1;
      },
    },

    // Advanced force-directed (best for large graphs)
    'cose-bilkent': {
      name: 'cose-bilkent',
      quality: 'default', // 'draft', 'default', 'proof'
      nodeDimensionsIncludeLabels: true,
      refresh: 30,
      fit: true,
      padding: 30,
      randomize: true,
      nodeRepulsion: 4500,
      idealEdgeLength: 100,
      edgeElasticity: 0.45,
      nestingFactor: 0.1,
      gravity: 0.25,
      numIter: 2500,
      tile: true,
      animate: 'end',
      animationDuration: 500,
      tilingPaddingVertical: 10,
      tilingPaddingHorizontal: 10,
      gravityRangeCompound: 1.5,
      gravityCompound: 1.0,
      gravityRange: 3.8,
    },

    // Circular (good for small, well-connected graphs)
    circle: {
      name: 'circle',
      fit: true,
      padding: 30,
      boundingBox: undefined,
      avoidOverlap: true,
      nodeDimensionsIncludeLabels: false,
      spacingFactor: 1.5,
      radius: undefined,
      startAngle: (3 / 2) * Math.PI,
      sweep: undefined,
      clockwise: true,
      sort: (a: any, b: any) => {
        // Sort by importance
        const importanceOrder = { CRITICAL: 0, IMPORTANT: 1, MINOR: 2 };
        return (
          importanceOrder[a.data('importance')] -
          importanceOrder[b.data('importance')]
        );
      },
    },
  };

  return configs[layout];
}
```

### 4.2 Layout Selector Component

```tsx
// components/graph/LayoutSelector.tsx
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface LayoutSelectorProps {
  value: string;
  onChange: (layout: string) => void;
}

export function LayoutSelector({ value, onChange }: LayoutSelectorProps) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-48">
        <SelectValue placeholder="Select layout" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="cose-bilkent">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-blue-500 rounded" />
            <div>
              <div className="font-medium">Cose-Bilkent</div>
              <div className="text-xs text-muted-foreground">
                Force-directed (recommended)
              </div>
            </div>
          </div>
        </SelectItem>
        
        <SelectItem value="cola">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-green-500 rounded" />
            <div>
              <div className="font-medium">Cola</div>
              <div className="text-xs text-muted-foreground">
                Simple force-directed
              </div>
            </div>
          </div>
        </SelectItem>
        
        <SelectItem value="dagre">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-purple-500 rounded" />
            <div>
              <div className="font-medium">Dagre</div>
              <div className="text-xs text-muted-foreground">
                Hierarchical (good for supersedes)
              </div>
            </div>
          </div>
        </SelectItem>
        
        <SelectItem value="circle">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-orange-500 rounded" />
            <div>
              <div className="font-medium">Circle</div>
              <div className="text-xs text-muted-foreground">
                Circular layout
              </div>
            </div>
          </div>
        </SelectItem>
      </SelectContent>
    </Select>
  );
}
```

---

## 5. Interactive Features

### 5.1 Highlight Connected Nodes

```typescript
// lib/graph/interactions.ts
import { Core } from 'cytoscape';

export function highlightConnected(cy: Core, nodeId: string) {
  // Reset all
  cy.elements().removeClass('highlighted dimmed');
  
  // Get node and neighbors
  const node = cy.getElementById(nodeId);
  const connected = node.neighborhood();
  
  // Highlight
  node.addClass('highlighted');
  connected.addClass('highlighted');
  
  // Dim others
  cy.elements().difference(connected.union(node)).addClass('dimmed');
}

export function clearHighlight(cy: Core) {
  cy.elements().removeClass('highlighted dimmed');
}
```

### 5.2 Expand/Collapse Nodes

```typescript
// lib/graph/expand.ts
export async function expandNode(
  cy: Core,
  nodeId: string,
  depth: number = 1
) {
  // Fetch additional neighbors
  const response = await fetch(
    `/api/v1/memories/${nodeId}/neighbors?depth=${depth}`
  );
  const data = await response.json();
  
  // Add new nodes and edges
  data.nodes.forEach((nodeData: any) => {
    if (!cy.getElementById(nodeData.data.id).length) {
      cy.add({
        group: 'nodes',
        data: nodeData.data,
        classes: nodeData.classes,
      });
    }
  });
  
  data.edges.forEach((edgeData: any) => {
    if (!cy.getElementById(edgeData.data.id).length) {
      cy.add({
        group: 'edges',
        data: edgeData.data,
        classes: edgeData.classes,
      });
    }
  });
  
  // Re-run layout
  cy.layout(getLayoutConfig('cose-bilkent')).run();
}
```

### 5.3 Context Menu

```typescript
// lib/graph/contextMenu.ts
import contextMenus from 'cytoscape-context-menus';
import 'cytoscape-context-menus/cytoscape-context-menus.css';

cytoscape.use(contextMenus);

export function initContextMenu(cy: Core, handlers: any) {
  cy.contextMenus({
    menuItems: [
      {
        id: 'expand',
        content: 'Expand',
        selector: 'node',
        onClickFunction: (event: any) => {
          const node = event.target || event.cyTarget;
          handlers.onExpand?.(node.id());
        },
      },
      {
        id: 'view-details',
        content: 'View Details',
        selector: 'node',
        onClickFunction: (event: any) => {
          const node = event.target || event.cyTarget;
          handlers.onViewDetails?.(node.id());
        },
      },
      {
        id: 'hide',
        content: 'Hide',
        selector: 'node',
        onClickFunction: (event: any) => {
          const node = event.target || event.cyTarget;
          node.style('display', 'none');
          node.connectedEdges().style('display', 'none');
        },
      },
      {
        id: 'focus',
        content: 'Focus on this',
        selector: 'node',
        onClickFunction: (event: any) => {
          const node = event.target || event.cyTarget;
          cy.animate({
            fit: { eles: node.neighborhood().union(node), padding: 50 },
            duration: 500,
          });
        },
      },
      {
        id: 'separator',
        content: '---',
      },
      {
        id: 'remove-relationship',
        content: 'Remove Relationship',
        selector: 'edge',
        onClickFunction: (event: any) => {
          const edge = event.target || event.cyTarget;
          handlers.onRemoveEdge?.(edge.id());
        },
      },
    ],
  });
}
```

---

## 6. Styling & Themes

### 6.1 Dark Mode Support

```typescript
// lib/graph/themes.ts
export function getDarkModeStyle(): Stylesheet[] {
  return [
    {
      selector: 'node',
      style: {
        'background-color': '#1e293b',
        'color': '#f1f5f9',
        'border-color': '#475569',
      },
    },
    {
      selector: 'edge',
      style: {
        'line-color': '#475569',
        'target-arrow-color': '#475569',
      },
    },
    // ... other dark mode styles
  ];
}

// Usage
const isDark = useTheme() === 'dark';
const style = isDark ? getDarkModeStyle() : getCytoscapeStyle();
```

### 6.2 Custom Node Renderer

```typescript
// For very custom nodes, use HTML overlays
export function addNodeOverlay(cy: Core, nodeId: string, content: string) {
  const node = cy.getElementById(nodeId);
  const position = node.renderedPosition();
  
  const overlay = document.createElement('div');
  overlay.className = 'node-overlay';
  overlay.innerHTML = content;
  overlay.style.position = 'absolute';
  overlay.style.left = `${position.x}px`;
  overlay.style.top = `${position.y}px`;
  
  cy.container()?.appendChild(overlay);
}
```

---

## 7. Performance Optimization

### 7.1 Lazy Loading

```typescript
// Load graph in chunks
export async function loadGraphProgressive(
  cy: Core,
  centerNodeId: string,
  maxDepth: number = 3
) {
  // Start with center node
  await loadLevel(cy, centerNodeId, 0);
  
  // Load subsequent levels
  for (let depth = 1; depth <= maxDepth; depth++) {
    await loadLevel(cy, centerNodeId, depth);
    await new Promise(resolve => setTimeout(resolve, 500));
  }
}

async function loadLevel(cy: Core, nodeId: string, depth: number) {
  const data = await fetchGraphData(nodeId, depth, depth);
  cy.add(data.nodes);
  cy.add(data.edges);
  cy.layout(getLayoutConfig('cose-bilkent')).run();
}
```

### 7.2 Viewport Culling

```typescript
// Only render nodes in viewport
export function enableViewportCulling(cy: Core) {
  cy.on('render', () => {
    const extent = cy.extent();
    
    cy.nodes().forEach((node) => {
      const pos = node.position();
      const inViewport = 
        pos.x >= extent.x1 && pos.x <= extent.x2 &&
        pos.y >= extent.y1 && pos.y <= extent.y2;
      
      node.style('display', inViewport ? 'element' : 'none');
    });
  });
}
```

### 7.3 Debounced Search

```typescript
// Debounce graph updates
import { debounce } from 'lodash';

const debouncedLayout = debounce((cy: Core, config: any) => {
  cy.layout(config).run();
}, 300);
```

---

## 8. Integration Examples

### 8.1 Full Page with Controls

```tsx
// app/memories/[id]/graph/page.tsx
export default function MemoryGraphPage({ params }: { params: { id: string } }) {
  const [layout, setLayout] = useState<LayoutType>('cose-bilkent');
  const [depth, setDepth] = useState(2);
  const [filters, setFilters] = useState<GraphFilters>({});

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h1>Memory Graph: {params.id}</h1>
        
        <div className="flex gap-2">
          <LayoutSelector value={layout} onChange={setLayout} />
          
          <Select value={depth.toString()} onValueChange={(v) => setDepth(+v)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1">Depth: 1</SelectItem>
              <SelectItem value="2">Depth: 2</SelectItem>
              <SelectItem value="3">Depth: 3</SelectItem>
              <SelectItem value="4">Depth: 4</SelectItem>
            </SelectContent>
          </Select>
          
          <GraphFilters filters={filters} onChange={setFilters} />
        </div>
      </div>
      
      <MemoryGraph
        memoryId={params.id}
        depth={depth}
        layout={layout}
        filters={filters}
        onNodeClick={(nodeId) => router.push(`/memories/${nodeId}`)}
      />
      
      <GraphLegend />
    </div>
  );
}
```

### 8.2 Embedded Mini Graph

```tsx
// components/memories/MiniGraph.tsx
export function MiniGraph({ memoryId }: { memoryId: string }) {
  return (
    <Card className="h-64">
      <CardHeader>
        <CardTitle className="text-sm">Relationships</CardTitle>
      </CardHeader>
      <CardContent className="h-48">
        <MemoryGraph
          memoryId={memoryId}
          depth={1}
          layout="circle"
          showToolbar={false}
        />
      </CardContent>
    </Card>
  );
}
```

### 8.3 Analytics Integration

```typescript
// Track graph interactions
cy.on('tap', 'node', (event) => {
  analytics.track('graph_node_clicked', {
    nodeId: event.target.id(),
    category: event.target.data('category'),
    importance: event.target.data('importance'),
  });
});

cy.on('layoutstop', (event) => {
  analytics.track('graph_layout_complete', {
    layout: event.layout.options.name,
    nodeCount: cy.nodes().length,
    edgeCount: cy.edges().length,
    duration: event.layout.options.maxSimulationTime,
  });
});
```

---

## 9. API Integration

### 9.1 Fetch Graph Data

```typescript
// lib/api/graph.ts
export async function fetchMemoryGraph(
  memoryId: string,
  depth: number = 2
): Promise<CytoscapeData> {
  const response = await apiClient.get(`/v1/memories/${memoryId}/graph`, {
    params: { depth },
  });
  
  return transformToCytoscape(response.data);
}

function transformToCytoscape(apiData: any): CytoscapeData {
  const nodes: MemoryNode[] = apiData.memories.map((memory: any) => ({
    data: {
      id: memory.id,
      label: memory.summary,
      category: memory.category,
      importance: memory.importance,
      summary: memory.summary,
      accessCount: memory.accessCount,
      helpfulnessRate: memory.helpfulnessRate,
    },
    classes: [
      memory.importance.toLowerCase(),
      memory.category.toLowerCase(),
      memory.id === apiData.centerId ? 'center' : '',
    ].filter(Boolean),
  }));
  
  const edges: MemoryEdge[] = apiData.relationships.map((rel: any) => ({
    data: {
      id: `${rel.fromId}-${rel.toId}`,
      source: rel.fromId,
      target: rel.toId,
      type: rel.type,
      frequency: rel.frequency,
      strength: rel.strength,
    },
    classes: [rel.type.toLowerCase().replace(/_/g, '-')],
  }));
  
  return { nodes, edges };
}
```

---

## 10. Best Practices

### 10.1 Performance

```typescript
// ✅ Good: Load incrementally
loadGraphProgressive(cy, centerId, 3);

// ❌ Bad: Load everything at once
loadEntireGraph(cy);

// ✅ Good: Debounce expensive operations
const debouncedLayout = debounce(() => cy.layout(...).run(), 300);

// ✅ Good: Limit initial render
const MAX_INITIAL_NODES = 100;
```

### 10.2 UX

```typescript
// Show loading state
cy.on('layoutstart', () => setLoading(true));
cy.on('layoutstop', () => setLoading(false));

// Animate transitions
cy.animate({ fit: { eles: selection, padding: 50 }, duration: 500 });

// Provide clear feedback
cy.on('tap', 'node', (e) => {
  toast.success(`Selected: ${e.target.data('label')}`);
});
```

### 10.3 Accessibility

```typescript
// Add keyboard navigation
document.addEventListener('keydown', (e) => {
  if (e.key === 'Tab') {
    const selected = cy.$(':selected');
    const next = selected.neighborhood('node').first();
    cy.$(selected).unselect();
    cy.$(next).select();
  }
});

// Add ARIA labels
containerRef.current?.setAttribute('role', 'img');
containerRef.current?.setAttribute('aria-label', 'Memory relationship graph');
```

---

## Summary: Why Cytoscape.js for Brain Sentry

```
✅ Built specifically for graph visualization
✅ Handles 10,000+ nodes smoothly
✅ Advanced layout algorithms (force-directed, hierarchical)
✅ Network analysis capabilities
✅ React integration (react-cytoscapejs)
✅ Highly customizable styling
✅ Rich interaction events
✅ Perfect for "brain" metaphor

Result: Professional, scalable graph visualization
        that can grow with your memory database
```

---

**Status:** ✅ Ready for Implementation  
**Next:** Integrate into `FRONTEND_SPECIFICATION.md`
