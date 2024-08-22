package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type VatNumber interface {
	// String returns the string representation for this vat number
	String() string
}

// A VAT number containing the land code and the numeric part
type vatNumber struct {
	vatString string
}

func (vat vatNumber) String() string {
	return vat.vatString
}

// Verifies if the format is valid for a Belgian vat number according to the checksum
// This does NOT check whether the VAT is actually registered
// Other countries could be verified based on regexes or api lookups
func checkVatValid(cleanedVat string, country string) error {
	if country != "BE" {
		return fmt.Errorf("Only Belgian VAT numbers are supported for now, not %q", country)
	}

	const vatNumberLength = 10
	const primeDivider = 97

	if len(cleanedVat) != vatNumberLength {
		return errors.New("The VAT should contain 10 digits")
	}

	firstPart, err := strconv.Atoi(cleanedVat[0:8])
	if err != nil {
		return err
	}
	checksum, err2 := strconv.Atoi(cleanedVat[8:10])
	if err2 != nil {
		return err2
	}
	calcValue := primeDivider - (firstPart - (firstPart/primeDivider)*primeDivider)
	if calcValue != checksum {
		return errors.New("The checksum does not match")
	}

	return nil
}

// NewBelgianVATNumber parses a vat number from any string.
func NewBelgianVatNumber(vat string) (VatNumber, error) {
	minLength := 2
	if len(vat) < minLength {
		return nil, errors.New("vat should have at least length 2")
	}
	reg, err := regexp.Compile(`\D`)
	if err != nil {
		return nil, errors.New("regexp could not be compiled")
	}
	cleanedVAT := reg.ReplaceAllLiteralString(strings.ToUpper(vat), "")

	country := strings.ToUpper(vat[0:2])
	if _, err := strconv.Atoi((country)); err == nil {
		country = "BE"
	}

	if err = checkVatValid(cleanedVAT, country); err != nil {
		return nil, fmt.Errorf("cannot parse vat number %q: %w", vat, err)
	}
	return vatNumber{country + cleanedVAT}, nil
}
