package processor

import (
	"testing"
)

func TestCodeCleaner(t *testing.T) {
	got := codeCleaner(`package processor

import (
	"testing"
)
`)
	expected := `package processor import "testing" package processor import "testing" package processor import "testing" package processor import testing package processor import testing`

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}
