package processor

import (
	"testing"
)

func TestCleanSearchString(t *testing.T) {
	SearchString = []string{
		"AND",
		"OR",
		"NOT",
		"",
		"THE",
	}
	CleanSearchString()

	if len(SearchBytes) != 4 {
		t.Error("Expected 4")
	}
}

func TestProcess(t *testing.T) {
	p := NewProcess()
	p.StartProcess()
}
