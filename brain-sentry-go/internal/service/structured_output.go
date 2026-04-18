package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// StructuredOutputOption configures a StructuredOutput call.
type StructuredOutputOption struct {
	MaxRetries int    // default 2
	Instructions string // extra instruction appended to system prompt
}

// StructuredOutput runs an LLM call and unmarshals the response into T.
// Uses reflection to generate a hint-schema from T and injects it into the prompt.
// Retries on JSON parse failure with structured feedback.
//
// Example:
//
//   type Answer struct {
//       Question string `json:"question"`
//       Confidence float64 `json:"confidence"`
//   }
//   var a Answer
//   err := StructuredOutput[Answer](ctx, llm, "What is Go?", &a, StructuredOutputOption{})
func StructuredOutput[T any](
	ctx context.Context,
	llm LLMProvider,
	userPrompt string,
	result *T,
	opts StructuredOutputOption,
) error {
	if llm == nil {
		return fmt.Errorf("llm provider is nil")
	}
	if result == nil {
		return fmt.Errorf("result pointer is nil")
	}

	maxRetries := opts.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}

	schema := describeSchema(reflect.TypeOf(*result))
	systemPrompt := buildStructuredSystemPrompt(schema, opts.Instructions)
	currentUser := userPrompt

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := llm.Chat(ctx, []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: currentUser},
		})
		if err != nil {
			lastErr = err
			continue
		}

		cleaned := cleanJSON(response)
		if err := json.Unmarshal([]byte(cleaned), result); err != nil {
			lastErr = fmt.Errorf("parse json: %w", err)
			// Provide feedback in next iteration
			currentUser = fmt.Sprintf(
				"Your previous response was invalid JSON (error: %s). Return valid JSON matching the schema only.\n\nOriginal request: %s",
				err.Error(), userPrompt,
			)
			continue
		}

		return nil
	}

	return fmt.Errorf("structured output failed after %d attempts: %w", maxRetries+1, lastErr)
}

// buildStructuredSystemPrompt constructs the system prompt with schema hints.
func buildStructuredSystemPrompt(schema, extraInstructions string) string {
	var b strings.Builder
	b.WriteString("You are a structured output engine. Respond with valid JSON matching the schema. No markdown, no explanation, no code fences.\n\nSchema:\n")
	b.WriteString(schema)
	if extraInstructions != "" {
		b.WriteString("\n\nAdditional instructions:\n")
		b.WriteString(extraInstructions)
	}
	return b.String()
}

// describeSchema returns a human-readable description of a Go type's JSON shape.
// This is a lightweight alternative to full JSON Schema — sufficient for LLM hinting.
func describeSchema(t reflect.Type) string {
	if t == nil {
		return "any"
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return describeStruct(t)
	case reflect.Slice, reflect.Array:
		return "array of " + describeSchema(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("object with %s keys and %s values", describeSchema(t.Key()), describeSchema(t.Elem()))
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Interface:
		return "any"
	default:
		return t.Kind().String()
	}
}

func describeStruct(t reflect.Type) string {
	var b strings.Builder
	b.WriteString("{\n")

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		name := field.Name
		optional := false
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
			for _, p := range parts[1:] {
				if p == "omitempty" {
					optional = true
				}
			}
		} else if jsonTag == "-" {
			continue
		}

		fieldType := describeSchema(field.Type)

		// Honor a "desc" tag for inline field description.
		if desc := field.Tag.Get("desc"); desc != "" {
			fieldType = fieldType + " — " + desc
		}

		if optional {
			fieldType += " (optional)"
		}

		fmt.Fprintf(&b, "  %q: %s\n", name, fieldType)
	}
	b.WriteString("}")
	return b.String()
}
