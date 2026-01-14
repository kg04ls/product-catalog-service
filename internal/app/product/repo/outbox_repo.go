package repo

import (
	//"encoding/json"

	"cloud.google.com/go/spanner"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/infra/spannerx"
	"product-catalog-service/internal/models/m_outbox"
	"product-catalog-service/internal/pkg/clock"
)

type OutboxRepo struct {
	model m_outbox.Model
	clock clock.Clock
}

func NewOutboxRepo(clk clock.Clock) *OutboxRepo {
	return &OutboxRepo{model: m_outbox.Model{}, clock: clk}
}

func (r *OutboxRepo) InsertMut(eventID, eventType, aggregateID string, payload []byte) contracts.Mutation {
	row := map[string]interface{}{
		m_outbox.EventID:     eventID,
		m_outbox.EventType:   eventType,
		m_outbox.AggregateID: aggregateID,
		m_outbox.Payload:     spanner.NullJSON{Value: string(payload), Valid: true},
		m_outbox.Status:      "NEW",
		m_outbox.CreatedAt:   r.clock.Now(),
	}

	return spannerx.Wrap(r.model.InsertMut(row))
}
