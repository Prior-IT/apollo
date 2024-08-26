package core_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

// Generates a valid belgian vat
func generateValidBelgianVAT(startWithOne bool) string {
	var base string
	if startWithOne {
		base = "1" + tests.Faker.DigitN(7)
	} else {
		base = "0" + tests.Faker.DigitN(7)
	}
	const primeDivider = 97

	firstPart, _ := strconv.Atoi(base)
	checksum := primeDivider - (firstPart - (firstPart/primeDivider)*primeDivider)
	return fmt.Sprintf("BE%s%02d", base, checksum)
}

func FuzzVatNumber(f *testing.F) {
	for _, seed := range []string{generateValidBelgianVAT(true), generateValidBelgianVAT(false), "BE 156", "XX ", "BE..2", ".BE.00", "hello world"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		vat, err := core.NewVatNumber(value)
		// We're not looking for valid vat here but rather for unexpected errors leading to a panic
		if err != nil {
			assert.Nil(t, vat, "If there is an error, there should not be a vat number")
		}
		if vat != nil {
			assert.Nil(t, err, "If there is a vat, there should be no error")
		}
	})
}

func TestVAT(t *testing.T) {
	t.Run("ok: valid belgian VAT starting with 0, case insensitive", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		vat1, err := core.NewVatNumber(value)
		assert.Nil(t, err)

		vat2, err := core.NewVatNumber(strings.ToLower(value))
		assert.Nil(t, err)

		assert.Equal(t, vat1, vat2)
	})

	t.Run("ok: valid belgian VAT with non-alphanumeric chars", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		nonAlphaNumericChars := "! @# .$%^ &*()_-=+[]{}|"
		var sb strings.Builder
		maxL := max(len(value), len(nonAlphaNumericChars))

		for idx := range maxL {
			if len(value) > idx {
				sb.WriteString(value[idx : idx+1])
			}
			if len(nonAlphaNumericChars) > idx {
				sb.WriteString(nonAlphaNumericChars[idx : idx+1])
			}
		}
		// example value: B!E 0@5#5 1$7%9^6 1&7*8()_-=+[]{}| -> BE0551796178
		_, err := core.NewVatNumber(sb.String())
		assert.Nil(t, err)
	})

	t.Run("ok: valid belgian VAT starting with 1, case insensitive", func(t *testing.T) {
		value := generateValidBelgianVAT(true)
		vat1, err := core.NewVatNumber(value)
		assert.Nil(t, err)

		vat2, err := core.NewVatNumber(strings.ToLower(value))
		assert.Nil(t, err)

		assert.Equal(t, vat1, vat2)
	})

	t.Run("ok: valid belgian VAT starting with 0 without BE", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		_, err := core.NewVatNumber(value[2:])
		assert.Nil(t, err)
	})

	t.Run("ok: valid belgian VAT starting with 1 without BE", func(t *testing.T) {
		value := generateValidBelgianVAT(true)
		_, err := core.NewVatNumber(value[2:])
		assert.Nil(t, err)
	})

	t.Run("err: invalid belgian VAT - starting with BE2", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		_, err := core.NewVatNumber("BE2" + value[3:])
		assert.NotNil(t, err)
	})

	t.Run("err: invalid VAT - unsupported country", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		_, err := core.NewVatNumber("NL" + value[2:])
		assert.ErrorIs(t, err, core.ErrVatCountryNotSupported)
	})

	t.Run("err: invalid VAT - non-iso country characters", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		_, err := core.NewVatNumber("XY" + value[2:])
		assert.ErrorIs(t, err, core.ErrVatInvalidCode)
	})

	t.Run("err: invalid Belgian VAT (too short)", func(t *testing.T) {
		value := generateValidBelgianVAT(false)
		_, err := core.NewVatNumber(value[:6])
		assert.NotNil(t, err)
	})
}
