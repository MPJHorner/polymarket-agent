package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	root := rootCmd
	b := bytes.NewBufferString("")
	root.SetOut(b)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out := b.String()
	if !contains(out, "polytracker [command]") {
		t.Errorf("expected usage info, got %s", out)
	}
}

func TestSubcommands(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{[]string{"scan"}, "Scanning Polymarket for recent activity..."},
		{[]string{"analyze", "0x123"}, "Fetching history for trader: 0x123"},
		{[]string{"export"}, "Exporting data..."},
	}

	for _, tc := range cases {
		root := rootCmd
		b := bytes.NewBufferString("")
		root.SetOut(b)
		root.SetArgs(tc.args)

		// Note: we might need to reset or mock config loading if it fails in CI
		_ = root.Execute()

		out := b.String()
		if !contains(out, tc.expected) {
			t.Errorf("args %v: expected output containing %q, got %q", tc.args, tc.expected, out)
		}
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
