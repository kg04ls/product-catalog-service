package spannerx

import (
	"testing"

	"cloud.google.com/go/spanner"
)

func TestWrap_Nil(t *testing.T) {
	if got := Wrap(nil); got != nil {
		t.Fatalf("expected nil, got %T", got)
	}
}

func TestWrap_NonNil(t *testing.T) {
	sm := &spanner.Mutation{}
	got := Wrap(sm)
	if got == nil {
		t.Fatalf("expected non-nil")
	}

	m, ok := got.(Mutation)
	if !ok {
		t.Fatalf("expected spannerx.Mutation, got %T", got)
	}
	if m.M != sm {
		t.Fatalf("expected wrapped pointer to match")
	}
}
