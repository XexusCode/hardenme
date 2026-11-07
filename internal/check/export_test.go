package check

import "time"

// SetNow replaces the clock used by builtins and returns a restore func.
func SetNow(f func() time.Time) func() {
	prev := now
	now = f
	return func() { now = prev }
}
