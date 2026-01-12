package domain

type ChangeTracker struct {
	dirty map[string]struct{}
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{dirty: map[string]struct{}{}}
}

func (ct *ChangeTracker) MarkDirty(field string) {
	ct.dirty[field] = struct{}{}
}

func (ct *ChangeTracker) Dirty(field string) bool {
	_, ok := ct.dirty[field]
	return ok
}

func (ct *ChangeTracker) Any() bool {
	return len(ct.dirty) > 0
}

func (ct *ChangeTracker) Clear() {
	for k := range ct.dirty {
		delete(ct.dirty, k)
	}
}
