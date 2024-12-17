package core_test

import (
	"strings"
	"testing"

	"github.com/prior-it/apollo/core"
	"github.com/prior-it/apollo/tests"
	"github.com/stretchr/testify/assert"
)

func FuzzEmail(f *testing.F) {
	for _, seed := range []string{"@email.com", "test@email", "test@.com", "@", "", ".", ".abc"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		email, err := core.ParseEmailAddress(value)
		// We're not looking for valid email addresses here but rather for unexpected errors leading to a panic
		if err != nil {
			assert.Nil(t, email, "If there is an error, the e-mail address should be nil")
		}
		if email != nil {
			assert.Nil(t, err, "If there is an e-mail address, there should be no error")
		}
	})
}

func TestEmail(t *testing.T) {
	t.Run("ok: email should be case-insensitive", func(t *testing.T) {
		value := tests.Faker.Email()
		email1, err := core.ParseEmailAddress(value)
		assert.Nil(t, err)

		email2, err := core.ParseEmailAddress(strings.ToUpper(value))
		assert.Nil(t, err)

		assert.Equal(t, email1, email2)
	})

	t.Run("err: invalid email addresses", func(t *testing.T) {
		for _, value := range []string{
			"",
			".",
			"@",
			"@.",
			"@.com",
			"@abc.",
			"abc@.",
			"abc@xyz.",
			"abc@.xyz",
		} {
			_, err := core.ParseEmailAddress(value)
			assert.NotNil(t, err, "%q should not be a valid e-mailaddress", value)
		}
	})

	t.Run("ok: String() should not transform e-mail addresses", func(t *testing.T) {
		value := strings.ToLower(tests.Faker.Email())
		email, err := core.ParseEmailAddress(value)
		assert.Nil(t, err)

		assert.Equal(t, value, email.String())
	})
}
