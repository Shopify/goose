package strcase

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// splitjoin_l1_p1         	   38.1 ns/op	     16 B/op	  1 allocs/op
// IDToCamelCase_l1_p1     	   88.6 ns/op	     48 B/op	  3 allocs/op
// IDToSnakeCase_l1_p1     	   87.7 ns/op	     48 B/op	  3 allocs/op
//
// splitjoin_l1_p10        	    253 ns/op	    176 B/op	  2 allocs/op
// IDToCamelCase_l1_p10    	    421 ns/op	     72 B/op	  3 allocs/op
// IDToSnakeCase_l1_p10    	    269 ns/op	     72 B/op	  3 allocs/op
//
// splitjoin_l1_p100       	   2137 ns/op	   1904 B/op	  2 allocs/op
// IDToCamelCase_l1_p100   	   3503 ns/op	    248 B/op	  3 allocs/op
// IDToSnakeCase_l1_p100   	   1879 ns/op	    296 B/op	  3 allocs/op
//
// splitjoin_l10_p1        	   38.0 ns/op	     16 B/op	  1 allocs/op
// IDToCamelCase_l10_p1    	    247 ns/op	    168 B/op	  6 allocs/op
// IDToSnakeCase_l10_p1    	    248 ns/op	    168 B/op	  6 allocs/op
//
// splitjoin_l10_p10       	    278 ns/op	    272 B/op	  2 allocs/op
// IDToCamelCase_l10_p10   	   1140 ns/op	    264 B/op	  6 allocs/op
// IDToSnakeCase_l10_p10   	    979 ns/op	    296 B/op	  6 allocs/op
//
// splitjoin_l10_p100      	   2267 ns/op	   2816 B/op	  2 allocs/op
// IDToCamelCase_l10_p100  	   9538 ns/op	   1304 B/op	  6 allocs/op
// IDToSnakeCase_l10_p100  	   8147 ns/op	   1560 B/op	  6 allocs/op
//
// splitjoin_l100_p1       	   41.1 ns/op	     16 B/op	  1 allocs/op
// IDToCamelCase_l100_p1   	   1114 ns/op	   1160 B/op	  9 allocs/op
// IDToSnakeCase_l100_p1   	   1104 ns/op	   1176 B/op	  9 allocs/op
//
// splitjoin_l100_p10      	    446 ns/op	   1184 B/op	  2 allocs/op
// IDToCamelCase_l100_p10  	   7692 ns/op	   2072 B/op	  9 allocs/op
// IDToSnakeCase_l100_p10  	   7589 ns/op	   2328 B/op	  9 allocs/op
//
// splitjoin_l100_p100     	   3877 ns/op	  12032 B/op	  2 allocs/op
// IDToCamelCase_l100_p100 	  72671 ns/op	  11288 B/op	  9 allocs/op
// IDToSnakeCase_l100_p100 	  71673 ns/op	  14616 B/op	  9 allocs/op
func Benchmark_splitJoin(b *testing.B) {
	for _, length := range []int{1, 10, 100} {
		part := strings.Repeat("a", length)

		for _, count := range []int{1, 10, 100} {
			input := part + strings.Repeat("_"+part, count-1)

			// Baseline, split and join all parts
			b.Run(fmt.Sprintf("splitjoin_l%d_p%d", length, count), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					strings.Join(strings.Split(input, "_"), "")
				}
			})

			b.Run(fmt.Sprintf("IDToCamelCase_l%d_p%d", length, count), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ToCamelCase(input)
				}
			})

			b.Run(fmt.Sprintf("IDToSnakeCase_l%d_p%d", length, count), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					ToSnakeCase(input)
				}
			})
		}
	}
}

// lower	       5.03 ns/op	       0 B/op	       0 allocs/op
// upper	       5.81 ns/op	       0 B/op	       0 allocs/op
// number	       6.59 ns/op	       0 B/op	       0 allocs/op
// symbol	       6.58 ns/op	       0 B/op	       0 allocs/op
// 16_bits	       153 ns/op	       0 B/op	       0 allocs/op
// 32_bits	       160 ns/op	       0 B/op	       0 allocs/op
func Benchmark_category(b *testing.B) {
	tests := map[string][]rune{
		"lower":   {'a', 'b'},
		"upper":   {'A', 'B'},
		"number":  {'0', '1'},
		"symbol":  {'_', ' '},
		"16 bits": {'™', '∞', '•', 'Ω'},
		"32 bits": {'𠁂', '𠁄', '𠁔', '𠁑'},
	}
	for name, runes := range tests {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, r := range runes {
					category(r)
				}
			}
		})
	}
}

func Test_splitJoin(t *testing.T) {
	tests := []struct {
		input    string
		camel    string
		camelGo  string
		pascal   string
		pascalGo string
		snake    string
	}{
		{
			// everything empty
		},
		{
			input:  "a",
			pascal: "A",
			camel:  "a",
			snake:  "a",
		},
		{
			input:  "A",
			pascal: "A",
			camel:  "a",
			snake:  "a",
		},
		{
			input:  "a_a",
			pascal: "AA",
			camel:  "aA",
			snake:  "a_a",
		},
		{
			input:  "__a___a_",
			pascal: "AA",
			camel:  "aA",
			snake:  "a_a",
		},
		{
			input:  "aa_bbb",
			pascal: "AaBbb",
			camel:  "aaBbb",
			snake:  "aa_bbb",
		},
		{
			input:    "aa_id",
			pascal:   "AaId",
			pascalGo: "AaID",
			camel:    "aaId",
			camelGo:  "aaID",
			snake:    "aa_id",
		},
		{
			input:  "fooBar",
			pascal: "FooBar",
			camel:  "fooBar",
			snake:  "foo_bar",
		},
		{
			input:  "FooBAR",
			pascal: "FooBar",
			camel:  "fooBar",
			snake:  "foo_bar",
		},
		{
			input:    "fooUrl",
			pascal:   "FooUrl",
			pascalGo: "FooURL",
			camel:    "fooUrl",
			camelGo:  "fooURL",
			snake:    "foo_url",
		},
		{
			input:    "fooURL",
			pascal:   "FooUrl",
			pascalGo: "FooURL",
			camel:    "fooUrl",
			camelGo:  "fooURL",
			snake:    "foo_url",
		},
		{
			input:    "url10",
			pascal:   "Url10",
			pascalGo: "URL10",
			camel:    "url10",
			snake:    "url_10",
		},
		{
			input:    "url_id",
			pascal:   "UrlId",
			pascalGo: "URLID",
			camel:    "urlId",
			camelGo:  "urlID",
			snake:    "url_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.pascal, ToPascalCase(tt.input))
			require.Equal(t, tt.camel, ToCamelCase(tt.input))
			require.Equal(t, tt.snake, ToSnakeCase(tt.input))

			if tt.pascalGo == "" {
				tt.pascalGo = tt.pascal
			}
			require.Equal(t, tt.pascalGo, ToPascalGoCase(tt.input))

			if tt.camelGo == "" {
				tt.camelGo = tt.camel
			}
			require.Equal(t, tt.camelGo, ToCamelGoCase(tt.input))
		})
	}
}
