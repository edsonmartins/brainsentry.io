package graph

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// EscapeCypher
// ---------------------------------------------------------------------------

func TestEscapeCypher_SingleQuote(t *testing.T) {
	got := EscapeCypher("it's a test")
	want := `it\'s a test`
	if got != want {
		t.Errorf("EscapeCypher single quote: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_Backslash(t *testing.T) {
	got := EscapeCypher(`path\to\file`)
	want := `path\\to\\file`
	if got != want {
		t.Errorf("EscapeCypher backslash: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_Newline(t *testing.T) {
	got := EscapeCypher("line1\nline2")
	want := `line1\nline2`
	if got != want {
		t.Errorf("EscapeCypher newline: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_CarriageReturn(t *testing.T) {
	got := EscapeCypher("line1\rline2")
	want := `line1\rline2`
	if got != want {
		t.Errorf("EscapeCypher carriage return: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_Tab(t *testing.T) {
	got := EscapeCypher("col1\tcol2")
	want := `col1\tcol2`
	if got != want {
		t.Errorf("EscapeCypher tab: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_NullByte(t *testing.T) {
	got := EscapeCypher("before\x00after")
	// Null bytes must be stripped entirely.
	if strings.Contains(got, "\x00") {
		t.Errorf("EscapeCypher null byte: result still contains null byte: %q", got)
	}
	want := "beforeafter"
	if got != want {
		t.Errorf("EscapeCypher null byte: got %q, want %q", got, want)
	}
}

func TestEscapeCypher_EmptyString(t *testing.T) {
	got := EscapeCypher("")
	if got != "" {
		t.Errorf("EscapeCypher empty: got %q, want %q", got, "")
	}
}

func TestEscapeCypher_NoSpecialChars(t *testing.T) {
	input := "hello world 123"
	got := EscapeCypher(input)
	if got != input {
		t.Errorf("EscapeCypher plain: got %q, want %q", got, input)
	}
}

func TestEscapeCypher_MultipleSpecialChars(t *testing.T) {
	// Input: it's complex\nwith tabs\t and null\x00bytes
	input := "it's complex\nwith tabs\t and null\x00bytes"
	got := EscapeCypher(input)

	// Single quote must be escaped as \' (backslash + apostrophe), so no
	// bare unescaped apostrophe should appear that is NOT preceded by \.
	// The simplest check: the result must contain \' and no literal newline/tab/null.
	if !strings.Contains(got, `\'`) {
		t.Error("EscapeCypher: expected escaped single quote \\' in result")
	}
	if strings.Contains(got, "\n") {
		t.Error("EscapeCypher: literal newline in result")
	}
	if strings.Contains(got, "\t") {
		t.Error("EscapeCypher: literal tab in result")
	}
	if strings.Contains(got, "\x00") {
		t.Error("EscapeCypher: null byte in result")
	}
}

func TestEscapeCypher_BackslashAndQuote(t *testing.T) {
	// A backslash immediately followed by a single quote.
	got := EscapeCypher(`\'`)
	want := `\\\'`
	if got != want {
		t.Errorf("EscapeCypher backslash+quote: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// EscapeCypherIdentifier
// ---------------------------------------------------------------------------

func TestEscapeCypherIdentifier_Plain(t *testing.T) {
	got := EscapeCypherIdentifier("MyLabel")
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier plain: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_WithBacktick(t *testing.T) {
	got := EscapeCypherIdentifier("My`Label")
	want := "`My``Label`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier backtick: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_WithNewline(t *testing.T) {
	got := EscapeCypherIdentifier("My\nLabel")
	// Newlines must be stripped.
	if strings.Contains(got, "\n") {
		t.Errorf("EscapeCypherIdentifier newline: literal newline in result: %q", got)
	}
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier newline: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_WithCarriageReturn(t *testing.T) {
	got := EscapeCypherIdentifier("My\rLabel")
	if strings.Contains(got, "\r") {
		t.Errorf("EscapeCypherIdentifier CR: literal CR in result: %q", got)
	}
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier CR: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_WithTab(t *testing.T) {
	got := EscapeCypherIdentifier("My\tLabel")
	if strings.Contains(got, "\t") {
		t.Errorf("EscapeCypherIdentifier tab: literal tab in result: %q", got)
	}
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier tab: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_WithNullByte(t *testing.T) {
	got := EscapeCypherIdentifier("My\x00Label")
	if strings.Contains(got, "\x00") {
		t.Errorf("EscapeCypherIdentifier null byte: null byte in result: %q", got)
	}
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier null byte: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_LeadingTrailingSpaces(t *testing.T) {
	got := EscapeCypherIdentifier("  MyLabel  ")
	want := "`MyLabel`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier trim spaces: got %q, want %q", got, want)
	}
}

func TestEscapeCypherIdentifier_EmptyString(t *testing.T) {
	got := EscapeCypherIdentifier("")
	// Must still be wrapped in backticks even if empty.
	if got != "``" {
		t.Errorf("EscapeCypherIdentifier empty: got %q, want %q", got, "``")
	}
}

func TestEscapeCypherIdentifier_MultipleBackticks(t *testing.T) {
	got := EscapeCypherIdentifier("a`b`c")
	want := "`a``b``c`"
	if got != want {
		t.Errorf("EscapeCypherIdentifier multiple backticks: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// GetString
// ---------------------------------------------------------------------------

func TestGetString_Present(t *testing.T) {
	m := map[string]any{"key": "value"}
	if got := GetString(m, "key"); got != "value" {
		t.Errorf("GetString present: got %q, want %q", got, "value")
	}
}

func TestGetString_Missing(t *testing.T) {
	m := map[string]any{}
	if got := GetString(m, "missing"); got != "" {
		t.Errorf("GetString missing: got %q, want %q", got, "")
	}
}

func TestGetString_WrongType(t *testing.T) {
	m := map[string]any{"key": 42}
	if got := GetString(m, "key"); got != "" {
		t.Errorf("GetString wrong type: got %q, want %q", got, "")
	}
}

func TestGetString_NilValue(t *testing.T) {
	m := map[string]any{"key": nil}
	if got := GetString(m, "key"); got != "" {
		t.Errorf("GetString nil value: got %q, want %q", got, "")
	}
}

func TestGetString_NilMap(t *testing.T) {
	var m map[string]any
	if got := GetString(m, "key"); got != "" {
		t.Errorf("GetString nil map: got %q, want %q", got, "")
	}
}

func TestGetString_EmptyString(t *testing.T) {
	m := map[string]any{"key": ""}
	if got := GetString(m, "key"); got != "" {
		t.Errorf("GetString empty string: got %q, want %q", got, "")
	}
}

// ---------------------------------------------------------------------------
// GetInt64
// ---------------------------------------------------------------------------

func TestGetInt64_PresentInt64(t *testing.T) {
	m := map[string]any{"key": int64(42)}
	if got := GetInt64(m, "key"); got != 42 {
		t.Errorf("GetInt64 int64: got %d, want 42", got)
	}
}

func TestGetInt64_PresentFloat64(t *testing.T) {
	m := map[string]any{"key": float64(3.9)}
	// float64 -> int64 truncates.
	if got := GetInt64(m, "key"); got != 3 {
		t.Errorf("GetInt64 float64: got %d, want 3", got)
	}
}

func TestGetInt64_Missing(t *testing.T) {
	m := map[string]any{}
	if got := GetInt64(m, "missing"); got != 0 {
		t.Errorf("GetInt64 missing: got %d, want 0", got)
	}
}

func TestGetInt64_WrongType(t *testing.T) {
	m := map[string]any{"key": "not a number"}
	if got := GetInt64(m, "key"); got != 0 {
		t.Errorf("GetInt64 wrong type: got %d, want 0", got)
	}
}

func TestGetInt64_NilValue(t *testing.T) {
	m := map[string]any{"key": nil}
	if got := GetInt64(m, "key"); got != 0 {
		t.Errorf("GetInt64 nil: got %d, want 0", got)
	}
}

func TestGetInt64_NilMap(t *testing.T) {
	var m map[string]any
	if got := GetInt64(m, "key"); got != 0 {
		t.Errorf("GetInt64 nil map: got %d, want 0", got)
	}
}

func TestGetInt64_NegativeValue(t *testing.T) {
	m := map[string]any{"key": int64(-100)}
	if got := GetInt64(m, "key"); got != -100 {
		t.Errorf("GetInt64 negative: got %d, want -100", got)
	}
}

func TestGetInt64_Zero(t *testing.T) {
	m := map[string]any{"key": int64(0)}
	if got := GetInt64(m, "key"); got != 0 {
		t.Errorf("GetInt64 zero: got %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// GetFloat64
// ---------------------------------------------------------------------------

func TestGetFloat64_PresentFloat64(t *testing.T) {
	m := map[string]any{"key": float64(3.14)}
	if got := GetFloat64(m, "key"); got != 3.14 {
		t.Errorf("GetFloat64 float64: got %f, want 3.14", got)
	}
}

func TestGetFloat64_PresentInt64(t *testing.T) {
	m := map[string]any{"key": int64(7)}
	if got := GetFloat64(m, "key"); got != 7.0 {
		t.Errorf("GetFloat64 int64: got %f, want 7.0", got)
	}
}

func TestGetFloat64_Missing(t *testing.T) {
	m := map[string]any{}
	if got := GetFloat64(m, "missing"); got != 0 {
		t.Errorf("GetFloat64 missing: got %f, want 0", got)
	}
}

func TestGetFloat64_WrongType(t *testing.T) {
	m := map[string]any{"key": "not a number"}
	if got := GetFloat64(m, "key"); got != 0 {
		t.Errorf("GetFloat64 wrong type: got %f, want 0", got)
	}
}

func TestGetFloat64_NilValue(t *testing.T) {
	m := map[string]any{"key": nil}
	if got := GetFloat64(m, "key"); got != 0 {
		t.Errorf("GetFloat64 nil: got %f, want 0", got)
	}
}

func TestGetFloat64_NilMap(t *testing.T) {
	var m map[string]any
	if got := GetFloat64(m, "key"); got != 0 {
		t.Errorf("GetFloat64 nil map: got %f, want 0", got)
	}
}

func TestGetFloat64_NegativeValue(t *testing.T) {
	m := map[string]any{"key": float64(-1.5)}
	if got := GetFloat64(m, "key"); got != -1.5 {
		t.Errorf("GetFloat64 negative: got %f, want -1.5", got)
	}
}

func TestGetFloat64_Zero(t *testing.T) {
	m := map[string]any{"key": float64(0)}
	if got := GetFloat64(m, "key"); got != 0 {
		t.Errorf("GetFloat64 zero: got %f, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// extractValue (internal – tested indirectly and also directly since same pkg)
// ---------------------------------------------------------------------------

func TestExtractValue_Nil(t *testing.T) {
	got := extractValue(nil)
	if got != nil {
		t.Errorf("extractValue nil: got %v, want nil", got)
	}
}

func TestExtractValue_String(t *testing.T) {
	got := extractValue("hello")
	if got != "hello" {
		t.Errorf("extractValue string: got %v, want 'hello'", got)
	}
}

func TestExtractValue_Int64(t *testing.T) {
	got := extractValue(int64(99))
	if got != int64(99) {
		t.Errorf("extractValue int64: got %v, want 99", got)
	}
}

func TestExtractValue_EmptySlice(t *testing.T) {
	got := extractValue([]any{})
	if got != nil {
		t.Errorf("extractValue empty slice: got %v, want nil", got)
	}
}

func TestExtractValue_TypedPair_Null(t *testing.T) {
	// [1, nil] => NULL => nil
	got := extractValue([]any{int64(1), nil})
	if got != nil {
		t.Errorf("extractValue NULL pair: got %v, want nil", got)
	}
}

func TestExtractValue_TypedPair_String(t *testing.T) {
	// [2, "hello"] => STRING => "hello"
	got := extractValue([]any{int64(2), "hello"})
	if got != "hello" {
		t.Errorf("extractValue STRING pair: got %v, want 'hello'", got)
	}
}

func TestExtractValue_TypedPair_Integer(t *testing.T) {
	// [3, int64(42)] => INTEGER => 42
	got := extractValue([]any{int64(3), int64(42)})
	if got != int64(42) {
		t.Errorf("extractValue INTEGER pair: got %v, want 42", got)
	}
}

func TestExtractValue_TypedPair_BooleanTrue(t *testing.T) {
	// [4, "true"] => BOOLEAN => true
	got := extractValue([]any{int64(4), "true"})
	if got != true {
		t.Errorf("extractValue BOOLEAN true: got %v, want true", got)
	}
}

func TestExtractValue_TypedPair_BooleanFalse(t *testing.T) {
	// [4, "false"] => BOOLEAN => false
	got := extractValue([]any{int64(4), "false"})
	if got != false {
		t.Errorf("extractValue BOOLEAN false: got %v, want false", got)
	}
}

func TestExtractValue_TypedPair_Double(t *testing.T) {
	// [5, "3.14"] => DOUBLE => 3.14
	got := extractValue([]any{int64(5), "3.14"})
	f, ok := got.(float64)
	if !ok {
		t.Fatalf("extractValue DOUBLE: expected float64, got %T", got)
	}
	if f < 3.13 || f > 3.15 {
		t.Errorf("extractValue DOUBLE: got %f, want ~3.14", f)
	}
}

func TestExtractValue_TypedPair_Array(t *testing.T) {
	// [6, [elem1, elem2]] => ARRAY
	inner := []any{int64(2), "item"}
	got := extractValue([]any{int64(6), []any{inner}})
	arr, ok := got.([]any)
	if !ok {
		t.Fatalf("extractValue ARRAY: expected []any, got %T", got)
	}
	if len(arr) != 1 {
		t.Errorf("extractValue ARRAY: expected 1 element, got %d", len(arr))
	}
}

func TestExtractValue_TypedPair_UnknownType(t *testing.T) {
	// Unknown type ID => return the value as-is.
	got := extractValue([]any{int64(99), "raw"})
	if got != "raw" {
		t.Errorf("extractValue unknown type: got %v, want 'raw'", got)
	}
}

// ---------------------------------------------------------------------------
// extractTypedValue – additional edge cases
// ---------------------------------------------------------------------------

func TestExtractTypedValue_NonInt64TypeID(t *testing.T) {
	// If the first element is not int64, the raw value should be returned.
	got := extractTypedValue([]any{"not-int64", "value"})
	if got != "value" {
		t.Errorf("extractTypedValue non-int64 type ID: got %v, want 'value'", got)
	}
}

func TestExtractTypedValue_WrongLength(t *testing.T) {
	// Pairs with length != 2 should be returned as-is.
	input := []any{int64(2), "hello", "extra"}
	got := extractTypedValue(input)
	// Should return the original pair slice unchanged.
	slice, ok := got.([]any)
	if !ok {
		t.Fatalf("extractTypedValue wrong length: expected []any, got %T", got)
	}
	if len(slice) != 3 {
		t.Errorf("extractTypedValue wrong length: expected 3, got %d", len(slice))
	}
}

// ---------------------------------------------------------------------------
// parseNode
// ---------------------------------------------------------------------------

func TestParseNode_NilInput(t *testing.T) {
	got := parseNode(nil)
	if got != nil {
		t.Errorf("parseNode nil: expected nil, got %v", got)
	}
}

func TestParseNode_NotSlice(t *testing.T) {
	got := parseNode("not a slice")
	if got != nil {
		t.Errorf("parseNode non-slice: expected nil, got %v", got)
	}
}

func TestParseNode_EmptyNode(t *testing.T) {
	// A node with fewer than 3 elements – props map should be empty.
	got := parseNode([]any{int64(1), []any{}})
	if got == nil {
		t.Fatal("parseNode empty: expected non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("parseNode empty: expected 0 props, got %d", len(got))
	}
}

func TestParseNode_WithStringKeyProp(t *testing.T) {
	// Node: [nodeID, [labelIDs], [[propKey, typeID, value]]]
	prop := []any{"name", int64(2), "Alice"}
	node := []any{int64(1), []any{}, []any{prop}}

	got := parseNode(node)
	if got == nil {
		t.Fatal("parseNode with prop: expected non-nil map")
	}
	if v, ok := got["name"]; !ok {
		t.Error("parseNode with prop: key 'name' not found")
	} else if v != "Alice" {
		t.Errorf("parseNode with prop: want 'Alice', got %v", v)
	}
}

// ---------------------------------------------------------------------------
// NewGraphRAGRepository – construction (no external service required)
// ---------------------------------------------------------------------------

func TestNewGraphRAGRepository_NilClient(t *testing.T) {
	// NewGraphRAGRepository wraps a *Client; passing nil is valid for
	// construction – methods that call Query will panic, but the object
	// itself is created without issues.
	repo := NewGraphRAGRepository(nil)
	if repo == nil {
		t.Fatal("expected non-nil GraphRAGRepository")
	}
}

// ---------------------------------------------------------------------------
// MultiHopSearch – empty seedIDs fast path
// ---------------------------------------------------------------------------

func TestMultiHopSearch_EmptySeedIDs(t *testing.T) {
	repo := NewGraphRAGRepository(nil) // client never called for empty seeds
	results, err := repo.MultiHopSearch(nil, nil, 3, 10, "tenant-1") //nolint:staticcheck
	if err != nil {
		t.Fatalf("MultiHopSearch empty seeds: unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("MultiHopSearch empty seeds: expected nil results, got %v", results)
	}
}
