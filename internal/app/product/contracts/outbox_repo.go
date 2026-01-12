package contracts

type OutboxRepo interface {
	InsertMut(
		eventID string,
		eventType string,
		aggregateID string,
		payload []byte,
	) any
}
