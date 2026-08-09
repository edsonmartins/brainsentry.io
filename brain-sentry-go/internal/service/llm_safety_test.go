package service

import (
	"strings"
	"testing"

	"github.com/integraltech/brainsentry/internal/security"
)

func TestFrameLLMDataPreservesLongContentAcrossSafeFrames(t *testing.T) {
	content := strings.Repeat("conteudo seguro ", 200)
	framed := frameLLMData("memory", "test", content)

	if !strings.Contains(framed, security.SystemPromptPreamble) {
		t.Fatal("missing untrusted-data preamble")
	}
	if strings.Count(framed, "<memory ") < 2 {
		t.Fatal("long content was not split into bounded memory frames")
	}
	if !strings.Contains(framed, "memory:2") {
		t.Fatal("missing second memory chunk")
	}
}
