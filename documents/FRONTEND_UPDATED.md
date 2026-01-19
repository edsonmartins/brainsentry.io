# Brain Sentry - Frontend Update: Cytoscape.js Integration

**Version:** 1.1  
**Updated:** January 2025  
**Change:** Replaced React Flow with Cytoscape.js for graph visualization  

---

## Updated Dependencies

```bash
# Remove React Flow
npm uninstall reactflow

# Install Cytoscape.js
npm install cytoscape@^3.28.1 \
             react-cytoscapejs@^2.0.0 \
             cytoscape-cola@^2.5.1 \
             cytoscape-dagre@^2.5.0 \
             cytoscape-cose-bilkent@^4.1.0 \
             cytoscape-context-menus@^4.1.0
```

## Updated package.json

```json
{
  "name": "brain-sentry-frontend",
  "version": "1.0.0",
  "dependencies": {
    "next": "^15.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    
    "@radix-ui/react-dialog": "^1.0.5",
    "@radix-ui/react-dropdown-menu": "^2.0.6",
    "@radix-ui/react-tabs": "^1.0.4",
    "@radix-ui/react-select": "^2.0.0",
    "@radix-ui/react-label": "^2.0.2",
    "@radix-ui/react-slot": "^1.0.2",
    "@radix-ui/react-toast": "^1.1.5",
    
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.1.0",
    "tailwind-merge": "^2.2.0",
    "lucide-react": "^0.309.0",
    
    "zustand": "^4.4.7",
    "@tanstack/react-query": "^5.17.0",
    "react-hook-form": "^7.49.2",
    "@hookform/resolvers": "^3.3.4",
    "zod": "^3.22.4",
    
    "recharts": "^2.10.3",
    "cytoscape": "^3.28.1",
    "react-cytoscapejs": "^2.0.0",
    "cytoscape-cola": "^2.5.1",
    "cytoscape-dagre": "^2.5.0",
    "cytoscape-cose-bilkent": "^4.1.0",
    "cytoscape-context-menus": "^4.1.0",
    
    "axios": "^1.6.5",
    "date-fns": "^3.0.6"
  }
}
```

## Component Updates

### RelationshipGraph.tsx (UPDATED)

**Old (React Flow):**
```tsx
import ReactFlow from 'reactflow';
import 'reactflow/dist/style.css';
```

**New (Cytoscape.js):**
```tsx
// components/memories/RelationshipGraph.tsx
import { useRef, useEffect } from 'react';
import cytoscape, { Core } from 'cytoscape';
import cola from 'cytoscape-cola';
import { Card } from '@/components/ui/card';

cytoscape.use(cola);

interface RelationshipGraphProps {
  memoryId: string;
}

export function RelationshipGraph({ memoryId }: RelationshipGraphProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const { data: graphData, loading } = useMemoryGraph(memoryId);
  
  useEffect(() => {
    if (!containerRef.current || !graphData) return;
    
    const cy = cytoscape({
      container: containerRef.current,
      elements: [...graphData.nodes, ...graphData.edges],
      style: getCytoscapeStyle(),
      layout: {
        name: 'cose-bilkent',
        quality: 'default',
        nodeDimensionsIncludeLabels: true,
        idealEdgeLength: 100,
        nodeRepulsion: 4500,
      },
    });
    
    cyRef.current = cy;
    
    // Node click handler
    cy.on('tap', 'node', (event) => {
      const nodeId = event.target.id();
      router.push(`/memories/${nodeId}`);
    });
    
    return () => cy.destroy();
  }, [graphData]);
  
  if (loading) return <LoadingSpinner />;
  
  return (
    <Card className="h-[600px]">
      <div ref={containerRef} className="w-full h-full" />
    </Card>
  );
}
```

## Key Improvements

### 1. Better Performance
```
React Flow:  ~500 nodes max
Cytoscape:   10,000+ nodes
```

### 2. Advanced Layouts
```typescript
// Multiple algorithms available
- cola (force-directed)
- dagre (hierarchical) 
- cose-bilkent (advanced force-directed)
- circle
- grid
- concentric
```

### 3. Network Analysis
```typescript
// Built-in graph analysis
cy.elements().degree()
cy.elements().pageRank()
cy.elements().betweennessCentrality()
cy.elements().degreeCentrality()
```

### 4. Better for "Brain" Metaphor
```
✅ Network/neural network visualization
✅ Used in neuroscience (connectomics)
✅ Handles complex relationships
✅ Advanced styling per node/edge
```

## Migration Guide

For existing code using React Flow:

```typescript
// OLD: React Flow
<ReactFlow
  nodes={nodes}
  edges={edges}
  onNodeClick={handleClick}
/>

// NEW: Cytoscape
const cy = cytoscape({
  container: containerRef.current,
  elements: {
    nodes: nodes.map(n => ({ data: n })),
    edges: edges.map(e => ({ data: e }))
  }
});

cy.on('tap', 'node', handleClick);
```

## See Also

- **GRAPH_VISUALIZATION.md** - Complete Cytoscape.js guide
- **FRONTEND_SPECIFICATION.md** - Original frontend spec

---

**Status:** ✅ Updated  
**Breaking Changes:** Yes (React Flow → Cytoscape)  
**Migration Time:** ~2 hours
