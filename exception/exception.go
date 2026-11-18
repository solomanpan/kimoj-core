// Package exception holds the two kinds of failure of the framework.
//
// DomainError is a *value*: business rules return it inside a Result (no panics).
// InfrastructureError is *raised* (panic): the failure of an external system (database, broker, cache); the
// transaction scope rolls back, the HTTP layer maps it to 502/500.
package exception

import "fmt"

type Kind string

const (
	Validation Kind = "validation"
	NotFound   Kind = "not_found"
	Conflict   Kind = "conflict"
	Forbidden  Kind = "forbidden"
)

// DomainError is a business failure.
type DomainError struct {
	Kind    Kind
	Code    string
	Message string
}

func (e DomainError) Error() string { return e.Code + ": " + e.Message }

// NotFoundTemplate is the generic "not found" message (also the i18n key).
const NotFoundTemplate = "{entity} '{id}' not found"

func NewNotFound(entity, id string) DomainError {
	return DomainError{Kind: NotFound, Code: lower(entity) + "_not_found", Message: fmt.Sprintf("%s '%s' not found", entity, id)}
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

// InfrastructureError is a technical failure; it travels as a panic value.
type InfrastructureError struct {
	Code    string
	Message string
	Cause   error
}

func (e *InfrastructureError) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *InfrastructureError) Unwrap() error { return e.Cause }

// Raise panics with an InfrastructureError (use for the failure of an external system).
func Raise(code, message string, cause error) {
	panic(&InfrastructureError{Code: code, Message: message, Cause: cause})
}

// Must raises an InfrastructureError when err is not nil.
func Must(err error, code, message string) {
	if err != nil {
		Raise(code, message, err)
	}
}

func (k Kind) String() string { return string(k) }
