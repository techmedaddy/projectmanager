package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TextValue converts a nullable string pointer into a pgtype.Text value.
func TextValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

// TextPointer converts a pgtype.Text value into a nullable string pointer.
func TextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	text := value.String
	return &text
}

// DateValue converts a nullable time pointer into a pgtype.Date value.
func DateValue(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}

	return pgtype.Date{
		Time:  *value,
		Valid: true,
	}
}

// DatePointer converts a pgtype.Date value into a nullable time pointer.
func DatePointer(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}

	date := value.Time
	return &date
}
