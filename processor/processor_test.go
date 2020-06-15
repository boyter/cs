// SPDX-License-Identifier: MIT OR Unlicense

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
}
