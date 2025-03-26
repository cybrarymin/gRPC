package errors

import (
	"errors"
	"fmt"
)

// Standard error types
var (
	// ErrNotFound represents a resource not found error
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists represents a duplicate resource error
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput represents invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabase represents a generic database error
	ErrDatabase = errors.New("database error")

	// ErrTimeout represents an operation timeout
	ErrTimeout = errors.New("operation timed out")

	// ErrConcurrentModification represents a conflict in updating a resource
	ErrConcurrentModification = errors.New("resource was modified concurrently")

	// ErrInsufficientBalance is specific to bank operations
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidCurrency is specific to currency operations
	ErrInvalidCurrency = errors.New("invalid currency")
)

// NotFoundError returns a formatted not found error with the resource type and identifier
func NotFoundError(resourceType string, identifier string) error {
	return fmt.Errorf("%w: %s with identifier %s", ErrNotFound, resourceType, identifier)
}

// AlreadyExistsError returns a formatted already exists error
func AlreadyExistsError(resourceType string, identifier string) error {
	return fmt.Errorf("%w: %s with identifier %s", ErrAlreadyExists, resourceType, identifier)
}

// InvalidInputError returns a formatted invalid input error with details
func InvalidInputError(details string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, details)
}

// DatabaseError wraps a database error with additional context
func DatabaseError(err error, operation string) error {
	return fmt.Errorf("%w: %s operation failed: %v", ErrDatabase, operation, err)
}

// TimeoutError returns a formatted timeout error
func TimeoutError(operation string) error {
	return fmt.Errorf("%w: %s", ErrTimeout, operation)
}

// ConcurrentModificationError returns a formatted concurrent modification error
func ConcurrentModificationError(resourceType string, identifier string) error {
	return fmt.Errorf("%w: %s with identifier %s", ErrConcurrentModification, resourceType, identifier)
}

// InsufficientBalanceError returns a formatted insufficient balance error
func InsufficientBalanceError(accountID string, currentBalance float64, requiredAmount float64) error {
	return fmt.Errorf("%w: account %s has balance %.2f, requires %.2f",
		ErrInsufficientBalance, accountID, currentBalance, requiredAmount)
}

// InvalidCurrencyError returns a formatted invalid currency error
func InvalidCurrencyError(currency string) error {
	return fmt.Errorf("%w: %s", ErrInvalidCurrency, currency)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if the error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsDatabase checks if the error is a database error
func IsDatabase(err error) bool {
	return errors.Is(err, ErrDatabase)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsConcurrentModification checks if the error is a concurrent modification error
func IsConcurrentModification(err error) bool {
	return errors.Is(err, ErrConcurrentModification)
}

// IsInsufficientBalance checks if the error is an insufficient balance error
func IsInsufficientBalance(err error) bool {
	return errors.Is(err, ErrInsufficientBalance)
}

// IsInvalidCurrency checks if the error is an invalid currency error
func IsInvalidCurrency(err error) bool {
	return errors.Is(err, ErrInvalidCurrency)
}
