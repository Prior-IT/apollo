package core_test

func Fuzz(f *testing.F) {
	for _, seed := range []string{"a-a", "a--b", "-", "----"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, id string) {
		_, err := domain.NewProjectID(id)
		// We're not looking for valid ids here but rather for unexpected errors
		if errors.Is(err, domain.ErrProjectIDHasInvalidCharacters) {
			t.Skip()
		}
		if err != nil {
			t.Errorf("Invalid project id input: %q", id)
		}
	})
}

func TestProjectID(t *testing.T) {
	t.Run("ok: id should be case-insensitive", func(t *testing.T) {
		id := gofakeit.LetterN(5)
		id1, err := domain.NewProjectID(id)
		assert.Nil(t, err)

		id2, err := domain.NewProjectID(strings.ToUpper(id))
		assert.Nil(t, err)

		assert.Equal(t, id1, id2)
	})
