package committer

import (
	"context"
)

type Committer interface {
	Apply(ctx context.Context, plan *Plan) error
}

type Nop struct{}

func (Nop) Apply(context.Context, *Plan) error { return nil }

var _ Committer = Nop{}
