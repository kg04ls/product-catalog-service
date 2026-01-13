package m_outbox

import "cloud.google.com/go/spanner"

type Model struct{}

func (Model) InsertMut(row map[string]interface{}) *spanner.Mutation {
	return spanner.InsertMap(Table, row)
}
