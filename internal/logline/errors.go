package logline

import "errors"

var (
	ErrFailedToAppendToJSON = errors.New("failed to append attribute to json")
)
