package service

import "testing"

func TestComputeSimHash_SameText(t *testing.T) {
	h1 := ComputeSimHash("The quick brown fox jumps over the lazy dog")
	h2 := ComputeSimHash("The quick brown fox jumps over the lazy dog")
	if h1 != h2 {
		t.Error("identical text should produce identical SimHash")
	}
}

func TestComputeSimHash_SimilarText(t *testing.T) {
	h1 := ComputeSimHash("The quick brown fox jumps over the lazy dog")
	h2 := ComputeSimHash("The quick brown fox leaps over the lazy dog")
	dist := SimHashHammingDistance(h1, h2)
	if dist > 10 {
		t.Errorf("similar text should have small Hamming distance, got %d", dist)
	}
}

func TestComputeSimHash_DifferentText(t *testing.T) {
	h1 := ComputeSimHash("The quick brown fox jumps over the lazy dog")
	h2 := ComputeSimHash("Quantum computing enables parallel processing of complex algorithms")
	dist := SimHashHammingDistance(h1, h2)
	if dist < 5 {
		t.Errorf("very different text should have large Hamming distance, got %d", dist)
	}
}

func TestSimHashHammingDistance_Zero(t *testing.T) {
	if SimHashHammingDistance(0xFF, 0xFF) != 0 {
		t.Error("same values should have distance 0")
	}
}

func TestSimHashHammingDistance_One(t *testing.T) {
	if SimHashHammingDistance(0xFF, 0xFE) != 1 {
		t.Error("expected distance 1")
	}
}

func TestSimHashToHex_RoundTrip(t *testing.T) {
	original := uint64(0x123456789ABCDEF0)
	hex := SimHashToHex(original)
	if len(hex) != 16 {
		t.Errorf("expected 16 chars, got %d", len(hex))
	}
	back := SimHashFromHex(hex)
	if back != original {
		t.Errorf("round-trip failed: got %x, want %x", back, original)
	}
}

func TestSimHashToHex_Zero(t *testing.T) {
	hex := SimHashToHex(0)
	if hex != "0000000000000000" {
		t.Errorf("expected all zeros, got %s", hex)
	}
}

func TestComputeSimHash_Empty(t *testing.T) {
	h := ComputeSimHash("")
	if h != 0 {
		t.Error("empty text should produce zero hash")
	}
}

func TestExtractChainOfThought(t *testing.T) {
	content := "Here is my answer. <THOUGHT>I need to think about this carefully.</THOUGHT> The result is 42."
	cleaned, cot := extractChainOfThought(content)
	if cot == "" {
		t.Error("expected non-empty chain of thought")
	}
	if cleaned == content {
		t.Error("expected THOUGHT tags to be removed")
	}
	if cot != "I need to think about this carefully." {
		t.Errorf("unexpected CoT: %s", cot)
	}
}

func TestExtractChainOfThought_Multiple(t *testing.T) {
	content := "<THOUGHT>Step 1</THOUGHT> Answer part 1. <THOUGHT>Step 2</THOUGHT> Answer part 2."
	_, cot := extractChainOfThought(content)
	if cot == "" {
		t.Error("expected non-empty chain of thought")
	}
}

func TestExtractChainOfThought_NoThought(t *testing.T) {
	content := "Just a normal text without thought tags"
	cleaned, cot := extractChainOfThought(content)
	if cot != "" {
		t.Error("expected empty chain of thought")
	}
	if cleaned != content {
		t.Error("expected unchanged content")
	}
}
