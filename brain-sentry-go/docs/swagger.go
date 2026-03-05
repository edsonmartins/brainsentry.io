// Package docs provides Swagger/OpenAPI documentation.
//
//	@title			Brain Sentry API
//	@version		1.0
//	@description	Brain Sentry - AI Agent Memory System. Manages memories, context injection, entity graphs, notes, and MCP protocol for AI agents.
//
//	@contact.name	Integral Tech
//	@contact.email	support@integraltech.com
//
//	@host		localhost:8080
//	@BasePath	/api
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Bearer token. Format: "Bearer {token}"
//
//	@tag.name			Auth
//	@tag.description	Authentication endpoints
//
//	@tag.name			Memories
//	@tag.description	Memory CRUD and search
//
//	@tag.name			Interception
//	@tag.description	Prompt interception and context injection
//
//	@tag.name			Relationships
//	@tag.description	Memory relationship management
//
//	@tag.name			Entity Graph
//	@tag.description	FalkorDB entity graph operations
//
//	@tag.name			Notes
//	@tag.description	Session analysis and note-taking
//
//	@tag.name			Compression
//	@tag.description	Context compression and summarization
//
//	@tag.name			MCP
//	@tag.description	Model Context Protocol (JSON-RPC 2.0)
//
//	@tag.name			Audit
//	@tag.description	Audit log queries
//
//	@tag.name			Stats
//	@tag.description	System statistics
//
//	@tag.name			Users
//	@tag.description	User management
//
//	@tag.name			Tenants
//	@tag.description	Tenant management
package docs
