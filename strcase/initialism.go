package strcase

import "sort"

var commonInitialisms [][]rune

func init() {
	// To follow go's convention of have acronyms in all caps, hard code a few of the common ones
	// Taken from https://github.com/golang/lint/blob/83fdc39ff7b56453e3793356bcff3070b9b96445/lint.go#L770-L809
	var initialisms = []string{
		"acl",
		"api",
		"ascii",
		"cpu",
		"css",
		"dns",
		"eof",
		"guid",
		"html",
		"http",
		"https",
		"id",
		"ip",
		"json",
		"lhs",
		"qps",
		"ram",
		"rhs",
		"rpc",
		"sla",
		"smtp",
		"sql",
		"ssh",
		"tcp",
		"tls",
		"ttl",
		"udp",
		"ui",
		"uid",
		"uuid",
		"uri",
		"url",
		"utf8",
		"vm",
		"xml",
		"xmpp",
		"xsrf",
		"xss",
	}
	sort.Strings(initialisms)

	for _, initialism := range initialisms {
		commonInitialisms = append(commonInitialisms, []rune(initialism))
	}
}

func IsInitialism(part string) bool {
	return isInitialism([]rune(part))
}

func isInitialism(part []rune) bool {
	// Adapted from sort.Search to benefit from the fact that we only deal with rune slices
	i := 0
	j := len(commonInitialisms)
out:
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j

		for k, r := range commonInitialisms[h] {
			switch {
			case len(part) < k+1 || part[k] < r:
				j = h
				continue out
			case part[k] > r:
				i = h + 1
				continue out
			}
		}
		return true
	}
	return false
}
