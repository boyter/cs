// SPDX-License-Identifier: MIT OR Unlicense

package str

import (
	"regexp"
	"testing"
)

var test_MatchEndCase = `1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890 1test`
var test_UnicodeMatchEndCase = `ȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺ Ⱥtest`

var test_MatchEndCaseLarge = `1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890 1test`
var test_UnicodeMatchEndCaseLarge = `ȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺȺ Ⱥtest`

func BenchmarkFindAllIndex(b *testing.B) {
	r := regexp.MustCompile(`test`)
	haystack := []byte(test_MatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_MatchEndCase, "test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexLarge(b *testing.B) {
	r := regexp.MustCompile(`test`)
	haystack := []byte(test_MatchEndCaseLarge)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexAllLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_MatchEndCaseLarge, "test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicode(b *testing.B) {
	r := regexp.MustCompile(`test`)
	haystack := []byte(test_UnicodeMatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexAllUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_UnicodeMatchEndCase, "test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicodeLarge(b *testing.B) {
	r := regexp.MustCompile(`test`)
	haystack := []byte(test_UnicodeMatchEndCaseLarge)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkIndexAllUnicodeLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_UnicodeMatchEndCaseLarge, "test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkFindAllIndexManyPartialMatches(b *testing.B) {
	r := regexp.MustCompile(`1test`)
	haystack := []byte(test_MatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkIndexAllManyPartialMatches(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_MatchEndCase, "1test", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkFindAllIndexUnicodeManyPartialMatches(b *testing.B) {
	r := regexp.MustCompile(`Ⱥtest`)
	haystack := []byte(test_UnicodeMatchEndCase)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

// This benchmark simulates a bad case of there being many
// partial matches where the first character in the needle
// can be found throughout the haystack
func BenchmarkIndexAllUnicodeManyPartialMatches(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matches := IndexAll(test_UnicodeMatchEndCase, "Ⱥtest", -1)
		if len(matches) != 1 {
			b.Error("Expected single match")
		}
	}
}

func BenchmarkFindAllIndexUnicodeManyPartialMatchesVeryLarge(b *testing.B) {
	r := regexp.MustCompile(`Ⱥtest`)

	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	haystack := []byte(large)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 101 {
			b.Error("Expected 101 match got", len(matches))
		}
	}
}

func BenchmarkIndexAllUnicodeManyPartialMatchesVeryLarge(b *testing.B) {
	var large string
	for i := 0; i <= 100; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	for i := 0; i < b.N; i++ {
		matches := IndexAll(large, "Ⱥtest", -1)
		if len(matches) != 101 {
			b.Error("Expected 101 match got", len(matches))
		}
	}
}

func BenchmarkFindAllIndexUnicodeManyPartialMatchesSuperLarge(b *testing.B) {
	r := regexp.MustCompile(`Ⱥtest`)

	var large string
	for i := 0; i <= 500; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	haystack := []byte(large)

	for i := 0; i < b.N; i++ {
		matches := r.FindAllIndex(haystack, -1)
		if len(matches) != 501 {
			b.Error("Expected 501 match got", len(matches))
		}
	}
}

func BenchmarkIndexAllUnicodeManyPartialMatchesSuperLarge(b *testing.B) {
	var large string
	for i := 0; i <= 500; i++ {
		large += test_UnicodeMatchEndCaseLarge
	}

	for i := 0; i < b.N; i++ {
		matches := IndexAll(large, "Ⱥtest", -1)
		if len(matches) != 501 {
			b.Error("Expected 501 match got", len(matches))
		}
	}
}
