package domain

import (
	"time"
)

type ProductStatus string

const (
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusActive   ProductStatus = "active"
)

type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus

	changes *ChangeTracker
	events  []DomainEvent
}

func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if id == "" {
		return nil, ErrInvalidProductID
	}
	if name == "" {
		return nil, ErrInvalidProductName
	}
	if category == "" {
		return nil, ErrInvalidCategory
	}
	if basePrice == nil {
		return nil, ErrInvalidMoney
	}

	t := now.UTC()

	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusInactive,
		changes:     NewChangeTracker(),
	}
	p.events = append(p.events, ProductCreatedEvent{ProductID: p.id, At: t})
	p.changes.MarkDirty(FieldName)
	p.changes.MarkDirty(FieldDescription)
	p.changes.MarkDirty(FieldCategory)
	p.changes.MarkDirty(FieldStatus)
	return p, nil
}

func HydrateProduct(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		changes:     NewChangeTracker(),
	}
}

func (p *Product) ID() string              { return p.id }
func (p *Product) Name() string            { return p.name }
func (p *Product) Description() string     { return p.description }
func (p *Product) Category() string        { return p.category }
func (p *Product) BasePrice() *Money       { return p.basePrice }
func (p *Product) Discount() *Discount     { return p.discount }
func (p *Product) Status() ProductStatus   { return p.status }
func (p *Product) Changes() *ChangeTracker { return p.changes }

func (p *Product) DomainEvents() []DomainEvent {
	out := make([]DomainEvent, len(p.events))
	copy(out, p.events)
	return out
}

func (p *Product) ClearDomainEvents() {
	p.events = nil
}

func (p *Product) UpdateDetails(name, description, category string, now time.Time) error {
	if name == "" {
		return ErrInvalidProductName
	}
	if category == "" {
		return ErrInvalidCategory
	}

	changed := false

	if p.name != name {
		p.name = name
		p.changes.MarkDirty(FieldName)
		changed = true
	}
	if p.description != description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
		changed = true
	}
	if p.category != category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
		changed = true
	}

	if changed {
		t := now.UTC()
		p.events = append(p.events, ProductUpdatedEvent{ProductID: p.id, At: t})
	}

	return nil
}

func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusActive {
		return nil
	}
	p.status = ProductStatusActive
	p.changes.MarkDirty(FieldStatus)
	t := now.UTC()
	p.events = append(p.events, ProductActivatedEvent{ProductID: p.id, At: t})
	return nil
}

func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusInactive {
		return nil
	}
	p.status = ProductStatusInactive
	p.changes.MarkDirty(FieldStatus)
	t := now.UTC()
	p.events = append(p.events, ProductDeactivatedEvent{ProductID: p.id, At: t})
	return nil
}

func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if discount == nil {
		return ErrInvalidDiscountPercent
	}
	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}
	if p.discount != nil && p.discount.Overlaps(discount) {
		return ErrDiscountOverlaps
	}

	p.discount = discount
	p.changes.MarkDirty(FieldDiscount)
	t := now.UTC()
	p.events = append(p.events, DiscountAppliedEvent{ProductID: p.id, At: t})
	return nil
}

func (p *Product) RemoveDiscount(now time.Time) error {
	if p.discount == nil {
		return nil
	}
	p.discount = nil
	p.changes.MarkDirty(FieldDiscount)
	t := now.UTC()
	p.events = append(p.events, DiscountRemovedEvent{ProductID: p.id, At: t})
	return nil
}
