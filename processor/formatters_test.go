package processor

import (
	"strings"
	"testing"
)

func TestPrintTrace(t *testing.T) {
	Trace = true
	printTrace("Testing print trace")
	Trace = false
	printTrace("Testing print trace")
}

func TestPrintDebug(t *testing.T) {
	Debug = true
	printDebug("Testing print debug")
	Debug = false
	printDebug("Testing print debug")
}

func TestPrintWarn(t *testing.T) {
	Verbose = true
	printWarn("Testing print warn")
	Verbose = false
	printWarn("Testing print warn")
}

func TestPrintError(t *testing.T) {
	printError("Testing print error")
}

func TestGetFormattedTime(t *testing.T) {
	res := getFormattedTime()

	if !strings.Contains(res, "T") {
		t.Error("String does not contain expected T", res)
	}

	if !strings.Contains(res, "Z") {
		t.Error("String does not contain expected Z", res)
	}
}

func TestToJson(t *testing.T) {
	fileListQueue := make(chan *FileJob, 100)

	fileListQueue <- &FileJob{
		Filename:  "",
		Extension: "",
		Location:  "",
		Content:   nil,
		Bytes:     0,
		Hash:      nil,
		Binary:    false,
		Score:     0,
		Locations: nil,
	}
	close(fileListQueue)

	json := toJSON(fileListQueue)

	if json == "" {
		t.Error("Expected something")
	}
}