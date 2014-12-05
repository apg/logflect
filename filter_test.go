package logflect

import (
	"testing"
)

func TestPasses_ContainsFilter(t *testing.T) {
	msg := StrMessage("foo bar baz")

	filterOk := NewContainsFilter("", "bar")
	if !filterOk.Passes(msg) {
		t.Errorf("'bar' is contained in %s", msg)
	}

	filterBad := NewContainsFilter("", "qwijibo")
	if filterBad.Passes(msg) {
		t.Errorf("'qwijibo' is not contained in %s", msg)
	}
}

func TestPasses_ComboFilter(t *testing.T) {
	msg := StrMessage("foo bar baz")
	filterOk := NewComboFilter(NewContainsFilter("", "bar"), NewContainsFilter("", "foo"))

	if !filterOk.Passes(msg) {
		t.Errorf("'foo' and 'bar' are contained within %s", msg)
	}

	filterBad := NewComboFilter(NewContainsFilter("", "qwijibo"), NewContainsFilter("", "monkey"))
	if filterBad.Passes(msg) {
		t.Errorf("neither 'qwijibo' nor 'monkey' are contained within %s", msg)
	}
}
