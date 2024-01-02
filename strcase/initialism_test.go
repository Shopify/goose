package strcase

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsInitialism(t *testing.T) {
	tests := []struct {
		input  string
		output bool
	}{
		{"", false},
		{"foo", false},
		{"id", true},
		{"url", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.output, IsInitialism(tt.input))
		})
	}
}

// foo	        18.3 ns/op	       0 B/op	       0 allocs/op
// url	        22.2 ns/op	       0 B/op	       0 allocs/op
// acl	        22.4 ns/op	       0 B/op	       0 allocs/op
func BenchmarkIsInitialism(b *testing.B) {
	for _, input := range []string{"foo", "url", string(commonInitialisms[0])} {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsInitialism(input)
			}
		})
	}
}
