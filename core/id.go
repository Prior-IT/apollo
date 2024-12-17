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

func (id *ID) UnmarshalText(text []byte) error {
	val, err := ParseID(string(text))
	if err != nil {
		return err
	}
	*id = val
	return nil
}

// ParseID parses a string into an ID.
func ParseID(id string) (ID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse id: ids cannot be negative")
	}
	return ID(int32(integerID)), nil
}
