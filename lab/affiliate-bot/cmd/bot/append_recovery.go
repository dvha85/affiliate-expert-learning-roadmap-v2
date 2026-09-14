package main

import (
	"errors"
	"fmt"
)

// publishedAppendUncertainty means a writer returned an error but the exact
// candidate is already visible through the canonical loader.  It is deliberately
// distinct from a retryable store error: callers must reconcile/retry the same
// immutable record rather than attempting a competing append.
//
// This only describes the single-file acknowledgement boundary after a
// successful canonical reload. It is not a claim about fsync/power loss,
// multi-file atomicity, or multi-host coordination.
type publishedAppendUncertainty struct {
	cause error
}

func (e *publishedAppendUncertainty) Error() string {
	return fmt.Sprintf("append acknowledgement uncertain after canonical record became visible: %v", e.cause)
}

func (e *publishedAppendUncertainty) Unwrap() error { return e.cause }

func isPublishedAppendUncertainty(err error) bool {
	var uncertain *publishedAppendUncertainty
	return errors.As(err, &uncertain)
}

// appendWithVisibleRecovery only reports publication uncertainty after the
// command's real canonical loader sees a byte/semantic-equal record. A loader
// failure or a different record preserves the original append error.
func appendWithVisibleRecovery[T any](appendLine func(string, []byte) error, path string, encoded []byte, candidate T, reload func() ([]T, error), equal func(T, T) bool) error {
	err := appendLine(path, encoded)
	if err == nil {
		return nil
	}
	values, reloadErr := reload()
	if reloadErr != nil {
		return err
	}
	for _, value := range values {
		if equal(value, candidate) {
			return &publishedAppendUncertainty{cause: err}
		}
	}
	return err
}
