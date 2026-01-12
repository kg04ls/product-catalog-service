package committer

import "product-catalog-service/internal/app/product/contracts"

type Plan struct {
	muts []contracts.Mutation
}

func NewPlan() *Plan {
	return &Plan{}
}

func (p *Plan) Add(m contracts.Mutation) {
	if m == nil {
		return
	}
	p.muts = append(p.muts, m)
}

func (p *Plan) Mutations() []contracts.Mutation {
	return p.muts
}
