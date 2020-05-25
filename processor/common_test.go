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
