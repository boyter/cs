package processor

import (
	"testing"
)

func TestProcessMatchesSingleMatch(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("match"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 1 {
		t.Errorf("Score should be 1")
	}
}

func TestProcessMatchesTwoMatch(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("match"),
		[]byte("this"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 1 {
		t.Error("Score should be 1 got", res.Score)
	}
}

func TestProcessMatchesTwoAndMatch(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("match"),
		[]byte("AND"),
		[]byte("this"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 1 {
		t.Error("Score should be 1 got", res.Score)
	}
}

func TestProcessMatchesTwoNotMatch(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("match"),
		[]byte("NOT"),
		[]byte("this"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score != 0 {
		t.Error("Score should be 0 got", res.Score)
	}
}

func TestProcessMatchesFuzzyOne(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("this~1"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score == 0 {
		t.Error("Score should not be 0 got", res.Score)
	}
}

func TestProcessMatchesFuzzyTwo(t *testing.T) {
	ResultLimit = 100
	SearchBytes = [][]byte{
		[]byte("this~2"),
	}

	res := fileJob{
		Locations: map[string][]int{},
	}

	matches := processMatches(&res, []byte("this is a match"))

	if matches {
		t.Errorf("Response should be false")
	}

	if res.Score == 0 {
		t.Error("Score should be not 0 got", res.Score)
	}
}

func TestFileReaderWorker(t *testing.T) {
	ResultLimit = 100

	input := make(chan *fileJob, 10)
	output := make(chan *fileJob, 10)

	input <- &fileJob{
		Filename:  "workers.go",
		Extension: "go",
		Location:  "./workers.go",
		Content:   nil,
		Bytes:     0,
		Hash:      nil,
		Binary:    false,
		Score:     0,
		Locations: nil,
	}
	close(input)

	FileReaderWorker(input, output)

	out := []*fileJob{}
	for o := range output {
		out = append(out, o)
	}

	if len(out) == 0 {
		t.Error("Expected at least one")
	}
}

func TestFileProcessorWorker(t *testing.T) {
	ResultLimit = 100

	input := make(chan *fileJob, 10)
	output := make(chan *fileJob, 10)

	input <- &fileJob{
		Filename:  "workers.go",
		Extension: "go",
		Location:  "./workers.go",
		Content:   []byte("this is some content of stuff"),
		Bytes:     0,
		Hash:      nil,
		Binary:    false,
		Score:     100,
		Locations: map[string][]int{},
	}
	close(input)

	FileProcessorWorker(input, output)

	out := []*fileJob{}
	for o := range output {
		out = append(out, o)
	}

	if len(out) == 0 {
		t.Error("Expected at least one")
	}
}
