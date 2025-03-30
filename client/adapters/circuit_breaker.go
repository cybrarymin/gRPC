package client_adapters

import (
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	circuitOpenState     string = "Open"
	circuitClosedState   string = "Closed"
	circuitHalfOpenState string = "HalfOpen"
)

var (
	ErrFuncTimeout        = errors.New("timeout executing function")
	ErrCircuitOpen        = errors.New("circuit open, request blocked")
	ErrUknownCircuitState = errors.New("unknown circuit state")
)

type CircuitBreaker struct {
	mu                   sync.Mutex // Guards the circuit breaker state because cicuit breaker can be used by multiple go routines
	state                string     // Current state of the circuit breaker
	failureCount         int        // Number of consecutive failures
	lastFailureTime      time.Time  // Time of the last failure
	halfOpenSuccessCount int        // Successful requests in half-open state

	failureThreshold    int           // number of failures to trigger open state
	recoveryTime        time.Duration // Wait time before changin from open state to half-open
	halfOpenMaxRequests int           // maximum number of requests needs to be succeed to turn from half-open to closed state
	timeout             time.Duration // Timeout for requests
	logger              *zerolog.Logger
}

func NewCircuitBreaker(failureThreshold int, recoveryTime time.Duration, halfOpenMaxRequests int, timeout time.Duration, logger *zerolog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:                circuitClosedState,
		failureCount:         0,
		halfOpenSuccessCount: 0,
		failureThreshold:     failureThreshold,
		recoveryTime:         recoveryTime,
		halfOpenMaxRequests:  halfOpenMaxRequests,
		timeout:              timeout,
		logger:               logger,
	}
}

func (cb *CircuitBreaker) Call(fn func() (any, error)) (any, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case circuitClosedState:
		return cb.HandleClosedState(fn)
	case circuitOpenState:
		return cb.HandleOpenState()
	case circuitHalfOpenState:
		return cb.HandleHalfOpenState(fn)
	default:
		return nil, ErrUknownCircuitState
	}
}

func (cb *CircuitBreaker) HandleClosedState(fn func() (any, error)) (any, error) {
	response, err := runWithTimeout(fn, cb.timeout)
	if err != nil {

		cb.failureCount++
		cb.lastFailureTime = time.Now()
		if cb.failureCount >= cb.failureThreshold {
			cb.state = circuitOpenState

		}
		return nil, err
	}
	cb.resetCircuit() // reset the failure counter
	return response, nil
}

func (cb *CircuitBreaker) HandleOpenState() (any, error) {
	if time.Since(cb.lastFailureTime) >= cb.recoveryTime {
		cb.logger.Info().Msg("circuit breaker state is changing to half-open")
		cb.state = circuitHalfOpenState
		cb.failureCount = 0
		cb.halfOpenSuccessCount = 0
		return nil, nil
	}
	cb.logger.Warn().Msg("circuit breaker state is open. rejecting the request")
	return nil, ErrCircuitOpen
}

func (cb *CircuitBreaker) HandleHalfOpenState(fn func() (any, error)) (any, error) {
	result, err := runWithTimeout(fn, cb.timeout)
	if err != nil {

		cb.logger.Warn().Msg("circuit breaker state is changing from half-open to open")
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		cb.state = circuitOpenState

		return nil, err
	}
	cb.halfOpenSuccessCount++
	cb.logger.Info().
		Int("half-open max success count", cb.halfOpenMaxRequests).
		Int("half-open current success count", cb.halfOpenSuccessCount).
		Msg("circuit breaker request succeeded in half-open state")
	if cb.halfOpenSuccessCount >= cb.halfOpenMaxRequests {
		cb.logger.Info().Msg("circuit breaker state is changing from half-open to closed")
		cb.resetCircuit()
	}
	return result, nil
}

func runWithTimeout(fn func() (any, error), timeout time.Duration) (any, error) {
	resultChan := make(chan any)
	errChan := make(chan error)
	go func() {
		result, err := fn()
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()
	select {
	case err := <-errChan:
		return nil, err
	case resp := <-resultChan:
		return resp, nil
	case <-time.After(timeout):
		return nil, ErrFuncTimeout
	}
}

func (cb *CircuitBreaker) resetCircuit() {
	cb.failureCount = 0
	cb.state = circuitClosedState
	cb.logger.Debug().Msg("circuit breaker state reseted to close")
}
