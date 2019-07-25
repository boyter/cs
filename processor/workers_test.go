package processor

import (
	"testing"
)

func TestIsBinaryTrue(t *testing.T) {
	DisableCheckBinary = false

	if !isBinary(0, 0) {
		t.Errorf("Expected to be true")
	}
}

func TestIsBinaryDisableCheck(t *testing.T) {
	DisableCheckBinary = true

	if isBinary(0, 0) {
		t.Errorf("Expected to be false")
	}
}
