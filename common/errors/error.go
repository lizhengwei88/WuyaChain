package errors

import (
	"errors"
	"fmt"
)

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// seeleError represents a seele error with code and message.
type wuyaError struct {
	code ErrorCode
	msg  string
}

// seeleParameterizedError represents a seele error with code and parameterized message.
// For type safe of common used business error, developer could define a concrete error to process.
type wuyaParameterizedError struct {
	wuyaError
	parameters []interface{}
}


func newWuyaError(code ErrorCode, msg string) error {
	return &wuyaError{code, msg}
}

// Error implements the error interface.
func (err *wuyaError) Error() string {
	return err.msg
}

// Create creates a seele error with specified error code and parameters.
func Create(code ErrorCode, args ...interface{}) error {
	errFormat, found := parameterizedErrors[code]
	if !found {
		return fmt.Errorf("system internal error, cannot find the error code %v", code)
	}

	return &wuyaParameterizedError{
		wuyaError: wuyaError{code, fmt.Sprintf(errFormat, args...)},
		parameters: args,
	}
}
