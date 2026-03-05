package service

import "testing"

func TestPIIService_DetectEmail(t *testing.T) {
	svc := NewPIIService()
	matches := svc.Detect("Contact me at user@example.com for details")
	found := false
	for _, m := range matches {
		if m.Type == PIIEmail {
			found = true
		}
	}
	if !found {
		t.Error("expected to detect email PII")
	}
}

func TestPIIService_DetectPhone(t *testing.T) {
	svc := NewPIIService()
	matches := svc.Detect("Call me at (555) 123-4567")
	found := false
	for _, m := range matches {
		if m.Type == PIIPhone {
			found = true
		}
	}
	if !found {
		t.Error("expected to detect phone PII")
	}
}

func TestPIIService_DetectSSN(t *testing.T) {
	svc := NewPIIService()
	if !svc.ContainsPII("SSN: 123-45-6789") {
		t.Error("expected to detect SSN PII")
	}
}

func TestPIIService_DetectCreditCard(t *testing.T) {
	svc := NewPIIService()
	if !svc.ContainsPII("Card: 4111 1111 1111 1111") {
		t.Error("expected to detect credit card PII")
	}
}

func TestPIIService_DetectJWT(t *testing.T) {
	svc := NewPIIService()
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	if !svc.ContainsPII(jwt) {
		t.Error("expected to detect JWT PII")
	}
}

func TestPIIService_DetectAPIKey(t *testing.T) {
	svc := NewPIIService()
	if !svc.ContainsPII("api_key = 'sk_test_FAKE_KEY_1234567890abcdef'") {
		t.Error("expected to detect API key PII")
	}
}

func TestPIIService_NoPII(t *testing.T) {
	svc := NewPIIService()
	if svc.ContainsPII("This is a normal text without any sensitive data") {
		t.Error("expected no PII detection")
	}
}

func TestPIIService_Mask(t *testing.T) {
	svc := NewPIIService()
	masked := svc.Mask("Email: user@example.com and SSN: 123-45-6789")
	if masked == "Email: user@example.com and SSN: 123-45-6789" {
		t.Error("expected text to be masked")
	}
	if !contains(masked, "[EMAIL]") {
		t.Error("expected [EMAIL] placeholder")
	}
	if !contains(masked, "[SSN]") {
		t.Error("expected [SSN] placeholder")
	}
}

func TestPIIService_MaskForLLM(t *testing.T) {
	svc := NewPIIService()
	masked, summary := svc.MaskForLLM("Contact user@test.com")
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if !contains(masked, "[EMAIL]") {
		t.Error("expected masked output")
	}
}

func TestPIIService_MaskForLLM_NoPII(t *testing.T) {
	svc := NewPIIService()
	text := "Normal text"
	masked, summary := svc.MaskForLLM(text)
	if masked != text {
		t.Error("expected unchanged text")
	}
	if summary != "" {
		t.Error("expected empty summary")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && containsSubstr(s, sub)
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
