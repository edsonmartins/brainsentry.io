package service

import (
	"context"
	"testing"
	"time"
)

func TestNewConnectorRegistry(t *testing.T) {
	reg := NewConnectorRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(reg.List()) != 0 {
		t.Error("expected empty registry")
	}
}

func TestConnectorRegistry_RegisterAndGet(t *testing.T) {
	reg := NewConnectorRegistry()
	gh := NewGitHubConnector(ConnectorConfig{
		Type: ConnectorGitHub,
		Name: "my-repo",
		Credentials: map[string]string{"token": "test"},
		Settings:    map[string]string{"repo": "owner/repo"},
	})

	reg.Register("github-main", gh)

	c, ok := reg.Get("github-main")
	if !ok {
		t.Fatal("expected to find connector")
	}
	if c.Type() != ConnectorGitHub {
		t.Errorf("expected github type, got %s", c.Type())
	}
	if c.Name() != "my-repo" {
		t.Errorf("expected my-repo, got %s", c.Name())
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent connector")
	}
}

func TestConnectorRegistry_List(t *testing.T) {
	reg := NewConnectorRegistry()
	reg.Register("a", NewGitHubConnector(ConnectorConfig{Name: "a"}))
	reg.Register("b", NewNotionConnector(ConnectorConfig{Name: "b"}))

	list := reg.List()
	if len(list) != 2 {
		t.Errorf("expected 2 connectors, got %d", len(list))
	}
}

func TestNewConnectorService(t *testing.T) {
	reg := NewConnectorRegistry()
	svc := NewConnectorService(reg, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.chunkSize != 500 {
		t.Errorf("expected chunk size 500, got %d", svc.chunkSize)
	}
	if svc.chunkOverlap != 50 {
		t.Errorf("expected chunk overlap 50, got %d", svc.chunkOverlap)
	}
}

func TestSyncConnector_NotFound(t *testing.T) {
	reg := NewConnectorRegistry()
	svc := NewConnectorService(reg, nil)

	_, err := svc.SyncConnector(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent connector")
	}
}

func TestGitHubConnector_Validate(t *testing.T) {
	// Missing token
	gh := NewGitHubConnector(ConnectorConfig{
		Settings: map[string]string{"repo": "owner/repo"},
	})
	if err := gh.Validate(); err == nil {
		t.Error("expected error for missing token")
	}

	// Missing repo
	gh = NewGitHubConnector(ConnectorConfig{
		Credentials: map[string]string{"token": "test"},
	})
	if err := gh.Validate(); err == nil {
		t.Error("expected error for missing repo")
	}

	// Valid
	gh = NewGitHubConnector(ConnectorConfig{
		Credentials: map[string]string{"token": "test"},
		Settings:    map[string]string{"repo": "owner/repo"},
	})
	if err := gh.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNotionConnector_Validate(t *testing.T) {
	nc := NewNotionConnector(ConnectorConfig{})
	if err := nc.Validate(); err == nil {
		t.Error("expected error for missing token")
	}

	nc = NewNotionConnector(ConnectorConfig{
		Credentials: map[string]string{"token": "test"},
	})
	if err := nc.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebCrawlerConnector_Validate(t *testing.T) {
	wc := NewWebCrawlerConnector(ConnectorConfig{})
	if err := wc.Validate(); err == nil {
		t.Error("expected error for missing URLs")
	}

	wc = NewWebCrawlerConnector(ConnectorConfig{
		Settings: map[string]string{"urls": "https://example.com"},
	})
	if err := wc.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGoogleDriveConnector_Validate(t *testing.T) {
	gd := NewGoogleDriveConnector(ConnectorConfig{})
	if err := gd.Validate(); err == nil {
		t.Error("expected error for missing token")
	}

	gd = NewGoogleDriveConnector(ConnectorConfig{
		Credentials: map[string]string{"token": "test"},
	})
	if err := gd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChunkText(t *testing.T) {
	svc := NewConnectorService(NewConnectorRegistry(), nil)
	svc.chunkSize = 50  // small for testing
	svc.chunkOverlap = 10

	// Generate text with enough words
	content := "word1 word2 word3 word4 word5 word6 word7 word8 word9 word10 " +
		"word11 word12 word13 word14 word15 word16 word17 word18 word19 word20 " +
		"word21 word22 word23 word24 word25 word26 word27 word28 word29 word30 " +
		"word31 word32 word33 word34 word35 word36 word37 word38 word39 word40 " +
		"word41 word42 word43 word44 word45 word46 word47 word48 word49 word50"

	chunks := svc.chunkText("doc1", content)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	if chunks[0].DocumentID != "doc1" {
		t.Error("expected doc1")
	}
	if chunks[0].Index != 0 {
		t.Error("expected index 0")
	}
}

func TestChunkCode(t *testing.T) {
	svc := NewConnectorService(NewConnectorRegistry(), nil)

	code := "func main() {\n\tfmt.Println(\"hello\")\n}\n\nfunc helper() {\n\t// help\n}"
	chunks := svc.chunkCode("doc2", code)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	if chunks[0].DocumentID != "doc2" {
		t.Error("expected doc2")
	}
}

func TestChunkDocument_Empty(t *testing.T) {
	svc := NewConnectorService(NewConnectorRegistry(), nil)
	doc := Document{ID: "d1", Content: ""}
	chunks := svc.chunkDocument(doc)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty doc, got %d", len(chunks))
	}
}

func TestChunkDocument_Code(t *testing.T) {
	svc := NewConnectorService(NewConnectorRegistry(), nil)
	doc := Document{
		ID:          "d1",
		Content:     "package main\n\nimport \"fmt\"\n\nfunc main() {}",
		ContentType: "code",
	}
	chunks := svc.chunkDocument(doc)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
}

func TestDocument_Structure(t *testing.T) {
	doc := Document{
		ID:         "d1",
		SourceType: ConnectorGitHub,
		Title:      "Test Doc",
		Content:    "content",
		Metadata:   map[string]string{"key": "value"},
		FetchedAt:  time.Now(),
	}
	if doc.SourceType != ConnectorGitHub {
		t.Error("expected github source")
	}
	if doc.Metadata["key"] != "value" {
		t.Error("expected key=value")
	}
}

func TestSyncResult_Structure(t *testing.T) {
	r := SyncResult{
		ConnectorType:    ConnectorNotion,
		DocumentsNew:     5,
		DocumentsUpdated: 2,
		ChunksCreated:    15,
	}
	if r.DocumentsNew != 5 {
		t.Error("expected 5 new docs")
	}
	if r.ChunksCreated != 15 {
		t.Error("expected 15 chunks")
	}
}

func TestConnectorStatus(t *testing.T) {
	gh := NewGitHubConnector(ConnectorConfig{Name: "test"})
	if gh.Status() != ConnectorStatusIdle {
		t.Errorf("expected idle status, got %s", gh.Status())
	}
}

func TestConnectorTypes(t *testing.T) {
	types := []ConnectorType{ConnectorGitHub, ConnectorNotion, ConnectorGoogleDrive, ConnectorWebCrawler}
	seen := make(map[ConnectorType]bool)
	for _, ct := range types {
		if seen[ct] {
			t.Errorf("duplicate connector type: %s", ct)
		}
		seen[ct] = true
	}
}
