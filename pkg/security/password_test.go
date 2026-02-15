package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashing(t *testing.T) {
	password := "my_secure_password"

	t.Run("Successful Hash and Verify", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		match := CheckPasswordHash(password, hash)
		assert.True(t, match)
	})

	t.Run("Wrong Password No Match", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err)

		match := CheckPasswordHash("wrong_password", hash)
		assert.False(t, match)
	})

	t.Run("Password Too Long Error", func(t *testing.T) {
		longPassword := "this_is_a_very_long_password_that_exceeds_the_seventy_two_character_limit_of_bcrypt_to_ensure_it_fails"
		hash, err := HashPassword(longPassword)
		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("Different Hashes for Same Password", func(t *testing.T) {
		// bcrypt uses random salt, so hashes should be different
		hash1, err := HashPassword(password)
		require.NoError(t, err)
		hash2, err := HashPassword(password)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})
}
