package logflect

import (
	"regexp"
	"strings"
)

type Filter interface {
	Passes(Message) bool
}

type NoFilter struct{}

// Ensures a message passes *only* if all `filters` also `Passes()`
type ComboFilter struct {
	filters []Filter
}

// Filters out messages which don't contain `needle` in Message's `field`
type ContainsFilter struct {
	field  string
	needle interface{}
}

// Filters out messages which don't match `regexp` in Message's `field`
type RegexpFilter struct {
	field  string
	regexp *regexp.Regexp
}

func NewNoFilter() Filter {
	return NoFilter{}
}

func NewComboFilter(fs ...Filter) Filter {
	return ComboFilter{
		filters: fs,
	}
}

func NewContainsFilter(field string, needle string) Filter {
	return ContainsFilter{
		field:  field,
		needle: needle,
	}
}

func NewRegexpFilter(field string, regexp *regexp.Regexp) Filter {
	return RegexpFilter{
		field:  field,
		regexp: regexp,
	}
}

func (f NoFilter) Passes(m Message) bool {
	return true
}

func (f ComboFilter) Passes(m Message) bool {
	for _, filter := range f.filters {
		if !filter.Passes(m) {
			return false
		}
	}
	return true
}

// Tests msg against filter to see if a given field contains `needle`
func (f ContainsFilter) Passes(m Message) bool {
	if value, ok := m.Field(f.field); ok {
		switch value.(type) {
		case StrMessage:
			if s, cok := f.needle.(string); cok {
				return strings.Contains(string(value.(StrMessage)), s)
			}
		default:
		}
	}

	return false
}

func (f RegexpFilter) Passes(m Message) bool {
	if value, ok := m.Field(f.field); ok {
		switch value.(type) {
		case StrMessage:
			return f.regexp.MatchString(string(value.(StrMessage)))
		default:
		}
	}

	return false
}
