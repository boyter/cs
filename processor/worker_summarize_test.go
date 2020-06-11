// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
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
	Error = true
	printError("Testing print error")
	Error = false
}

//func TestToJson(t *testing.T) {
//	ResultLimit = 100
//	fileListQueue := make(chan *fileJob, 100)
//
//	fileListQueue <- &fileJob{
//		Filename:  "",
//		Extension: "",
//		Location:  "",
//		Content:   nil,
//		Bytes:     0,
//		Hash:      nil,
//		Binary:    false,
//		Score:     0,
//		Locations: nil,
//	}
//	close(fileListQueue)
//
//	json := toJSON(fileListQueue)
//
//	if json == "" {
//		t.Error("Expected something")
//	}
//}

//func TestToJsonMultiple(t *testing.T) {
//	ResultLimit = 100
//	fileListQueue := make(chan *fileJob, 100)
//
//	fileListQueue <- &fileJob{
//		Filename:  "Something",
//		Extension: "",
//		Location:  "",
//		Content:   nil,
//		Bytes:     0,
//		Hash:      nil,
//		Binary:    false,
//		Score:     100,
//		Locations: nil,
//	}
//
//	for i := 0; i < 10; i++ {
//		fileListQueue <- &fileJob{
//			Filename:  strconv.Itoa(i),
//			Extension: "",
//			Location:  "",
//			Content:   nil,
//			Bytes:     0,
//			Hash:      nil,
//			Binary:    false,
//			Score:     10,
//			Locations: nil,
//		}
//	}
//	close(fileListQueue)
//
//	json := toJSON(fileListQueue)
//
//	if json == "" {
//		t.Error("Expected something")
//	}
//}

//func TestFileSummerize(t *testing.T) {
//	ResultLimit = 100
//	Format = "text"
//	fileListQueue := make(chan *fileJob, 100)
//
//	fileListQueue <- &fileJob{
//		Filename:  "Something",
//		Extension: "",
//		Location:  "",
//		Content:   nil,
//		Bytes:     0,
//		Hash:      nil,
//		Binary:    false,
//		Score:     100,
//		Locations: nil,
//	}
//
//	for i := 0; i < 10; i++ {
//		fileListQueue <- &fileJob{
//			Filename:  strconv.Itoa(i),
//			Extension: "",
//			Location:  "",
//			Content:   nil,
//			Bytes:     0,
//			Hash:      nil,
//			Binary:    false,
//			Score:     10,
//			Locations: nil,
//		}
//	}
//	close(fileListQueue)
//
//	res := fileSummarize(fileListQueue)
//
//	if res != "" {
//		t.Error("Expected something")
//	}
//}
