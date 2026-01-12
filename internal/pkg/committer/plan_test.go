package committer

import (
	"testing"

	"product-catalog-service/internal/app/product/contracts"
)

type mut struct{}

func (mut) IsMutation() {}

func TestPlan_AddNilIgnored(t *testing.T) {
	p := NewPlan()
	p.Add(nil)
	if got := len(p.Mutations()); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestPlan_AddKeepsOrder(t *testing.T) {
	p := NewPlan()
	m1 := contracts.Mutation(mut{})
	m2 := contracts.Mutation(mut{})

	p.Add(m1)
	p.Add(m2)

	ms := p.Mutations()
	if len(ms) != 2 {
		t.Fatalf("expected 2, got %d", len(ms))
	}
	if ms[0] != m1 {
		t.Fatalf("expected first mutation to be m1")
	}
	if ms[1] != m2 {
		t.Fatalf("expected second mutation to be m2")
	}
}
