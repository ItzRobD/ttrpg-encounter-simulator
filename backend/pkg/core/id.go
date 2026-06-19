package core

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// ID represents a unique identifier for an entity, typically used as a string in UUID or numeric format.
type ID string

// String returns the string representation of the ID.
func (id ID) String() string { return string(id) }

// IsUUID checks if the ID is in a valid UUID format and returns true if it is, otherwise false.
func (id ID) IsUUID() bool {
	_, err := uuid.Parse(id.String())
	return err == nil
}

// IsInt determines if the ID can be parsed as an integer and returns true if it is valid, otherwise false.
func (id ID) IsInt() bool {
	_, err := strconv.Atoi(id.String())
	return err == nil
}

// Int converts the ID to an integer by parsing the string representation and returns the result.
func (id ID) Int() int {
	i, _ := strconv.Atoi(id.String())
	return i
}

// MakeID converts a given value to an ID. Supports string and various integer types, otherwise returns an empty ID.
func MakeID(v interface{}) ID {
	switch t := v.(type) {
	case string:
		return ID(t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ID(fmt.Sprintf("%d", t))
	default:
		return ID("")
	}
}
