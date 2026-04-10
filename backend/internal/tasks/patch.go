package tasks

import "encoding/json"

// NullableStringPatch tracks whether a string field was provided in a PATCH
// request and whether it was explicitly set to null.
type NullableStringPatch struct {
	Set   bool
	Value *string
}

func (p *NullableStringPatch) UnmarshalJSON(data []byte) error {
	p.Set = true

	if string(data) == "null" {
		p.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	p.Value = &value
	return nil
}

// NullableStatusPatch tracks PATCH updates to a task status field.
type NullableStatusPatch struct {
	Set   bool
	Value *Status
}

func (p *NullableStatusPatch) UnmarshalJSON(data []byte) error {
	p.Set = true

	if string(data) == "null" {
		p.Value = nil
		return nil
	}

	var value Status
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	p.Value = &value
	return nil
}

// NullablePriorityPatch tracks PATCH updates to a task priority field.
type NullablePriorityPatch struct {
	Set   bool
	Value *Priority
}

func (p *NullablePriorityPatch) UnmarshalJSON(data []byte) error {
	p.Set = true

	if string(data) == "null" {
		p.Value = nil
		return nil
	}

	var value Priority
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	p.Value = &value
	return nil
}
