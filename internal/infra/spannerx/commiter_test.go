package spannerx

import (
	"context"
	"testing"

	"product-catalog-service/internal/pkg/committer"
)

type badMutation struct{}

func (badMutation) IsMutation() {}

func TestCommitterApply_EmptyPlanOK(t *testing.T) {
	c := NewCommitter(nil)

	if err := c.Apply(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	p := committer.NewPlan()
	if err := c.Apply(context.Background(), p); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCommitterApply_UnsupportedMutationType(t *testing.T) {
	c := NewCommitter(nil)

	p := committer.NewPlan()
	p.Add(badMutation{})

	err := c.Apply(context.Background(), p)
	if err == nil {
		t.Fatalf("expected error")
	}
}
