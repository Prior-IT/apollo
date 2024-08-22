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
func generateValidVAT(startWithOne bool) string {
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

func TestVAT(t *testing.T) {
	t.Run("ok: valid belgian VAT starting with 0, case insensitive", func(t *testing.T) {
		value := generateValidVAT(false)
		vat1, err := core.NewBelgianVatNumber(value)
		assert.Nil(t, err)

		vat2, err := core.NewBelgianVatNumber(strings.ToLower(value))
		assert.Nil(t, err)

		assert.Equal(t, vat1, vat2)
	})

	t.Run("ok: valid belgian VAT starting with 1, case insensitive", func(t *testing.T) {
		value := generateValidVAT(true)
		vat1, err := core.NewBelgianVatNumber(value)
		assert.Nil(t, err)

		vat2, err := core.NewBelgianVatNumber(strings.ToLower(value))
		assert.Nil(t, err)

		assert.Equal(t, vat1, vat2)
	})

	t.Run("ok: valid belgian VAT starting with 0 without BE", func(t *testing.T) {
		value := generateValidVAT(false)
		_, err := core.NewBelgianVatNumber(value[2:])
		assert.Nil(t, err)
	})

	t.Run("ok: valid belgian VAT starting with 1 without BE", func(t *testing.T) {
		value := generateValidVAT(true)
		_, err := core.NewBelgianVatNumber(value[2:])
		assert.Nil(t, err)
	})

	t.Run("nok: invalid belgian VAT - starting with BE2", func(t *testing.T) {
		value := generateValidVAT(false)
		_, err := core.NewBelgianVatNumber("BE2" + value[3:])
		assert.NotNil(t, err)
	})

	t.Run("nok: invalid VAT - starting with NL", func(t *testing.T) {
		value := generateValidVAT(false)
		_, err := core.NewBelgianVatNumber("NL" + value[2:])
		assert.NotNil(t, err)
	})

	t.Run("nok: invalid VAT (too short)", func(t *testing.T) {
		value := generateValidVAT(false)
		_, err := core.NewBelgianVatNumber(value[:6])
		assert.NotNil(t, err)
	})
}
