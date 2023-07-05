package clierrors

import (
	"errors"
	"fmt"
)

var ErrValidation = errors.New("validation error")

func ValidationError(op string, details string) error {
	return fmt.Errorf("%w - %s: %s", ErrValidation, op, details)
}
