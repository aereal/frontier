package cli

import "errors"

var (
	ErrFunctionNameRequired = errors.New("function name is required")
	ErrFunctionPathRequired = errors.New("function path is required")
	ErrConfigPathRequired   = errors.New("config path is required")
)
