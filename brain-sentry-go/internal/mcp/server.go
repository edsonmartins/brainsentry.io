package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/integraltech/brainsentry/internal/service"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "brainsentry"
	ServerVersion   = "1.0.0"
)

// Server implements the MCP (Model Context Protocol) server.
type Server struct {
	memoryService       *service.MemoryService
	noteService         *service.NoteTakingService
	compService         *service.CompressionService
	interceptionService *service.InterceptionService
	tools               map[string]Tool
	resources           map[string]Resource
	prompts             map[string]Prompt
	mu                  sync.RWMutex
}

// NewServer creates a new MCP server.
func NewServer(memoryService *service.MemoryService, noteService *service.NoteTakingService, compService *service.CompressionService, interceptionService *service.InterceptionService) *Server {
	s := &Server{
		memoryService:       memoryService,
		noteService:         noteService,
		compService:         compService,
		interceptionService: interceptionService,
		tools:               make(map[string]Tool),
		resources:           make(map[string]Resource),
		prompts:             make(map[string]Prompt),
	}
	s.registerTools()
	s.registerResources()
	s.registerPrompts()
	return s
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Result  any         `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Handler     func(ctx context.Context, params json.RawMessage) (any, error)
}

// Resource represents an MCP resource definition.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Handler     func(ctx context.Context) (any, error)
}

// Prompt represents an MCP prompt definition.
type Prompt struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Arguments   []PromptArg     `json:"arguments,omitempty"`
	Handler     func(ctx context.Context, args map[string]string) ([]PromptMessage, error)
}

// PromptArg represents a prompt argument.
type PromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptMessage represents a message in a prompt response.
type PromptMessage struct {
	Role    string         `json:"role"`
	Content PromptContent  `json:"content"`
}

// PromptContent represents content in a prompt message.
type PromptContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// HandleMessage processes a single JSON-RPC message and returns a response.
func (s *Server) HandleMessage(ctx context.Context, data []byte) []byte {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return s.errorResponse(nil, -32700, "Parse error")
	}

	resp := s.dispatch(ctx, &req)
	if resp == nil {
		return nil // notification, no response needed
	}

	out, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal response", "error", err)
		return s.errorResponse(req.ID, -32603, "Internal error")
	}
	return out
}

func (s *Server) dispatch(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return nil // notification
	case "ping":
		return s.successResponse(req.ID, map[string]string{})
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(ctx, req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(ctx, req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	return s.successResponse(req.ID, map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities": map[string]any{
			"tools":     map[string]any{"listChanged": false},
			"resources": map[string]any{"subscribe": false, "listChanged": false},
			"prompts":   map[string]any{"listChanged": false},
		},
		"serverInfo": map[string]string{
			"name":    ServerName,
			"version": ServerVersion,
		},
	})
}

func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return s.successResponse(req.ID, map[string]any{"tools": tools})
}

func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}

	s.mu.RLock()
	tool, ok := s.tools[params.Name]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)}}
	}

	result, err := tool.Handler(ctx, params.Arguments)
	if err != nil {
		// Return structured error per MCP_SERVER_API.md spec
		errResp := toolErrorMap(ErrCategoryInternal, "internal", err.Error())
		text, _ := json.Marshal(errResp)
		return s.successResponse(req.ID, map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(text)}},
			"isError": true,
		})
	}

	text, _ := json.Marshal(result)
	return s.successResponse(req.ID, map[string]any{
		"content": []map[string]any{{"type": "text", "text": string(text)}},
	})
}

func (s *Server) handleResourcesList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]map[string]string, 0, len(s.resources))
	for _, r := range s.resources {
		res := map[string]string{
			"uri":  r.URI,
			"name": r.Name,
		}
		if r.Description != "" {
			res["description"] = r.Description
		}
		if r.MimeType != "" {
			res["mimeType"] = r.MimeType
		}
		resources = append(resources, res)
	}
	return s.successResponse(req.ID, map[string]any{"resources": resources})
}

func (s *Server) handleResourcesRead(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}

	s.mu.RLock()
	resource, ok := s.resources[params.URI]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: fmt.Sprintf("Unknown resource: %s", params.URI)}}
	}

	result, err := resource.Handler(ctx)
	if err != nil {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32603, Message: err.Error()}}
	}

	text, _ := json.Marshal(result)
	return s.successResponse(req.ID, map[string]any{
		"contents": []map[string]string{
			{"uri": params.URI, "mimeType": "application/json", "text": string(text)},
		},
	})
}

func (s *Server) handlePromptsList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]map[string]any, 0, len(s.prompts))
	for _, p := range s.prompts {
		prompt := map[string]any{
			"name": p.Name,
		}
		if p.Description != "" {
			prompt["description"] = p.Description
		}
		if len(p.Arguments) > 0 {
			prompt["arguments"] = p.Arguments
		}
		prompts = append(prompts, prompt)
	}
	return s.successResponse(req.ID, map[string]any{"prompts": prompts})
}

func (s *Server) handlePromptsGet(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}

	s.mu.RLock()
	prompt, ok := s.prompts[params.Name]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: fmt.Sprintf("Unknown prompt: %s", params.Name)}}
	}

	messages, err := prompt.Handler(ctx, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32603, Message: err.Error()}}
	}

	return s.successResponse(req.ID, map[string]any{
		"description": prompt.Description,
		"messages":    messages,
	})
}

func (s *Server) successResponse(id any, result any) *JSONRPCResponse {
	return &JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) errorResponse(id any, code int, message string) []byte {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	out, _ := json.Marshal(resp)
	return out
}
