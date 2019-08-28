package processor

import (
	"testing"
	"time"
)

func TestMakeTimestampNano(t *testing.T) {
	res := makeTimestampNano()
	time.Sleep(5 * time.Nanosecond)

	if res == makeTimestampNano() {
		t.Error("Should not match")
	}
}

func TestMakeTimestampMilli(t *testing.T) {
	res := makeTimestampMilli()
	time.Sleep(5 * time.Millisecond)

	if res == makeTimestampMilli() {
		t.Error("Should not match")
	}
}

func TestMakeFuzzy(t *testing.T) {
	fuzzy := makeFuzzyDistanceOne("test")

	if len(fuzzy) == 0 {
		t.Error("Should get back non empty slice")
	}
}

func TestMakeFuzzyTwo(t *testing.T) {
	fuzzy := makeFuzzyDistanceTwo("test")

	if len(fuzzy) == 0 {
		t.Error("Should get back non empty slice")
	}
}
