package strcase

import (
	"math"
	"strings"
	"unicode"
)

// ToPascalCase transforms a string in any form to PascalCase.
func ToPascalCase(input string) string {
	return splitJoin(input, 0, 0, false)
}

// ToPascalGoCase transforms a string in any form to PascalCase, but with recognized initialisms in uppercase, matching the Go style.
func ToPascalGoCase(input string) string {
	return splitJoin(input, 0, 0, true)
}

// ToCamelCase transforms a string in any form to camelCase.
func ToCamelCase(input string) string {
	return splitJoin(input, 1, 0, false)
}

// ToCamelGoCase transforms a string in any form to camelCase, but with recognized initialisms in uppercase, matching the Go style.
func ToCamelGoCase(input string) string {
	return splitJoin(input, 1, 0, true)
}

// ToSnakeCase transforms a string in any form to snake_case.
func ToSnakeCase(input string) string {
	return splitJoin(input, math.MaxInt64, '_', false)
}

func allocateBuilder(input string, separator rune) *strings.Builder {
	var b strings.Builder
	length := len(input)
	if separator != 0 {
		// Heuristic to add about 25% buffer for separators
		// Not having perfect match isn't terrible, it will only result in a few more memory allocations.
		// Ex:
		//   foo_bar_baz: 9 original chars, 11 final. 9 * 5 / 4 = 11
		//   foo_id: 5 original chars, 6 final. 5 * 5 / 4 = 6
		//   a_b_c_d: 4 original chars, 7 final. 4 * 5 / 4 = 5, which will result in an extra allocation.
		length = length * 5 / 4
	}

	b.Grow(length)
	return &b
}

func splitJoin(input string, firstUpper int, separator rune, initialism bool) string {
	b := allocateBuilder(input, separator)
	var buf []rune
	var currentPartIndex int
	var lastCategory runeCategory

	// Flush the buffer as a part
	flush := func() {
		if len(buf) == 0 {
			// Nothing was added since last flush
			return
		}
		if separator != 0 && currentPartIndex > 0 {
			b.WriteRune(separator)
		}
		if currentPartIndex >= firstUpper {
			pascalPart(buf, initialism)
		}
		for _, r := range buf {
			b.WriteRune(r)
		}
		currentPartIndex++
		lastCategory = unknown
		buf = buf[0:0] // Clear buffer, but keep current allocation
	}

	for _, r := range input {
		switch cat := category(r); cat {
		case upper:
			if lastCategory != upper {
				flush()
			}
			lastCategory = cat
			buf = append(buf, unicode.ToLower(r))
		case lower, number:
			if (lastCategory > number) != (cat > number) {
				flush()
			}
			lastCategory = cat
			buf = append(buf, r)
		default:
			// separator
			flush()
		}
	}
	flush()

	return b.String()
}

// Convert to uppercase if initialism and `initialism` is true.
// Convert first rune to uppercase otherwise.
func pascalPart(part []rune, initialism bool) {
	if initialism && isInitialism(part) {
		for ri, r := range part {
			part[ri] = unicode.ToUpper(r)
		}
	} else {
		part[0] = unicode.ToUpper(part[0])
	}
}

type runeCategory int

const (
	unknown runeCategory = iota
	number
	lower
	upper
)

func category(r rune) runeCategory {
	switch {
	case unicode.IsLower(r):
		return lower
	case unicode.IsUpper(r):
		return upper
	case unicode.IsNumber(r):
		return number
	default:
		return unknown
	}
}
