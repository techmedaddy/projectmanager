package response

// FieldErrors stores field-specific validation failures in the API contract
// shape required by the assignment.
type FieldErrors map[string]string

// NewFieldErrors initializes a field error collection.
func NewFieldErrors() FieldErrors {
	return make(FieldErrors)
}

// Add records a validation error for a field.
func (f FieldErrors) Add(field, message string) {
	if f == nil {
		return
	}

	f[field] = message
}

// HasAny reports whether any validation failures have been collected.
func (f FieldErrors) HasAny() bool {
	return len(f) > 0
}
