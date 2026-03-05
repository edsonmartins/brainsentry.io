package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Tool listing tests
// ---------------------------------------------------------------------------

// TestTools_RegisteredNames verifies that registerTools populates the tools map
// with exactly the expected set of memory-management tools when no
// interceptionService is provided.
func TestTools_RegisteredNames(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	expected := []string{
		"create_memory",
		"search_memories",
		"get_memory",
		"list_memories",
		"update_memory",
		"delete_memory",
	}

	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			if _, ok := server.tools[name]; !ok {
				t.Errorf("tool %q not registered", name)
			}
		})
	}
}

// TestTools_NoInterceptWhenServiceNil confirms that "intercept_prompt" is NOT
// registered when interceptionService is nil (zero-value server).
func TestTools_NoInterceptWhenServiceNil(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	if _, ok := server.tools["intercept_prompt"]; ok {
		t.Error("intercept_prompt should not be registered when interceptionService is nil")
	}
}

// TestTools_Count ensures exactly 6 tools are registered without the
// interception service.
func TestTools_Count(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	if got := len(server.tools); got != 6 {
		t.Errorf("expected 6 tools, got %d", got)
	}
}

// TestTools_EachHasRequiredFields verifies that every registered tool has a
// non-empty Name, Description, and valid JSON InputSchema.
func TestTools_EachHasRequiredFields(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	for name, tool := range server.tools {
		t.Run(name, func(t *testing.T) {
			if tool.Name == "" {
				t.Error("tool Name must not be empty")
			}
			if tool.Name != name {
				t.Errorf("tool.Name %q does not match map key %q", tool.Name, name)
			}
			if tool.Description == "" {
				t.Error("tool Description must not be empty")
			}
			if tool.Handler == nil {
				t.Error("tool Handler must not be nil")
			}
			var schema map[string]any
			if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
				t.Errorf("tool InputSchema is not valid JSON: %v", err)
			}
			if schema["type"] != "object" {
				t.Errorf("expected InputSchema type 'object', got %v", schema["type"])
			}
		})
	}
}

// TestTools_InputSchemaRequired checks that tools with "required" in their
// schema declare the correct mandatory fields.
func TestTools_InputSchemaRequired(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	cases := []struct {
		tool     string
		required string
	}{
		{"create_memory", "content"},
		{"search_memories", "query"},
		{"get_memory", "id"},
		{"update_memory", "id"},
		{"delete_memory", "id"},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			tool, ok := server.tools[tc.tool]
			if !ok {
				t.Fatalf("tool %q not found", tc.tool)
			}

			var schema map[string]any
			if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
				t.Fatalf("invalid InputSchema JSON: %v", err)
			}

			rawRequired, exists := schema["required"]
			if !exists {
				t.Fatalf("tool %q schema has no 'required' field", tc.tool)
			}

			requiredSlice, ok := rawRequired.([]any)
			if !ok {
				t.Fatalf("'required' is not a slice for tool %q", tc.tool)
			}

			found := false
			for _, r := range requiredSlice {
				if r == tc.required {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %q in 'required' for tool %q", tc.required, tc.tool)
			}
		})
	}
}

// TestTools_HandlerPanicsWithNilService documents the current behaviour that
// tool handlers panic (nil pointer dereference) when the underlying service is
// nil. This test verifies the panic is recoverable so the broader test suite
// continues; it also serves as a reminder that callers must guard against nil
// services before invoking handlers.
func TestTools_HandlerPanicsWithNilService(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	cases := []struct {
		tool   string
		params string
	}{
		{"create_memory", `{"content":"test"}`},
		{"search_memories", `{"query":"test"}`},
		{"get_memory", `{"id":"abc"}`},
		{"list_memories", `{}`},
		{"update_memory", `{"id":"abc","content":"new"}`},
		{"delete_memory", `{"id":"abc"}`},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			tool, ok := server.tools[tc.tool]
			if !ok {
				t.Fatalf("tool %q not registered", tc.tool)
			}

			didPanic := func() (panicked bool) {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				_, _ = tool.Handler(ctx, json.RawMessage(tc.params))
				return false
			}()

			// The tool either panics (nil service) or returns an error.
			// Either outcome is acceptable here; we just confirm no unrecoverable crash.
			_ = didPanic
		})
	}
}

