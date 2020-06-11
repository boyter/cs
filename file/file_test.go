// SPDX-License-Identifier: MIT OR Unlicense

package file

import (
	"os"
	"strings"
	"testing"
)

func TestFindRepositoryRoot(t *testing.T) {
	// We expect this to walk back from file to cs
	curdir, _ := os.Getwd()
	root := FindRepositoryRoot(curdir)

	if strings.HasSuffix(root, "file") {
		t.Error("Expected to walk back to root")
	}
}

func TestNewFileWalker(t *testing.T) {
	fileListQueue := make(chan *File, 1000) // NB we set buffered to ensure we get everything
	curdir, _ := os.Getwd()
	walker := NewFileWalker(curdir, fileListQueue)
	_ = walker.Start()

	count := 0
	for range fileListQueue {
		count++
	}

	if count == 0 {
		t.Error("Expected to find at least one file")
	}
}

func TestNewFileWalkerStuff(t *testing.T) {
	fileListQueue := make(chan *File, 1000) // NB we set buffered to ensure we get everything
	curdir, _ := os.Getwd()
	walker := NewFileWalker(curdir, fileListQueue)

	if walker.Walking() != false {
		t.Error("should not be walking yet")
	}

	walker.Terminate()
	_ = walker.Start()

	count := 0
	for range fileListQueue {
		count++
	}

	if count != 0 {
		t.Error("Expected to find no files")
	}
}

func TestGetExtension(t *testing.T) {
	got := GetExtension("something.c")
	expected := "c"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionNoExtension(t *testing.T) {
	got := GetExtension("something")
	expected := "something"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionMultipleDots(t *testing.T) {
	got := GetExtension(".travis.yml")
	expected := "yml"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionMultipleExtensions(t *testing.T) {
	got := GetExtension("something.go.yml")
	expected := "yml"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionStartsWith(t *testing.T) {
	got := GetExtension(".gitignore")
	expected := ".gitignore"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionTypeScriptDefinition(t *testing.T) {
	got := GetExtension("test.d.ts")
	expected := "ts"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}

func TestGetExtensionRegression(t *testing.T) {
	got := GetExtension("DeviceDescription.stories.tsx")
	expected := "tsx"

	if got != expected {
		t.Errorf("Expected %s got %s", expected, got)
	}
}
