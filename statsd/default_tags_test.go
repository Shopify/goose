package statsd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultTagsFromEnv(t *testing.T) {
	oldValueTags := os.Getenv(StatsdDefaultTags)
	defer restoreStatsdTagsValue()(oldValueTags)

	tt := []struct {
		name        string
		envVarValue string
		wantTags    []string
	}{
		{
			name:        "when not set, returns nil",
			envVarValue: "",
			wantTags:    nil,
		},
		{
			name:        "when no comma, returns one tag",
			envVarValue: "foo:bar",
			wantTags:    []string{"foo:bar"},
		},
		{
			name:        "when there is comma, break into multiple tags",
			envVarValue: "foo:bar,kube_namespace:production-unrestricted-k9asf9",
			wantTags:    []string{"foo:bar", "kube_namespace:production-unrestricted-k9asf9"},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			err := os.Setenv(StatsdDefaultTags, test.envVarValue)
			require.NoError(t, err)

			gotTags := defaultTagsFromEnv()

			assert.EqualValues(t, test.wantTags, gotTags)
		})
	}
}

func restoreStatsdTagsValue() func(oldValue string) {
	return func(oldValue string) {
		_ = os.Setenv("STATSD_DEFAULT_TAGS", oldValue)
	}
}
