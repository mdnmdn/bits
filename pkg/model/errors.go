package model

import "errors"

var (
	ErrUnsupportedMarket  = errors.New("unsupported market type")
	ErrUnsupportedFeature = errors.New("unsupported feature")
)
