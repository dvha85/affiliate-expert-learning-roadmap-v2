package main

import "time"

// Reject unrepresentable durations before multiplication; never clamp authority.
func checkedMissionSeconds(seconds int) (time.Duration, bool) {
	if seconds < 0 || int64(seconds) > int64((1<<63-1)/time.Second) {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}