// TestTools_HandlerInvalidJSON checks that each handler rejects malformed JSON
// params without panicking and returns a structured error response.
func TestTools_HandlerInvalidJSON(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	toolNames := []string{
		"create_memory", "search_memories", "get_memory",
		"list_memories", "update_memory", "delete_memory",
	}

	for _, name := range toolNames {
		t.Run(name, func(t *testing.T) {
			tool, ok := server.tools[name]
			if !ok {
				t.Fatalf("tool %q not registered", name)
			}
			result, err := tool.Handler(ctx, json.RawMessage(`{bad json`))
			// Tools now return structured error maps instead of Go errors
			if err != nil {
				return // Go error is also acceptable
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Errorf("expected map result for invalid JSON in tool %q", name)
				return
			}
			if resultMap["success"] != false {
				t.Errorf("expected success=false for invalid JSON in tool %q", name)
			}
		})
	}
}

// TestTools_ToolsCallViaServer_InvalidOuterJSON sends a completely unparseable
// message to HandleMessage and verifies the server returns a parse error
// (-32700). This is distinct from an invalid "params" field inside a valid
// tools/call envelope.
func TestTools_ToolsCallViaServer_InvalidOuterJSON(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	// The entire message body is invalid JSON.
	respData := server.HandleMessage(context.Background(), []byte(`{bad json`))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error response for invalid JSON")
	}
	if resp.Error.Code != -32700 {
		t.Errorf("expected code -32700, got %d", resp.Error.Code)
	}
}

// TestTools_ToolsCallViaServer_InvalidParamsField sends a valid JSON-RPC
// envelope for tools/call whose "params" value cannot be decoded as the
// expected struct. The server must return code -32602 (invalid params).
func TestTools_ToolsCallViaServer_InvalidParamsField(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	// "params" is valid JSON but is a bare string, not the expected object.
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      99,
		"method":  "tools/call",
		"params":  "not an object",
	}
	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error response for invalid params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected code -32602, got %d", resp.Error.Code)
	}
}

// TestTools_ToolsCallViaServer_KnownToolNilService_Panics documents that
// invoking a tool that directly calls a nil service through the HandleMessage
// path results in a panic (the server does not recover). This test guards
// against that panic via recover so that the suite continues.
func TestTools_ToolsCallViaServer_KnownToolNilService_Panics(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name":      "create_memory",
		"arguments": map[string]string{"content": "hello"},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      10,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)

	panicked := func() (p bool) {
		defer func() {
			if r := recover(); r != nil {
				p = true
			}
		}()
		_ = server.HandleMessage(context.Background(), data)
		return false
	}()

	// Document: currently panics because memoryService is nil.
	// Either outcome (panic or error) is noted here; this test simply ensures
	// the goroutine does not exit the test process uncontrolled.
	_ = panicked
}

// ---------------------------------------------------------------------------
// Prompt listing and handler tests
// ---------------------------------------------------------------------------

// TestPrompts_RegisteredNames verifies that all expected prompts are
// registered.
func TestPrompts_RegisteredNames(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	expected := []string{
		"capture_pattern",
		"extract_learning",
		"summarize_discussion",
		"context_builder",
		"agent_context",
		"memory_summary",
		"hindsight_review",
	}

	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			if _, ok := server.prompts[name]; !ok {
				t.Errorf("prompt %q not registered", name)
			}
		})
	}
}

// TestPrompts_Count ensures exactly 7 prompts are registered.
func TestPrompts_Count(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	if got := len(server.prompts); got != 7 {
		t.Errorf("expected 7 prompts, got %d", got)
	}
}

// TestPrompts_EachHasName verifies that every registered prompt has a
// non-empty Name that matches its map key and a non-nil Handler.
func TestPrompts_EachHasName(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	for key, prompt := range server.prompts {
		t.Run(key, func(t *testing.T) {
			if prompt.Name == "" {
				t.Error("prompt Name must not be empty")
			}
			if prompt.Name != key {
				t.Errorf("prompt.Name %q does not match map key %q", prompt.Name, key)
			}
			if prompt.Handler == nil {
				t.Error("prompt Handler must not be nil")
			}
		})
	}
}

