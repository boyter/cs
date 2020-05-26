// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"testing"
)

func TestRemoveIntDuplicates(t *testing.T) {
	clean := removeIntDuplicates([]int{1, 1})

	if len(clean) != 1 {
		t.Error("Should have no duplicates")
	}
}

func TestRemoveIntDuplicatesMultiple(t *testing.T) {
	clean := removeIntDuplicates([]int{1, 1, 1, 1, 1, 2})

	if len(clean) != 2 {
		t.Error("Should have no duplicates")
	}
}

func TestAbs(t *testing.T) {
	v := abs(-1)

	if v != 1 {
		t.Error("expect absolute value for 1 got", v)
	}
}

func TestAbsPositive(t *testing.T) {
	v := abs(100)

	if v != 100 {
		t.Error("expect absolute value for 100 got", v)
	}
}

func TestTryParseInt(t *testing.T) {
	if tryParseInt("1", 0) != 1 {
		t.Error("failed")
	}

	if tryParseInt("1a", 0) != 0 {
		t.Error("failed")
	}

	if tryParseInt("9999", 0) != 9999 {
		t.Error("failed")
	}

	if tryParseInt("9,999", 0) != 0 {
		t.Error("failed")
	}
}
