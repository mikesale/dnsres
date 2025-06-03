package circuitbreaker

import "errors"

// ErrCircuitOpen is returned when the circuit breaker is in the open state
var ErrCircuitOpen = errors.New("circuit breaker is open")
