// SPDX-License-Identifier: MIT

package main

import "testing"

func TestCycleRanker(t *testing.T) {
	order := []string{"simple", "tfidf", "tfidf2", "bm25", "structural"}
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}

	for i, want := range order {
		m.cfg.Ranker = order[i]
		m.cycleRanker()
		expected := order[(i+1)%len(order)]
		if m.cfg.Ranker != expected {
			t.Errorf("cycleRanker from %q: got %q, want %q", want, m.cfg.Ranker, expected)
		}
	}
}

func TestCycleRankerUnknownResetsToSimple(t *testing.T) {
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}
	m.cfg.Ranker = "unknown"
	m.cycleRanker()
	if m.cfg.Ranker != "simple" {
		t.Errorf("cycleRanker from unknown: got %q, want %q", m.cfg.Ranker, "simple")
	}
}

func TestCycleCodeFilter(t *testing.T) {
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}

	// default → only-code
	m.cfg.OnlyCode = false
	m.cfg.OnlyComments = false
	m.cfg.OnlyStrings = false
	m.cycleCodeFilter()
	if !m.cfg.OnlyCode || m.cfg.OnlyComments || m.cfg.OnlyStrings {
		t.Errorf("step 1: expected OnlyCode=true, OnlyComments=false, OnlyStrings=false")
	}

	// only-code → only-comments
	m.cycleCodeFilter()
	if m.cfg.OnlyCode || !m.cfg.OnlyComments || m.cfg.OnlyStrings {
		t.Errorf("step 2: expected OnlyCode=false, OnlyComments=true, OnlyStrings=false")
	}

	// only-comments → only-strings
	m.cycleCodeFilter()
	if m.cfg.OnlyCode || m.cfg.OnlyComments || !m.cfg.OnlyStrings {
		t.Errorf("step 3: expected OnlyCode=false, OnlyComments=false, OnlyStrings=true")
	}

	// only-strings → default
	m.cycleCodeFilter()
	if m.cfg.OnlyCode || m.cfg.OnlyComments || m.cfg.OnlyStrings {
		t.Errorf("step 4: expected OnlyCode=false, OnlyComments=false, OnlyStrings=false")
	}
}

func TestCycleCodeFilterAutoSelectsStructuralRanker(t *testing.T) {
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}
	m.cfg.Ranker = "bm25"
	m.cfg.OnlyCode = false
	m.cfg.OnlyComments = false

	m.cycleCodeFilter() // → only-code
	if m.cfg.Ranker != "structural" {
		t.Errorf("expected ranker=structural when only-code active, got %q", m.cfg.Ranker)
	}
}

func TestCycleGravity(t *testing.T) {
	order := []string{"off", "low", "default", "logic", "brain"}
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}

	for i, want := range order {
		m.cfg.GravityIntent = order[i]
		m.cycleGravity()
		expected := order[(i+1)%len(order)]
		if m.cfg.GravityIntent != expected {
			t.Errorf("cycleGravity from %q: got %q, want %q", want, m.cfg.GravityIntent, expected)
		}
	}
}

func TestCycleGravityUnknownResetsToOff(t *testing.T) {
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}
	m.cfg.GravityIntent = "unknown"
	m.cycleGravity()
	if m.cfg.GravityIntent != "off" {
		t.Errorf("cycleGravity from unknown: got %q, want %q", m.cfg.GravityIntent, "off")
	}
}

func TestCodeFilterLabel(t *testing.T) {
	cfg := DefaultConfig()
	m := &model{cfg: &cfg}

	m.cfg.OnlyCode = false
	m.cfg.OnlyComments = false
	m.cfg.OnlyStrings = false
	if got := m.codeFilterLabel(); got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}

	m.cfg.OnlyCode = true
	if got := m.codeFilterLabel(); got != "only-code" {
		t.Errorf("expected 'only-code', got %q", got)
	}

	m.cfg.OnlyCode = false
	m.cfg.OnlyComments = true
	if got := m.codeFilterLabel(); got != "only-comments" {
		t.Errorf("expected 'only-comments', got %q", got)
	}

	m.cfg.OnlyComments = false
	m.cfg.OnlyStrings = true
	if got := m.codeFilterLabel(); got != "only-strings" {
		t.Errorf("expected 'only-strings', got %q", got)
	}
}
