package core

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/biter777/countries"
	"github.com/pingcap/log"
)

var (
	SupportedCountryIso2Codes = []string{"BE"}
	ErrVatCountryNotSupported = errors.New("VAT numbers for this country are not supported (yet)")
	ErrVatInvalidCode         = errors.New("The VAT number doesn't start with a valid code")
)

type VatNumber interface {
	// String returns the string representation for this vat number
	String() string
	// Display returns a formatted string representation for this vat number
	Display() string
}

// A VAT number containing the land code and the remaining part
type vatNumber struct {
	countryIso2 string
	vatString   string
}

func (vat vatNumber) String() string {
	return vat.countryIso2 + vat.vatString
}

func (vat vatNumber) Display() string {
	if vat.countryIso2 == countries.BE.Alpha2() && len(vat.vatString) == 10 {
		return fmt.Sprintf("%s %s.%s.%s", vat.countryIso2, vat.vatString[0:4], vat.vatString[4:7], vat.vatString[7:])
	}
	return vat.String()
}

// Verifies if the format is valid for a Belgian vat number according to the checksum
// This does NOT check whether the VAT is actually registered
// Other countries could be verified based on regexes or api lookups
func checkVatValid(cleanedVat string, countryIso2 string) error {
	if !slices.Contains(SupportedCountryIso2Codes, countryIso2) {
		return ErrVatCountryNotSupported
	}

	const vatNumberLength = 10
	const primeDivider = 97

	if len(cleanedVat) != vatNumberLength {
		return fmt.Errorf("The VAT should contain %d digits", vatNumberLength)
	}

	firstPart, err := strconv.Atoi(cleanedVat[0:8])
	if err != nil {
		return err
	}
	checksum, err := strconv.Atoi(cleanedVat[8:10])
	if err != nil {
		return err
	}
	calcValue := primeDivider - (firstPart - (firstPart/primeDivider)*primeDivider)
	if calcValue != checksum {
		return errors.New("The checksum does not match")
	}

	return nil
}

// NewVatNumber parses a vat number from any string.
func NewVatNumber(vat string) (VatNumber, error) {
	notAlphaNumericReg, err := regexp.Compile(`[^0-9a-zA-Z]`)
	if err != nil {
		log.Panic("invalid token regexp could not be compiled")
	}
	cleanedVAT := notAlphaNumericReg.ReplaceAllLiteralString(strings.ToUpper(vat), "")
	minLength := 2
	if len(cleanedVAT) < minLength {
		return nil, errors.New("vat should have at least length 2")
	}
	iso2 := cleanedVAT[:2]
	_, err = strconv.Atoi(iso2)
	if err == nil {
		// iso2 part is numeric, so we assume a Belgian VAT where the isocode was left out
		iso2 = countries.BE.Alpha2()
	} else {
		cleanedVAT = cleanedVAT[2:]
	}
	if countries.ByName(iso2) == countries.Unknown && !slices.Contains([]string{"XI", "EL", "EU"}, iso2) {
		return nil, ErrVatInvalidCode
	}

	if err = checkVatValid(cleanedVAT, iso2); err != nil {
		return nil, fmt.Errorf("cannot parse vat number %q: %w", vat, err)
	}
	return vatNumber{iso2, cleanedVAT}, nil
}
