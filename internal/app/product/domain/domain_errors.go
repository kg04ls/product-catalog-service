package domain

import "errors"

var (
	ErrInvalidProductID       = errors.New("invalid product ID")
	ErrInvalidProductName     = errors.New("invalid product name")
	ErrInvalidCategory        = errors.New("invalid category")
	ErrInvalidMoney           = errors.New("invalid money")
	ErrInvalidDiscountPercent = errors.New("invalid discount percent")
	ErrInvalidDiscountPeriod  = errors.New("invalid discount period")
	ErrDiscountOverlaps       = errors.New("discount overlaps existing")
	ErrProductNotActive       = errors.New("product not active")
)
