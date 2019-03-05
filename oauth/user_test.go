package oauth

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUser_LogFields(t *testing.T) {
	origLogEmail := LogUserEmail
	origLogProfile := LogUserProfile
	defer func() {
		LogUserProfile = origLogProfile
		LogUserEmail = origLogEmail
	}()

	u := &User{
		Profile: "foo",
		Email:   "bar@example.com",
	}
	t.Run("log-none", func(t *testing.T) {
		LogUserEmail = false
		LogUserProfile = false
		assert.Equal(t, logrus.Fields{}, u.LogFields())
	})
	t.Run("log-email", func(t *testing.T) {
		LogUserEmail = true
		LogUserProfile = false
		assert.Equal(t, logrus.Fields{
			"email": "bar@example.com",
		}, u.LogFields())
	})
	t.Run("log-profile", func(t *testing.T) {
		LogUserEmail = false
		LogUserProfile = true
		assert.Equal(t, logrus.Fields{
			"profile": "foo",
		}, u.LogFields())
	})
	t.Run("log-both", func(t *testing.T) {
		LogUserEmail = true
		LogUserProfile = true
		assert.Equal(t, logrus.Fields{
			"profile": "foo",
			"email":   "bar@example.com",
		}, u.LogFields())
	})
}
