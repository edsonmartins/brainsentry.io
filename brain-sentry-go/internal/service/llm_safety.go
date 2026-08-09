package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/integraltech/brainsentry/internal/security"
)

func frameLLMData(id, source, content string) string {
	var framed strings.Builder
	framed.WriteString(security.SystemPromptPreamble)
	framed.WriteString("\n\n")

	for chunk := 1; len(content) > 0; chunk++ {
		end := len(content)
		if end > security.MaxContentChars {
			end = security.MaxContentChars
			for end > 0 && !utf8.RuneStart(content[end]) {
				end--
			}
		}
		framed.WriteString(security.FrameMemory(fmt.Sprintf("%s:%d", id, chunk), source, content[:end]))
		content = content[end:]
		if len(content) > 0 {
			framed.WriteString("\n")
		}
	}

	if content == "" && framed.Len() == len(security.SystemPromptPreamble)+2 {
		framed.WriteString(security.FrameMemory(id, source, ""))
	}
	return framed.String()
}
