package printer

import "testing"

func TestWriteColored(t *testing.T) {
	WriteColored("", map[string][]int{}, "", "")
}
