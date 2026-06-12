package git

import "testing"

func TestParseNumstatValue(t *testing.T) {
	if n, err := parseNumstatValue("42"); err != nil || n != 42 {
		t.Errorf("parseNumstatValue(\"42\") = %d, %v; want 42, nil", n, err)
	}

	// Binary files show "-" in numstat output.
	if n, err := parseNumstatValue("-"); err != nil || n != 0 {
		t.Errorf("parseNumstatValue(\"-\") = %d, %v; want 0, nil", n, err)
	}

	if _, err := parseNumstatValue("abc"); err == nil {
		t.Error("parseNumstatValue(\"abc\") should fail")
	}
}

func TestCombineNumstatOutputs(t *testing.T) {
	tests := []struct {
		name     string
		unstaged string
		staged   string
		want     string
	}{
		{"both empty", "", "", ""},
		{"only staged", "", "1\t2\ta.go", "1\t2\ta.go"},
		{"only unstaged", "3\t4\tb.go", "", "3\t4\tb.go"},
		{"both", "3\t4\tb.go\n", "1\t2\ta.go\n", "3\t4\tb.go\n1\t2\ta.go"},
	}

	for _, tt := range tests {
		if got := combineNumstatOutputs(tt.unstaged, tt.staged); got != tt.want {
			t.Errorf("%s: combineNumstatOutputs() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
