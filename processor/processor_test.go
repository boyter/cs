package processor

import "testing"

func TestCleanSearchString(t *testing.T) {
	SearchString = []string{
		"AND",
		"OR",
		"NOT",
		"",
		"THE",
	}
	CleanSearchString()

	if len(SearchString) != 4 {
		t.Error("Expected 4")
	}

	if SearchString[3] != "the" {
		t.Error("Expected the not THE")
	}
}

func TestProcess(t *testing.T) {
	p := NewProcess()
	p.StartProcess()
}
