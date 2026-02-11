package output

import "errors"

// errorWriter simulates a writer that always fails
type errorWriter struct{}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

// limitedErrorWriter allows a certain number of bytes to be written before failing
type limitedErrorWriter struct {
	limit   int
	written int
}

func (lw *limitedErrorWriter) Write(p []byte) (n int, err error) {
	if lw.written >= lw.limit {
		return 0, errors.New("write limit exceeded")
	}
	remaining := lw.limit - lw.written
	if len(p) <= remaining {
		lw.written += len(p)
		return len(p), nil
	}
	lw.written += remaining
	return remaining, errors.New("write limit exceeded")
}
