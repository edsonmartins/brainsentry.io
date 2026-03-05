package handler

import (
	"encoding/json"
	"net/http"
)

// SwaggerSpec returns a hand-crafted OpenAPI 3.0 spec for the Brain Sentry API.
// For a full generated spec, install swaggo/swag and run: swag init -g docs/swagger.go
func SwaggerSpec(w http.ResponseWriter, r *http.Request) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Brain Sentry API",
			"description": "AI Agent Memory System - manages memories, context injection, entity graphs, notes, and MCP protocol",
			"version":     "1.0.0",
		},
		"servers": []map[string]string{
			{"url": "/api", "description": "Default server"},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
		"tags": []map[string]string{
			{"name": "Auth", "description": "Authentication"},
			{"name": "Memories", "description": "Memory CRUD and search"},
			{"name": "Interception", "description": "Prompt interception and context injection"},
			{"name": "Relationships", "description": "Memory relationships"},
			{"name": "Entity Graph", "description": "FalkorDB entity graph"},
			{"name": "Notes", "description": "Session analysis and notes"},
			{"name": "Compression", "description": "Context compression"},
			{"name": "MCP", "description": "Model Context Protocol"},
			{"name": "Audit", "description": "Audit logs"},
			{"name": "Stats", "description": "System statistics"},
			{"name": "Users", "description": "User management"},
			{"name": "Tenants", "description": "Tenant management"},
		},
		"paths": buildPaths(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func buildPaths() map[string]any {
	return map[string]any{ //nolint:funlen
		// Auth
		"/v1/auth/login":          map[string]any{"post": endpoint("Auth", "Login", "Authenticate and receive JWT", reqBody("LoginRequest"), resp200("LoginResponse"))},
		"/v1/auth/logout":         map[string]any{"post": endpoint("Auth", "Logout", "Logout (client discards token)", nil, resp200("message"))},
		"/v1/auth/refresh":        map[string]any{"post": endpoint("Auth", "Refresh token", "Refresh JWT token", nil, resp200("LoginResponse"))},
		"/v1/auth/sso/authorize":  map[string]any{"get": endpoint("Auth", "SSO authorize", "Get SSO authorization URL", nil, resp200("SSOAuthURL"))},
		"/v1/auth/sso/callback":   map[string]any{"post": endpoint("Auth", "SSO callback", "Handle SSO callback", reqBody("SSOCallbackRequest"), resp200("LoginResponse"))},
		"/v1/auth/sso/config":     map[string]any{"get": endpoint("Auth", "SSO config", "Get SSO configuration", nil, resp200("SSOConfig"))},

		// Users
		"/v1/users":      map[string]any{"get": endpoint("Users", "List users", "List users for current tenant", nil, resp200("User[]")), "post": endpoint("Users", "Create user", "Create a new user", reqBody("CreateUserRequest"), resp201("User"))},
		"/v1/users/{id}": map[string]any{"get": endpointWithParams("Users", "Get user", "Get user by ID", []map[string]any{pathParam("id", "User ID")}, resp200("User"))},

		// Tenants
		"/v1/tenants":      map[string]any{"get": endpoint("Tenants", "List tenants", "List all tenants", nil, resp200("Tenant[]")), "post": endpoint("Tenants", "Create tenant", "Create a new tenant", reqBody("CreateTenantRequest"), resp201("Tenant"))},
		"/v1/tenants/{id}": map[string]any{"get": endpointWithParams("Tenants", "Get tenant", "Get tenant by ID", []map[string]any{pathParam("id", "Tenant ID")}, resp200("Tenant"))},

		// Memories
		"/v1/memories": map[string]any{
			"get":  endpointWithParams("Memories", "List memories", "Paginated list of memories", []map[string]any{queryParam("page", "integer", "Page number"), queryParam("size", "integer", "Page size")}, resp200("MemoryListResponse")),
			"post": endpoint("Memories", "Create memory", "Create a new memory with auto-analysis", reqBody("CreateMemoryRequest"), resp201("MemoryResponse")),
		},
		"/v1/memories/search":                           map[string]any{"post": endpoint("Memories", "Search memories", "Full-text and vector search", reqBody("SearchRequest"), resp200("SearchResponse"))},
		"/v1/memories/by-category/{category}":           map[string]any{"get": endpointWithParams("Memories", "By category", "Get memories by category", []map[string]any{pathParam("category", "Category")}, resp200("MemoryResponse[]"))},
		"/v1/memories/by-importance/{importance}":       map[string]any{"get": endpointWithParams("Memories", "By importance", "Get memories by importance", []map[string]any{pathParam("importance", "Importance")}, resp200("MemoryResponse[]"))},
		"/v1/memories/{id}":                             map[string]any{"get": endpointWithParams("Memories", "Get memory", "Get memory by ID with related memories", []map[string]any{pathParam("id", "Memory ID")}, resp200("MemoryResponse")), "put": endpoint("Memories", "Update memory", "Update memory with versioning", reqBody("UpdateMemoryRequest"), resp200("MemoryResponse")), "delete": endpointWithParams("Memories", "Delete memory", "Soft-delete a memory", []map[string]any{pathParam("id", "Memory ID")}, resp200("message"))},
		"/v1/memories/{id}/versions":                    map[string]any{"get": endpointWithParams("Memories", "Version history", "Get memory version history", []map[string]any{pathParam("id", "Memory ID")}, resp200("MemoryVersion[]"))},
		"/v1/memories/{id}/feedback":                    map[string]any{"post": endpointWithParams("Memories", "Record feedback", "Record helpful/not helpful feedback", []map[string]any{pathParam("id", "Memory ID")}, resp200("message"))},
		"/v1/memories/{id}/flag":                        map[string]any{"post": endpoint("Memories", "Flag memory", "Flag a memory as incorrect", reqBody("FlagMemoryRequest"), resp200("message"))},
		"/v1/memories/{id}/review":                      map[string]any{"post": endpoint("Memories", "Review correction", "Approve or reject a flagged correction", reqBody("ReviewCorrectionRequest"), resp200("message"))},
		"/v1/memories/{id}/rollback":                    map[string]any{"post": endpoint("Memories", "Rollback memory", "Rollback to a previous version", reqBody("RollbackRequest"), resp200("MemoryResponse"))},

		// Interception
		"/v1/intercept": map[string]any{"post": endpoint("Interception", "Intercept prompt", "Analyze prompt and inject relevant context", reqBody("InterceptRequest"), resp200("InterceptResponse"))},

		// Relationships
		"/v1/relationships":                          map[string]any{"get": endpoint("Relationships", "List relationships", "List all memory relationships", nil, resp200("MemoryRelationship[]")), "post": endpoint("Relationships", "Create relationship", "Create a relationship between memories", reqBody("CreateRelationshipRequest"), resp201("MemoryRelationship"))},
		"/v1/relationships/bidirectional":            map[string]any{"post": endpoint("Relationships", "Create bidirectional", "Create relationships in both directions", reqBody("CreateBidirectionalRequest"), resp201("message"))},
		"/v1/relationships/from/{memoryId}":          map[string]any{"get": endpointWithParams("Relationships", "Get outgoing", "Get outgoing relationships from a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("MemoryRelationship[]"))},
		"/v1/relationships/to/{memoryId}":            map[string]any{"get": endpointWithParams("Relationships", "Get incoming", "Get incoming relationships to a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("MemoryRelationship[]"))},
		"/v1/relationships/between":                  map[string]any{"get": endpointWithParams("Relationships", "Get between", "Get relationship between two memories", []map[string]any{queryParam("from", "string", "From memory ID"), queryParam("to", "string", "To memory ID")}, resp200("MemoryRelationship")), "delete": endpointWithParams("Relationships", "Delete between", "Delete relationship between two memories", []map[string]any{queryParam("from", "string", "From memory ID"), queryParam("to", "string", "To memory ID")}, resp200("message"))},
		"/v1/relationships/{memoryId}/related":       map[string]any{"get": endpointWithParams("Relationships", "Get related", "Get related memories with min strength", []map[string]any{pathParam("memoryId", "Memory ID"), queryParam("minStrength", "number", "Minimum strength")}, resp200("MemoryRelationship[]"))},
		"/v1/relationships/{relationshipId}/strength": map[string]any{"put": endpoint("Relationships", "Update strength", "Update relationship strength", reqBody("UpdateStrengthRequest"), resp200("MemoryRelationship"))},
		"/v1/relationships/{memoryId}":               map[string]any{"delete": endpointWithParams("Relationships", "Delete all", "Delete all relationships for a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("message"))},
		"/v1/relationships/{memoryId}/suggest":        map[string]any{"post": endpointWithParams("Relationships", "Suggest relationships", "Auto-detect relationships using LLM", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("message"))},

		// Entity Graph
		"/v1/entity-graph/memory/{memoryId}/entities":      map[string]any{"get": endpointWithParams("Entity Graph", "Entities by memory", "Get entities for a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("Entity[]"))},
		"/v1/entity-graph/memory/{memoryId}/relationships": map[string]any{"get": endpointWithParams("Entity Graph", "Relationships by memory", "Get entity relationships for a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("EntityRelationship[]"))},
		"/v1/entity-graph/search":                          map[string]any{"get": endpointWithParams("Entity Graph", "Search entities", "Search entities by name", []map[string]any{queryParam("q", "string", "Search query")}, resp200("Entity[]"))},
		"/v1/entity-graph/knowledge-graph":                 map[string]any{"get": endpointWithParams("Entity Graph", "Knowledge graph", "Get the full knowledge graph", []map[string]any{queryParam("limit", "integer", "Max nodes")}, resp200("KnowledgeGraphResponse"))},
		"/v1/entity-graph/extract/{memoryId}":              map[string]any{"post": endpointWithParams("Entity Graph", "Extract entities", "Extract entities from a memory using LLM", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("message"))},
		"/v1/entity-graph/extract-batch":                   map[string]any{"post": endpoint("Entity Graph", "Batch extract", "Extract entities from multiple memories", reqBody("BatchExtractRequest"), resp200("message"))},

		// Audit
		"/v1/audit/logs":                           map[string]any{"get": endpointWithParams("Audit", "List audit logs", "List audit logs for current tenant", []map[string]any{queryParam("limit", "integer", "Max results")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/by-event/{eventType}":      map[string]any{"get": endpointWithParams("Audit", "By event type", "Filter audit logs by event type", []map[string]any{pathParam("eventType", "Event type")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/by-user/{userId}":          map[string]any{"get": endpointWithParams("Audit", "By user", "Filter audit logs by user", []map[string]any{pathParam("userId", "User ID")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/by-session/{sessionId}":    map[string]any{"get": endpointWithParams("Audit", "By session", "Filter audit logs by session", []map[string]any{pathParam("sessionId", "Session ID")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/recent":                    map[string]any{"get": endpointWithParams("Audit", "Recent logs", "Get most recent audit logs", []map[string]any{queryParam("limit", "integer", "Max results")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/by-date-range":             map[string]any{"get": endpointWithParams("Audit", "By date range", "Filter audit logs by date range", []map[string]any{queryParam("from", "string", "Start date RFC3339"), queryParam("to", "string", "End date RFC3339")}, resp200("AuditLog[]"))},
		"/v1/audit/logs/stats":                     map[string]any{"get": endpoint("Audit", "Audit stats", "Get audit log statistics", nil, resp200("AuditStats"))},
		"/v1/audit/memory/{memoryId}/history":      map[string]any{"get": endpointWithParams("Audit", "Memory history", "Get audit history for a specific memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("AuditLog[]"))},

		// Stats
		"/v1/stats/overview":     map[string]any{"get": endpoint("Stats", "System overview", "Get system statistics overview", nil, resp200("StatsResponse"))},
		"/v1/stats/top-patterns": map[string]any{"get": endpoint("Stats", "Top patterns", "Get top memory patterns", nil, resp200("TopPatternsResponse"))},
		"/v1/stats/health":       map[string]any{"get": endpoint("Stats", "Health stats", "Get system health statistics", nil, resp200("HealthStats"))},

		// Notes
		"/v1/notes":                                 map[string]any{"get": endpoint("Notes", "List notes", "List all notes", nil, resp200("Note[]"))},
		"/v1/notes/analyze":                         map[string]any{"post": endpoint("Notes", "Analyze session", "Analyze session to extract insights", reqBody("SessionAnalysisRequest"), resp200("SessionAnalysisResponse"))},
		"/v1/notes/hindsight":                       map[string]any{"get": endpoint("Notes", "List hindsight", "List hindsight notes", nil, resp200("HindsightNote[]")), "post": endpoint("Notes", "Create hindsight", "Create a manual hindsight note", reqBody("CreateHindsightNoteRequest"), resp201("HindsightNote"))},
		"/v1/notes/session/{sessionId}":             map[string]any{"get": endpointWithParams("Notes", "Session notes", "Get notes for a session", []map[string]any{pathParam("sessionId", "Session ID")}, resp200("Note[]"))},
		"/v1/notes/session/{sessionId}/hindsight":   map[string]any{"get": endpointWithParams("Notes", "Session hindsight", "Get hindsight notes for a session", []map[string]any{pathParam("sessionId", "Session ID")}, resp200("HindsightNote[]"))},

		// Compression
		"/v1/compression/compress":                    map[string]any{"post": endpoint("Compression", "Compress context", "Compress conversation context using LLM", reqBody("CompressionRequest"), resp200("CompressionResult"))},
		"/v1/compression/session/{sessionId}":         map[string]any{"get": endpointWithParams("Compression", "Session summaries", "Get compression summaries for a session", []map[string]any{pathParam("sessionId", "Session ID")}, resp200("ContextSummary[]"))},
		"/v1/compression/session/{sessionId}/latest":  map[string]any{"get": endpointWithParams("Compression", "Latest summary", "Get latest summary for a session", []map[string]any{pathParam("sessionId", "Session ID")}, resp200("ContextSummary"))},

		// Sessions
		"/v1/sessions":            map[string]any{"post": endpoint("Sessions", "Create session", "Start a new session", reqBody("CreateSessionRequest"), resp201("Session"))},
		"/v1/sessions/active":     map[string]any{"get": endpoint("Sessions", "List active", "List active sessions for tenant", nil, resp200("Session[]"))},
		"/v1/sessions/{id}":       map[string]any{"get": endpointWithParams("Sessions", "Get session", "Get session by ID", []map[string]any{pathParam("id", "Session ID")}, resp200("Session"))},
		"/v1/sessions/{id}/touch": map[string]any{"post": endpointWithParams("Sessions", "Touch session", "Update session activity timestamp", []map[string]any{pathParam("id", "Session ID")}, resp200("message"))},
		"/v1/sessions/{id}/end":   map[string]any{"post": endpointWithParams("Sessions", "End session", "End a session", []map[string]any{pathParam("id", "Session ID")}, resp200("message"))},

		// Batch
		"/v1/batch/import": map[string]any{"post": endpoint("Batch", "Import memories", "Bulk import memories", reqBody("BatchImportRequest"), resp200("BatchImportResponse"))},
		"/v1/batch/export": map[string]any{"get": endpoint("Batch", "Export memories", "Export all memories", nil, resp200("Memory[]"))},

		// Conflicts
		"/v1/conflicts/detect/{memoryId}": map[string]any{"post": endpointWithParams("Conflicts", "Detect conflicts", "Detect conflicts for a memory", []map[string]any{pathParam("memoryId", "Memory ID")}, resp200("ConflictResult[]"))},
		"/v1/conflicts/scan":              map[string]any{"post": endpoint("Conflicts", "Scan all", "Scan all memories for conflicts", nil, resp200("ConflictResult[]"))},

		// Webhooks
		"/v1/webhooks":                  map[string]any{"get": endpoint("Webhooks", "List webhooks", "List registered webhooks", nil, resp200("Webhook[]")), "post": endpoint("Webhooks", "Register webhook", "Register a new webhook", reqBody("RegisterWebhookRequest"), resp201("Webhook"))},
		"/v1/webhooks/{id}":             map[string]any{"delete": endpointWithParams("Webhooks", "Unregister webhook", "Remove a webhook", []map[string]any{pathParam("id", "Webhook ID")}, resp200("message"))},
		"/v1/webhooks/{id}/deliveries":  map[string]any{"get": endpointWithParams("Webhooks", "Delivery history", "Get webhook delivery history", []map[string]any{pathParam("id", "Webhook ID")}, resp200("WebhookDelivery[]"))},

		// MCP
		"/v1/mcp/message": map[string]any{"post": endpoint("MCP", "MCP message", "Send JSON-RPC 2.0 message to MCP server", reqBody("JSONRPCRequest"), resp200("JSONRPCResponse"))},
		"/v1/mcp/sse":     map[string]any{"post": endpoint("MCP", "MCP SSE", "Server-Sent Events transport for MCP", nil, resp200("SSE stream"))},
		"/v1/mcp/batch":   map[string]any{"post": endpoint("MCP", "MCP batch", "Batch JSON-RPC messages", reqBody("JSONRPCRequest[]"), resp200("JSONRPCResponse[]"))},

		// Health
		"/health": map[string]any{"get": map[string]any{"summary": "Health check", "operationId": "health", "responses": map[string]any{"200": map[string]any{"description": "OK"}}}},
	}
}

func endpoint(tag, summary, description string, requestBody, response map[string]any) map[string]any {
	ep := map[string]any{
		"tags":        []string{tag},
		"summary":     summary,
		"description": description,
		"security":    []map[string]any{{"BearerAuth": []string{}}},
		"responses":   map[string]any{"200": response, "400": map[string]any{"description": "Bad Request"}, "500": map[string]any{"description": "Internal Server Error"}},
	}
	if requestBody != nil {
		ep["requestBody"] = requestBody
	}
	return ep
}

func endpointWithParams(tag, summary, description string, params []map[string]any, response map[string]any) map[string]any {
	ep := endpoint(tag, summary, description, nil, response)
	ep["parameters"] = params
	return ep
}

func reqBody(schemaRef string) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": map[string]string{"$ref": "#/components/schemas/" + schemaRef},
			},
		},
	}
}

func resp200(schemaRef string) map[string]any {
	return map[string]any{"description": "Success", "content": map[string]any{
		"application/json": map[string]any{"schema": map[string]string{"$ref": "#/components/schemas/" + schemaRef}},
	}}
}

func resp201(schemaRef string) map[string]any {
	return map[string]any{"description": "Created", "content": map[string]any{
		"application/json": map[string]any{"schema": map[string]string{"$ref": "#/components/schemas/" + schemaRef}},
	}}
}

func queryParam(name, typ, desc string) map[string]any {
	return map[string]any{"name": name, "in": "query", "schema": map[string]string{"type": typ}, "description": desc}
}

func pathParam(name, desc string) map[string]any {
	return map[string]any{"name": name, "in": "path", "required": true, "schema": map[string]string{"type": "string"}, "description": desc}
}
