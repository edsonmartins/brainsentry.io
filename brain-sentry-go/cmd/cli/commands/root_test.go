package commands

import (
	"strings"
	"testing"
)

func TestExecute_Help(t *testing.T) {
	// Execute with --help should not error
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	expected := []string{"add", "search", "list", "edit", "correct", "review", "import"}
	cmds := rootCmd.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}
	for _, e := range expected {
		if !names[e] {
			t.Errorf("missing subcommand: %s", e)
		}
	}
}

func TestRootCmd_DefaultTenant(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("tenant")
	if flag == nil {
		t.Fatal("expected --tenant flag")
	}
	if !strings.Contains(flag.DefValue, "a9f814d2") {
		t.Errorf("expected default tenant UUID, got %s", flag.DefValue)
	}
}

func TestRootCmd_OutputFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("expected --output flag")
	}
	if flag.DefValue != "table" {
		t.Errorf("expected default output=table, got %s", flag.DefValue)
	}
}
