package metrics

import (
	"fmt"
)

type Tags map[string]interface{}

func (t Tags) StringSlice() []string {
	tags := make([]string, len(t))
	for k, v := range t {
		tags = append(tags, fmt.Sprintf("%s:%v", k, v))
	}
	return tags
}

func (t Tags) Merge(tags ...Tags) Tags {
	return MergeTagsList(append(append([]Tags{}, t), tags...)...)
}

func (t Tags) WithTag(key string, value interface{}) Tags {
	return t.Merge(Tags{key: value})
}

func MergeTagsList(tagsList ...Tags) Tags {
	ret := Tags{}
	for _, tags := range tagsList {
		for k, v := range tags {
			ret[k] = v
		}
	}
	return ret
}
