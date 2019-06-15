package framework

import "strings"

// AggregatedError aggregates multiple errors.
type AggregatedError struct {
	Errors []error
}

// Error implements error
func (e *AggregatedError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	msg := make([]string, len(e.Errors)+1)
	msg[0] = "Multiple errors:"
	for n, err := range e.Errors {
		msg[n+1] = err.Error()
	}
	return strings.Join(msg, "\n")
}

// Add adds errors to be aggregated. nil will be skipped.
func (e *AggregatedError) Add(errs ...error) *AggregatedError {
	for _, err := range errs {
		if err != nil {
			e.Errors = append(e.Errors, err)
		}
	}
	return e
}

// Aggregate returns aggregated error if any error happened.
func (e *AggregatedError) Aggregate() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}