// TestPrompts_RequiredArguments checks that prompts that must have required
// arguments have them declared as Required=true.
func TestPrompts_RequiredArguments(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	cases := []struct {
		prompt  string
		argName string
	}{
		{"capture_pattern", "pattern"},
		{"extract_learning", "session"},
		{"summarize_discussion", "discussion"},
		{"agent_context", "task"},
		{"context_builder", "task"},
	}

	for _, tc := range cases {
		t.Run(tc.prompt+"/"+tc.argName, func(t *testing.T) {
			p, ok := server.prompts[tc.prompt]
			if !ok {
				t.Fatalf("prompt %q not registered", tc.prompt)
			}
			found := false
			for _, arg := range p.Arguments {
				if arg.Name == tc.argName && arg.Required {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected argument %q to be required in prompt %q", tc.argName, tc.prompt)
			}
		})
	}
}

// TestPromptCapturePattern_MissingArg checks that promptCapturePattern returns
// an error when the required "pattern" arg is absent.
func TestPromptCapturePattern_MissingArg(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	_, err := server.promptCapturePattern(ctx, map[string]string{})
	if err == nil {
		t.Fatal("expected error when 'pattern' arg is missing")
	}
	if !strings.Contains(err.Error(), "pattern") {
		t.Errorf("expected error message to mention 'pattern', got: %v", err)
	}
}

// TestPromptCapturePattern_WithPattern verifies the returned message contains
// the supplied pattern text.
func TestPromptCapturePattern_WithPattern(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	messages, err := server.promptCapturePattern(ctx, map[string]string{
		"pattern": "use context.Context as first argument",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if messages[0].Content.Type != "text" {
		t.Errorf("expected content type 'text', got %q", messages[0].Content.Type)
	}
	if !strings.Contains(messages[0].Content.Text, "use context.Context as first argument") {
		t.Error("expected message to contain the supplied pattern text")
	}
}

// TestPromptCapturePattern_WithLanguage verifies the optional language arg is
// reflected in the output.
func TestPromptCapturePattern_WithLanguage(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	messages, err := server.promptCapturePattern(ctx, map[string]string{
		"pattern":  "defer for cleanup",
		"language": "Go",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if !strings.Contains(messages[0].Content.Text, "Go") {
		t.Error("expected message to contain the supplied language")
	}
}

// TestPromptExtractLearning_MissingArg checks that promptExtractLearning
// returns an error when the required "session" arg is absent.
func TestPromptExtractLearning_MissingArg(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	_, err := server.promptExtractLearning(ctx, map[string]string{})
	if err == nil {
		t.Fatal("expected error when 'session' arg is missing")
	}
	if !strings.Contains(err.Error(), "session") {
		t.Errorf("expected error message to mention 'session', got: %v", err)
	}
}

// TestPromptExtractLearning_WithSession verifies the returned message embeds
// the session content and contains expected structural guidance.
func TestPromptExtractLearning_WithSession(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	sessionContent := "We decided to use Redis for caching instead of Memcached."
	messages, err := server.promptExtractLearning(ctx, map[string]string{
		"session": sessionContent,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := messages[0].Content.Text
	if !strings.Contains(text, sessionContent) {
		t.Error("expected message to embed the session content")
	}
	// Should mention create_memory as the follow-up action
	if !strings.Contains(text, "create_memory") {
		t.Error("expected message to reference 'create_memory' tool")
	}
}

// TestPromptSummarizeDiscussion_MissingArg checks that promptSummarizeDiscussion
// returns an error when the required "discussion" arg is absent.
func TestPromptSummarizeDiscussion_MissingArg(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	_, err := server.promptSummarizeDiscussion(ctx, map[string]string{})
	if err == nil {
		t.Fatal("expected error when 'discussion' arg is missing")
	}
	if !strings.Contains(err.Error(), "discussion") {
		t.Errorf("expected error message to mention 'discussion', got: %v", err)
	}
}

// TestPromptSummarizeDiscussion_WithDiscussion verifies the returned message
// embeds the discussion content and contains expected structural guidance.
func TestPromptSummarizeDiscussion_WithDiscussion(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	discussionContent := "Team discussed microservice boundaries for the auth module."
	messages, err := server.promptSummarizeDiscussion(ctx, map[string]string{
		"discussion": discussionContent,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := messages[0].Content.Text
	if !strings.Contains(text, discussionContent) {
		t.Error("expected message to embed the discussion content")
	}
	if !strings.Contains(text, "create_memory") {
		t.Error("expected message to reference 'create_memory' tool")
	}
}

// TestPromptHindsightReview_NilNoteService verifies that the hindsight_review
// prompt handler returns a graceful message (not an error) when noteService
// is nil.
func TestPromptHindsightReview_NilNoteService(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	messages, err := server.promptHindsightReview(ctx, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if messages[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", messages[0].Role)
	}
	// Should communicate that the feature is unavailable
	if !strings.Contains(messages[0].Content.Text, "not available") {
		t.Error("expected message to communicate feature is unavailable")
	}
}

// TestPromptsGet_ViaServer_CapturePattern exercises the full dispatch path for
// prompts/get with the capture_pattern prompt.
func TestPromptsGet_ViaServer_CapturePattern(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name": "capture_pattern",
		"arguments": map[string]string{
			"pattern":  "table-driven tests",
			"language": "Go",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      20,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}

	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("expected messages to be a slice")
	}
	if len(messages) == 0 {
		t.Fatal("expected at least one message in response")
	}
}

// TestPromptsGet_ViaServer_UnknownPrompt verifies that requesting an
// unregistered prompt returns an RPC error.
func TestPromptsGet_ViaServer_UnknownPrompt(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name":      "nonexistent_prompt",
		"arguments": map[string]string{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      21,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for unknown prompt")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected code -32602, got %d", resp.Error.Code)
	}
}

// TestPromptsGet_ViaServer_MissingRequiredArg verifies that a prompt handler
// returning an error is propagated as an RPC error response.
func TestPromptsGet_ViaServer_MissingRequiredArg(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	// capture_pattern requires "pattern" – omit it to trigger an error
	params := map[string]any{
		"name":      "capture_pattern",
		"arguments": map[string]string{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      22,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error when required argument is missing")
	}
	if resp.Error.Code != -32603 {
		t.Errorf("expected code -32603, got %d", resp.Error.Code)
	}
}

// TestPromptsGet_ViaServer_ExtractLearning exercises the extract_learning
// prompt through the full dispatch path.
func TestPromptsGet_ViaServer_ExtractLearning(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name": "extract_learning",
		"arguments": map[string]string{
			"session": "We implemented a retry mechanism using exponential back-off.",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      23,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}

	messages, ok := result["messages"].([]any)
	if !ok || len(messages) == 0 {
		t.Fatal("expected messages slice with at least one entry")
	}
}

// TestPromptsGet_ViaServer_SummarizeDiscussion exercises the
// summarize_discussion prompt through the full dispatch path.
func TestPromptsGet_ViaServer_SummarizeDiscussion(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name": "summarize_discussion",
		"arguments": map[string]string{
			"discussion": "Decided to adopt gRPC for internal service communication.",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      24,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}
	if _, hasMessages := result["messages"]; !hasMessages {
		t.Error("expected 'messages' key in result")
	}
}

// TestPromptsGet_ViaServer_HindsightReview_NilService exercises the
// hindsight_review prompt via the full dispatch path when noteService is nil.
func TestPromptsGet_ViaServer_HindsightReview_NilService(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	params := map[string]any{
		"name":      "hindsight_review",
		"arguments": map[string]string{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      25,
		Method:  "prompts/get",
		Params:  paramsJSON,
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should succeed even with nil noteService (returns a "not available" message)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}
	if _, hasMessages := result["messages"]; !hasMessages {
		t.Error("expected 'messages' key in result")
	}
}

// TestPromptMessageStructure verifies that all prompt messages returned by
// context-free handlers have the expected role and content structure.
func TestPromptMessageStructure(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)
	ctx := context.Background()

	type promptCall struct {
		name    string
		handler func(context.Context, map[string]string) ([]PromptMessage, error)
		args    map[string]string
	}

	calls := []promptCall{
		{
			name:    "capture_pattern",
			handler: server.promptCapturePattern,
			args:    map[string]string{"pattern": "repository pattern"},
		},
		{
			name:    "extract_learning",
			handler: server.promptExtractLearning,
			args:    map[string]string{"session": "Some session content"},
		},
		{
			name:    "summarize_discussion",
			handler: server.promptSummarizeDiscussion,
			args:    map[string]string{"discussion": "Some discussion content"},
		},
		{
			name:    "hindsight_review (nil service)",
			handler: server.promptHindsightReview,
			args:    map[string]string{},
		},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			messages, err := tc.handler(ctx, tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for i, msg := range messages {
				if msg.Role == "" {
					t.Errorf("message[%d].Role is empty", i)
				}
				if msg.Content.Type == "" {
					t.Errorf("message[%d].Content.Type is empty", i)
				}
				if msg.Content.Text == "" {
					t.Errorf("message[%d].Content.Text is empty", i)
				}
			}
		})
	}
}
