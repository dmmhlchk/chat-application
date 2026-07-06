package domain

import "errors"

type Kind uint8

const (
	KindNotFound     Kind = iota + 1 // 404 - Not Found
	KindConflict                     // 409 - Already exists, duplicate state
	KindValidation                   // 404 - Invalid input
	KindInvalidState                 // 422 - Business rule violation
)

type DomainError struct {
	Kind    Kind
	Message string
}

func newErr(kind Kind, message string) *DomainError {
	return &DomainError{
		Kind:    kind,
		Message: message,
	}
}

// Satisfying the "error" interface
func (de *DomainError) Error() string {
	return de.Message
}

func IsKind(err error, kind Kind) bool {
	var de *DomainError
	return errors.As(err, &de) && de.Kind == kind
}

// ___ Chat _________________________________________________________________
var (
	ErrChatNotFound = newErr(KindNotFound, "chat not found")
)
