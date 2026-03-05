package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestServer_HandleInitialize(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2024-11-05"}`),
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

	if result["protocolVersion"] != ProtocolVersion {
		t.Errorf("expected protocol version %s, got %v", ProtocolVersion, result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("expected serverInfo to be a map")
	}
	if serverInfo["name"] != ServerName {
		t.Errorf("expected server name %s, got %v", ServerName, serverInfo["name"])
	}
}

func TestServer_HandlePing(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "ping",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestServer_HandleToolsList(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/list",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("expected tools to be an array")
	}

	if len(tools) != 6 {
		t.Errorf("expected 6 tools, got %d", len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if ok {
			toolNames[toolMap["name"].(string)] = true
		}
	}

	expectedTools := []string{"create_memory", "search_memories", "get_memory", "list_memories", "update_memory", "delete_memory"}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("expected tool '%s' not found", name)
		}
	}
}

func TestServer_HandleResourcesList(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "resources/list",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}

	resources, ok := result["resources"].([]any)
	if !ok {
		t.Fatal("expected resources to be an array")
	}

	if len(resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(resources))
	}
}

func TestServer_HandlePromptsList(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "prompts/list",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be a map")
	}

	prompts, ok := result["prompts"].([]any)
	if !ok {
		t.Fatal("expected prompts to be an array")
	}

	if len(prompts) != 7 {
		t.Errorf("expected 7 prompts, got %d", len(prompts))
	}
}

func TestServer_HandleUnknownMethod(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "unknown/method",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

func TestServer_HandleInvalidJSON(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	respData := server.HandleMessage(context.Background(), []byte("not json"))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected parse error")
	}
	if resp.Error.Code != -32700 {
		t.Errorf("expected error code -32700, got %d", resp.Error.Code)
	}
}

func TestServer_HandleInitializedNotification(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialized",
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	// Notifications should return nil
	if respData != nil {
		t.Error("expected nil response for notification")
	}
}

func TestServer_HandleToolCallUnknown(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "nonexistent_tool", "arguments": {}}`),
	}

	data, _ := json.Marshal(req)
	respData := server.HandleMessage(context.Background(), data)

	var resp JSONRPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
}
