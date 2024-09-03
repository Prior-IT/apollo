package core

import (
	"errors"
	"fmt"
	"strconv"
)

type (
	ID int32
)

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

// ParseID parses a string into an ID.
func ParseID(id string) (ID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse  id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse  id:  ids cannot be negative")
	}
	ID := ID(int32(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse  id: %w", err)
	}
	return ID, nil
}