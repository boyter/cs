package string

type Horspool struct {
	table   map[rune]int
	indexes []int
}

func (t *Horspool) Search(text, pattern []rune) []int {

	table := t.table
	if table == nil {
		table = map[rune]int{}
	} else {
		for r := range table {
			delete(table, r)
		}
	}
	indexes := t.indexes
	if cap(indexes) < 1 {
		indexes = make([]int, 0, 100)
	}
	indexes = indexes[:0]

	textLength := len(text)
	patternLength := len(pattern)

	if textLength == 0 || patternLength == 0 || patternLength > textLength {
		return indexes
	}

	lastPatternByte := patternLength - 1

	{
		for _, r := range pattern {
			table[r] = patternLength
		}

		for i := 0; i < lastPatternByte; i++ {
			table[pattern[i]] = patternLength - 1 - i
		}
	}

	index := 0
	for index <= (textLength - patternLength) {
		for i := lastPatternByte; text[index+i] == pattern[i]; i-- {
			if i == 0 {
				indexes = append(indexes, index)
				break
			}
		}
		x, ok := table[text[index+lastPatternByte]]
		if ok {
			index += x
		} else {
			index += lastPatternByte
		}
	}
	t.table = table
	t.indexes = indexes
	return indexes
}

