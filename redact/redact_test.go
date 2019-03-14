package redact

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	testCases := []struct {
		testName       string
		input          map[string]interface{}
		expectedResult map[string]interface{}
	}{
		{
			"empty map",
			map[string]interface{}{},
			map[string]interface{}{},
		},

		{
			"sensitive",
			map[string]interface{}{
				"AuTHORization": "foo",
				"keep this":     123123123,
				"some_password": "s3cr3t",
				"idToken":       "aldsfjalsdfj",
			},
			map[string]interface{}{
				"AuTHORization": "[FILTERED]",
				"keep this":     123123123,
				"some_password": "[FILTERED]",
				"idToken":       "[FILTERED]",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			result := Map(tc.input)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestMapCustomSubstrings(t *testing.T) {
	oldSensitiveSubstrings := make([]string, len(GlobalSensitiveSubstrings))
	copy(oldSensitiveSubstrings, GlobalSensitiveSubstrings)
	defer func() {
		GlobalSensitiveSubstrings = oldSensitiveSubstrings
	}()

	data := map[string]interface{}{"foo": "bar", "123": 456}

	require.Equal(t, map[string]interface{}{"foo": "bar", "123": 456}, Map(data))

	AddSensitiveSubstring("Foo")

	require.Equal(t, map[string]interface{}{"foo": "[FILTERED]", "123": 456}, Map(data))

	require.Equal(t, []string{
		"authorization",
		"cookie",
		"token",
		"password",
		"secret",
		"foo",
	}, GlobalSensitiveSubstrings)
}

func TestHeaders(t *testing.T) {
	testCases := []struct {
		testName       string
		input          http.Header
		expectedResult map[string]string
	}{
		{
			"empty header",
			http.Header{},
			map[string]string{},
		},

		{
			"sensitive",
			http.Header{
				"AuTHORization": []string{"foo"},
				"keep this":     []string{"123123123"},
				"some_password": []string{"s3cr3t"},
				"idToken":       []string{"aldsfjalsdfj"},
				"Many-Things":   []string{"one", "two", "three"},
			},
			map[string]string{
				"AuTHORization": "[FILTERED]",
				"keep this":     "123123123",
				"some_password": "[FILTERED]",
				"idToken":       "[FILTERED]",
				"Many-Things":   "one,two,three",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			result := Headers(tc.input)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
