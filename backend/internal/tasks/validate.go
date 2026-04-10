package tasks

// IsValidStatus reports whether the status is one of the supported enum values.
func IsValidStatus(status Status) bool {
	switch status {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// IsValidPriority reports whether the priority is one of the supported enum
// values.
func IsValidPriority(priority Priority) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}
