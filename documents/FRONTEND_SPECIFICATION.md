# Brain Sentry - Frontend Specification

**Version:** 1.0  
**Stack:** Next.js 15 + TypeScript + Radix UI  
**Styling:** Tailwind CSS  
**State Management:** Zustand  

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Core Pages](#core-pages)
3. [Component Library](#component-library)
4. [State Management](#state-management)
5. [API Integration](#api-integration)
6. [Routing](#routing)
7. [Styling Guide](#styling-guide)

---

## Project Structure

```
brain-sentry-frontend/
├── package.json
├── tsconfig.json
├── tailwind.config.ts
├── next.config.js
├── .env.local
│
├── public/
│   ├── logo.svg
│   └── favicon.ico
│
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   ├── globals.css
│   │   │
│   │   ├── dashboard/
│   │   │   └── page.tsx
│   │   │
│   │   ├── memories/
│   │   │   ├── page.tsx
│   │   │   ├── [id]/
│   │   │   │   └── page.tsx
│   │   │   └── new/
│   │   │       └── page.tsx
│   │   │
│   │   ├── audit/
│   │   │   └── page.tsx
│   │   │
│   │   ├── settings/
│   │   │   └── page.tsx
│   │   │
│   │   └── api/
│   │       └── [...fallback].ts
│   │
│   ├── components/
│   │   ├── ui/                      # Radix UI components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── table.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── select.tsx
│   │   │   ├── input.tsx
│   │   │   ├── textarea.tsx
│   │   │   ├── form.tsx
│   │   │   ├── toast.tsx
│   │   │   └── dropdown-menu.tsx
│   │   │
│   │   ├── layout/
│   │   │   ├── AppLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   └── Footer.tsx
│   │   │
│   │   ├── dashboard/
│   │   │   ├── StatsCards.tsx
│   │   │   ├── ActivityFeed.tsx
│   │   │   ├── TopPatterns.tsx
│   │   │   └── Charts/
│   │   │       ├── InjectionRateChart.tsx
│   │   │       ├── CategoryDistribution.tsx
│   │   │       └── LatencyChart.tsx
│   │   │
│   │   ├── memories/
│   │   │   ├── MemoryList.tsx
│   │   │   ├── MemoryCard.tsx
│   │   │   ├── MemoryDetail.tsx
│   │   │   ├── MemoryForm.tsx
│   │   │   ├── MemoryFilters.tsx
│   │   │   ├── MemorySearch.tsx
│   │   │   ├── RelationshipGraph.tsx
│   │   │   └── VersionHistory.tsx
│   │   │
│   │   ├── interception/
│   │   │   ├── InterceptionTester.tsx
│   │   │   ├── PromptInput.tsx
│   │   │   ├── ContextViewer.tsx
│   │   │   └── EnhancedPromptViewer.tsx
│   │   │
│   │   ├── audit/
│   │   │   ├── AuditLogList.tsx
│   │   │   ├── AuditLogDetail.tsx
│   │   │   └── AuditFilters.tsx
│   │   │
│   │   └── common/
│   │       ├── LoadingSpinner.tsx
│   │       ├── EmptyState.tsx
│   │       ├── ErrorBoundary.tsx
│   │       ├── Pagination.tsx
│   │       └── ConfirmDialog.tsx
│   │
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── memories.ts
│   │   │   ├── interception.ts
│   │   │   ├── audit.ts
│   │   │   └── stats.ts
│   │   │
│   │   ├── hooks/
│   │   │   ├── useMemories.ts
│   │   │   ├── useInterception.ts
│   │   │   ├── useAuditLogs.ts
│   │   │   ├── useStats.ts
│   │   │   └── useToast.ts
│   │   │
│   │   ├── store/
│   │   │   ├── memoryStore.ts
│   │   │   ├── uiStore.ts
│   │   │   └── authStore.ts
│   │   │
│   │   └── utils/
│   │       ├── cn.ts
│   │       ├── formatters.ts
│   │       ├── validators.ts
│   │       └── date.ts
│   │
│   └── types/
│       ├── memory.ts
│       ├── interception.ts
│       ├── audit.ts
│       └── stats.ts
│
├── .eslintrc.json
├── .prettierrc
└── README.md
```

---

## Core Pages

### 1. Dashboard (`/dashboard`)

**Purpose:** Visão geral do sistema com métricas em tempo real.

**Components:**
```tsx
// app/dashboard/page.tsx
export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1>Brain Sentry Dashboard</h1>
      
      {/* Stats Cards */}
      <StatsCards />
      
      {/* Charts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <InjectionRateChart />
        <CategoryDistribution />
      </div>
      
      {/* Activity & Patterns */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <ActivityFeed />
        <TopPatterns />
      </div>
    </div>
  );
}
```

**Features:**
- ✅ Real-time metrics (requests, injections, latency)
- ✅ Charts (injection rate, category distribution)
- ✅ Recent activity feed
- ✅ Top used patterns
- ✅ System health indicators

---

### 2. Memories List (`/memories`)

**Purpose:** Listar e gerenciar todas as memórias.

```tsx
// app/memories/page.tsx
export default function MemoriesPage() {
  const [filters, setFilters] = useState<MemoryFilters>({});
  const { memories, loading } = useMemories(filters);
  
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1>Memories</h1>
        <Link href="/memories/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            New Memory
          </Button>
        </Link>
      </div>
      
      {/* Search & Filters */}
      <div className="space-y-4">
        <MemorySearch onSearch={handleSearch} />
        <MemoryFilters filters={filters} onChange={setFilters} />
      </div>
      
      {/* Memory List */}
      {loading ? (
        <LoadingSpinner />
      ) : memories.length === 0 ? (
        <EmptyState 
          title="No memories found"
          description="Create your first memory to get started"
        />
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {memories.map(memory => (
            <MemoryCard key={memory.id} memory={memory} />
          ))}
        </div>
      )}
      
      {/* Pagination */}
      <Pagination 
        currentPage={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </div>
  );
}
```

**Features:**
- ✅ Search by text
- ✅ Filter by category, importance, status
- ✅ Sort by various fields
- ✅ Pagination
- ✅ Bulk actions
- ✅ Quick preview

---

### 3. Memory Detail (`/memories/[id]`)

**Purpose:** Visualização detalhada e edição de memória.

```tsx
// app/memories/[id]/page.tsx
export default function MemoryDetailPage({ params }: { params: { id: string } }) {
  const { memory, loading } = useMemory(params.id);
  const [isEditing, setIsEditing] = useState(false);
  
  if (loading) return <LoadingSpinner />;
  if (!memory) return <EmptyState title="Memory not found" />;
  
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1>{memory.summary}</h1>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setIsEditing(true)}>
            Edit
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            Delete
          </Button>
        </div>
      </div>
      
      <Tabs defaultValue="details">
        <TabsList>
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="relationships">Relationships</TabsTrigger>
          <TabsTrigger value="usage">Usage Stats</TabsTrigger>
          <TabsTrigger value="history">History</TabsTrigger>
        </TabsList>
        
        <TabsContent value="details">
          <MemoryDetail memory={memory} />
        </TabsContent>
        
        <TabsContent value="relationships">
          <RelationshipGraph memoryId={memory.id} />
        </TabsContent>
        
        <TabsContent value="usage">
          <UsageStats memory={memory} />
        </TabsContent>
        
        <TabsContent value="history">
          <VersionHistory memoryId={memory.id} />
        </TabsContent>
      </Tabs>
      
      {/* Edit Dialog */}
      <Dialog open={isEditing} onOpenChange={setIsEditing}>
        <DialogContent>
          <MemoryForm 
            memory={memory}
            onSave={handleSave}
            onCancel={() => setIsEditing(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}
```

**Features:**
- ✅ Full memory details
- ✅ Relationship graph visualization
- ✅ Usage statistics
- ✅ Version history with diff
- ✅ Edit in-place
- ✅ Impact analysis

---

### 4. New Memory (`/memories/new`)

**Purpose:** Criar nova memória.

```tsx
// app/memories/new/page.tsx
export default function NewMemoryPage() {
  const router = useRouter();
  const { createMemory, loading } = useMemories();
  
  const handleSubmit = async (data: CreateMemoryRequest) => {
    const memory = await createMemory(data);
    toast.success('Memory created successfully');
    router.push(`/memories/${memory.id}`);
  };
  
  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <h1>Create New Memory</h1>
      
      <Card>
        <CardContent>
          <MemoryForm 
            onSave={handleSubmit}
            onCancel={() => router.back()}
            loading={loading}
          />
        </CardContent>
      </Card>
    </div>
  );
}
```

---

### 5. Audit Logs (`/audit`)

**Purpose:** Visualizar logs de auditoria.

```tsx
// app/audit/page.tsx
export default function AuditPage() {
  const [filters, setFilters] = useState<AuditFilters>({});
  const { logs, loading } = useAuditLogs(filters);
  
  return (
    <div className="space-y-6">
      <h1>Audit Logs</h1>
      
      <AuditFilters filters={filters} onChange={setFilters} />
      
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>Event Type</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Details</TableHead>
              <TableHead>Outcome</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {logs.map(log => (
              <TableRow key={log.id}>
                <TableCell>{formatDate(log.timestamp)}</TableCell>
                <TableCell>
                  <Badge>{log.eventType}</Badge>
                </TableCell>
                <TableCell>{log.userId}</TableCell>
                <TableCell>{log.reasoning}</TableCell>
                <TableCell>
                  <Badge variant={log.outcome === 'success' ? 'success' : 'destructive'}>
                    {log.outcome}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
}
```

---

## Component Library

### Memory Components

#### MemoryCard
```tsx
// components/memories/MemoryCard.tsx
interface MemoryCardProps {
  memory: Memory;
  onView?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}

export function MemoryCard({ memory, onView, onEdit, onDelete }: MemoryCardProps) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader>
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <CardTitle className="text-lg">{memory.summary}</CardTitle>
            <CardDescription className="mt-1">
              {memory.content.substring(0, 150)}...
            </CardDescription>
          </div>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onView?.(memory.id)}>
                <Eye className="mr-2 h-4 w-4" />
                View
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onEdit?.(memory.id)}>
                <Edit className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem 
                onClick={() => onDelete?.(memory.id)}
                className="text-destructive"
              >
                <Trash className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant={getImportanceVariant(memory.importance)}>
            {memory.importance}
          </Badge>
          <Badge variant="outline">{memory.category}</Badge>
          
          {memory.tags.map(tag => (
            <Badge key={tag} variant="secondary">
              {tag}
            </Badge>
          ))}
        </div>
        
        <div className="mt-4 flex items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <Eye className="h-3 w-3" />
            <span>{memory.accessCount} views</span>
          </div>
          <div className="flex items-center gap-1">
            <ThumbsUp className="h-3 w-3" />
            <span>{memory.helpfulnessRate * 100}% helpful</span>
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            <span>{formatRelativeTime(memory.createdAt)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
```

#### MemoryForm
```tsx
// components/memories/MemoryForm.tsx
interface MemoryFormProps {
  memory?: Memory;
  onSave: (data: CreateMemoryRequest) => void;
  onCancel: () => void;
  loading?: boolean;
}

export function MemoryForm({ memory, onSave, onCancel, loading }: MemoryFormProps) {
  const form = useForm<CreateMemoryRequest>({
    resolver: zodResolver(createMemorySchema),
    defaultValues: memory || {
      content: '',
      summary: '',
      category: 'PATTERN',
      importance: 'IMPORTANT',
      tags: [],
    },
  });
  
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSave)} className="space-y-6">
        <FormField
          control={form.control}
          name="summary"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Summary</FormLabel>
              <FormControl>
                <Input placeholder="Brief description..." {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <FormField
          control={form.control}
          name="content"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Content</FormLabel>
              <FormControl>
                <Textarea 
                  placeholder="Detailed content..."
                  rows={10}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="category"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Category</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select category" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="DECISION">Decision</SelectItem>
                    <SelectItem value="PATTERN">Pattern</SelectItem>
                    <SelectItem value="ANTIPATTERN">Anti-pattern</SelectItem>
                    <SelectItem value="DOMAIN">Domain</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
          
          <FormField
            control={form.control}
            name="importance"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Importance</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select importance" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="CRITICAL">Critical</SelectItem>
                    <SelectItem value="IMPORTANT">Important</SelectItem>
                    <SelectItem value="MINOR">Minor</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        
        <FormField
          control={form.control}
          name="tags"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Tags</FormLabel>
              <FormControl>
                <TagInput
                  value={field.value}
                  onChange={field.onChange}
                  placeholder="Add tags..."
                />
              </FormControl>
              <FormDescription>
                Press Enter to add a tag
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {memory ? 'Update' : 'Create'} Memory
          </Button>
        </div>
      </form>
    </Form>
  );
}
```

#### RelationshipGraph
```tsx
// components/memories/RelationshipGraph.tsx
import ReactFlow, { Node, Edge } from 'reactflow';
import 'reactflow/dist/style.css';

interface RelationshipGraphProps {
  memoryId: string;
}

export function RelationshipGraph({ memoryId }: RelationshipGraphProps) {
  const { relationships, loading } = useMemoryRelationships(memoryId);
  
  if (loading) return <LoadingSpinner />;
  
  const { nodes, edges } = buildGraphData(memoryId, relationships);
  
  return (
    <Card className="h-[600px]">
      <CardContent className="p-0 h-full">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          fitView
          attributionPosition="bottom-left"
        >
          <Controls />
          <MiniMap />
          <Background />
        </ReactFlow>
      </CardContent>
    </Card>
  );
}

function buildGraphData(centerId: string, relationships: Relationship[]): {
  nodes: Node[];
  edges: Edge[];
} {
  const nodes: Node[] = [
    {
      id: centerId,
      data: { label: 'Current Memory' },
      position: { x: 250, y: 250 },
      style: { background: '#3b82f6', color: 'white' },
    },
  ];
  
  const edges: Edge[] = [];
  
  relationships.forEach((rel, index) => {
    const angle = (index * 360) / relationships.length;
    const x = 250 + 200 * Math.cos((angle * Math.PI) / 180);
    const y = 250 + 200 * Math.sin((angle * Math.PI) / 180);
    
    nodes.push({
      id: rel.toMemoryId,
      data: { label: rel.summary },
      position: { x, y },
    });
    
    edges.push({
      id: `${centerId}-${rel.toMemoryId}`,
      source: centerId,
      target: rel.toMemoryId,
      label: rel.type,
      animated: rel.type === 'USED_WITH',
    });
  });
  
  return { nodes, edges };
}
```

---

## State Management

### Memory Store (Zustand)

```typescript
// lib/store/memoryStore.ts
import { create } from 'zustand';

interface MemoryState {
  memories: Memory[];
  selectedMemory: Memory | null;
  filters: MemoryFilters;
  loading: boolean;
  error: string | null;
  
  // Actions
  fetchMemories: (filters?: MemoryFilters) => Promise<void>;
  fetchMemory: (id: string) => Promise<void>;
  createMemory: (data: CreateMemoryRequest) => Promise<Memory>;
  updateMemory: (id: string, data: UpdateMemoryRequest) => Promise<Memory>;
  deleteMemory: (id: string) => Promise<void>;
  setFilters: (filters: MemoryFilters) => void;
}

export const useMemoryStore = create<MemoryState>((set, get) => ({
  memories: [],
  selectedMemory: null,
  filters: {},
  loading: false,
  error: null,
  
  fetchMemories: async (filters) => {
    set({ loading: true, error: null });
    try {
      const memories = await memoriesApi.list(filters);
      set({ memories, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },
  
  fetchMemory: async (id) => {
    set({ loading: true, error: null });
    try {
      const memory = await memoriesApi.get(id);
      set({ selectedMemory: memory, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },
  
  createMemory: async (data) => {
    set({ loading: true, error: null });
    try {
      const memory = await memoriesApi.create(data);
      set(state => ({
        memories: [memory, ...state.memories],
        loading: false,
      }));
      return memory;
    } catch (error) {
      set({ error: error.message, loading: false });
      throw error;
    }
  },
  
  updateMemory: async (id, data) => {
    set({ loading: true, error: null });
    try {
      const memory = await memoriesApi.update(id, data);
      set(state => ({
        memories: state.memories.map(m => m.id === id ? memory : m),
        selectedMemory: memory,
        loading: false,
      }));
      return memory;
    } catch (error) {
      set({ error: error.message, loading: false });
      throw error;
    }
  },
  
  deleteMemory: async (id) => {
    set({ loading: true, error: null });
    try {
      await memoriesApi.delete(id);
      set(state => ({
        memories: state.memories.filter(m => m.id !== id),
        loading: false,
      }));
    } catch (error) {
      set({ error: error.message, loading: false });
      throw error;
    }
  },
  
  setFilters: (filters) => {
    set({ filters });
    get().fetchMemories(filters);
  },
}));
```

---

## API Integration

### API Client

```typescript
// lib/api/client.ts
import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor (auth)
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor (error handling)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### Memories API

```typescript
// lib/api/memories.ts
import { apiClient } from './client';

export const memoriesApi = {
  list: async (filters?: MemoryFilters): Promise<Memory[]> => {
    const { data } = await apiClient.get('/v1/memories', { params: filters });
    return data.content;
  },
  
  get: async (id: string): Promise<Memory> => {
    const { data } = await apiClient.get(`/v1/memories/${id}`);
    return data;
  },
  
  create: async (request: CreateMemoryRequest): Promise<Memory> => {
    const { data } = await apiClient.post('/v1/memories', request);
    return data;
  },
  
  update: async (id: string, request: UpdateMemoryRequest): Promise<Memory> => {
    const { data } = await apiClient.put(`/v1/memories/${id}`, request);
    return data;
  },
  
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/memories/${id}`);
  },
  
  search: async (request: SearchRequest): Promise<SearchResponse> => {
    const { data } = await apiClient.post('/v1/memories/search', request);
    return data;
  },
};
```

---

## Styling Guide

### Tailwind Configuration

```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: ['class'],
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
};

export default config;
```

### Design Tokens

```css
/* app/globals.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 221.2 83.2% 53.3%;
    
    --radius: 0.5rem;
  }
  
  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 47.4% 11.2%;
    
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 48%;
  }
}
```

---

## TypeScript Types

```typescript
// types/memory.ts
export interface Memory {
  id: string;
  content: string;
  summary: string;
  category: MemoryCategory;
  importance: ImportanceLevel;
  validationStatus: ValidationStatus;
  embedding?: number[];
  metadata?: Record<string, any>;
  tags: string[];
  sourceType: string;
  sourceReference?: string;
  createdBy: string;
  createdAt: string;
  updatedAt?: string;
  lastAccessedAt?: string;
  version: number;
  accessCount: number;
  injectionCount: number;
  helpfulCount: number;
  notHelpfulCount: number;
  helpfulnessRate: number;
  codeExample?: string;
  programmingLanguage?: string;
  relatedMemories?: RelatedMemory[];
}

export type MemoryCategory = 
  | 'DECISION'
  | 'PATTERN'
  | 'ANTIPATTERN'
  | 'DOMAIN'
  | 'BUG'
  | 'OPTIMIZATION'
  | 'INTEGRATION';

export type ImportanceLevel = 'CRITICAL' | 'IMPORTANT' | 'MINOR';

export type ValidationStatus = 'APPROVED' | 'PENDING' | 'FLAGGED' | 'REJECTED';

export interface CreateMemoryRequest {
  content: string;
  summary?: string;
  category: MemoryCategory;
  importance?: ImportanceLevel;
  tags?: string[];
  sourceType: string;
  codeExample?: string;
  programmingLanguage?: string;
}

export interface MemoryFilters {
  category?: MemoryCategory;
  importance?: ImportanceLevel;
  status?: ValidationStatus;
  search?: string;
  page?: number;
  size?: number;
}
```

---

**Document Status:** ✅ Complete  
**Ready for:** Phase 1 implementation  
**Dependencies:** Backend API running
