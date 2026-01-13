package repo

import (
	//"encoding/json"
	"time"

	"cloud.google.com/go/spanner"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/infra/spannerx"
	"product-catalog-service/internal/models/m_outbox"
)

type OutboxRepo struct {
	model m_outbox.Model
}

func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{model: m_outbox.Model{}}
}

func (r *OutboxRepo) InsertMut(eventID, eventType, aggregateID string, payload []byte) contracts.Mutation {
	row := map[string]interface{}{
		m_outbox.EventID:     eventID,
		m_outbox.EventType:   eventType,
		m_outbox.AggregateID: aggregateID,
		m_outbox.Payload:     spanner.NullJSON{Value: string(payload), Valid: true},
		m_outbox.Status:      "NEW",
		m_outbox.CreatedAt:   time.Now().UTC(),
	}

	return spannerx.Wrap(r.model.InsertMut(row))
}
