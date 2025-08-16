package gadget

import (
	"fmt"
)

// Error types for better error handling
type (
	ErrGadgetNotFound struct {
		Name string
	}

	ErrGadgetAlreadyExists struct {
		Name string
	}

	ErrGadgetNotEnabled struct {
		Name string
	}

	ErrGadgetAlreadyEnabled struct {
		Name string
	}

	ErrInvalidGadgetName struct {
		Name string
	}

	ErrGadgetOperation struct {
		Operation string
		Name      string
		Err       error
	}
)

func (e ErrGadgetNotFound) Error() string {
	return fmt.Sprintf("gadget '%s' not found", e.Name)
}

func (e ErrGadgetAlreadyExists) Error() string {
	return fmt.Sprintf("gadget '%s' already exists", e.Name)
}

func (e ErrGadgetNotEnabled) Error() string {
	return fmt.Sprintf("gadget '%s' is not enabled", e.Name)
}

func (e ErrGadgetAlreadyEnabled) Error() string {
	return fmt.Sprintf("gadget '%s' is already enabled", e.Name)
}

func (e ErrInvalidGadgetName) Error() string {
	return fmt.Sprintf("invalid gadget name: '%s'", e.Name)
}

func (e ErrGadgetOperation) Error() string {
	return fmt.Sprintf("failed to %s gadget '%s': %v", e.Operation, e.Name, e.Err)
}

func (e ErrGadgetOperation) Unwrap() error {
	return e.Err
}

// IsGadgetNotFound checks if an error is ErrGadgetNotFound
func IsGadgetNotFound(err error) bool {
	_, ok := err.(ErrGadgetNotFound)
	return ok
}

// IsGadgetAlreadyExists checks if an error is ErrGadgetAlreadyExists
func IsGadgetAlreadyExists(err error) bool {
	_, ok := err.(ErrGadgetAlreadyExists)
	return ok
}

// IsGadgetNotEnabled checks if an error is ErrGadgetNotEnabled
func IsGadgetNotEnabled(err error) bool {
	_, ok := err.(ErrGadgetNotEnabled)
	return ok
}

// IsGadgetAlreadyEnabled checks if an error is ErrGadgetAlreadyEnabled
func IsGadgetAlreadyEnabled(err error) bool {
	_, ok := err.(ErrGadgetAlreadyEnabled)
	return ok
}
