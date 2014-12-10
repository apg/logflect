package logflect

import (
	"bytes"
	"testing"
)

var (
	TestSessionRequest_ValidOne = `
{
  "drain_id": "a.good.drain.id.with.single.filter",
  "filters": [
     {"field": "process", "type": "contains", "param": "python"}
  ]
}`

	TestSessionRequest_ValidMultiple = `
{
  "drain_id": "a.good.drain.id.with.multiple.filters",
  "filters": [
     {"field": "process", "type": "contains", "param": "python"},
     {"field": "process", "type": "regexp", "param": ".py$"}
  ]
}`

	TestSessionRequest_InvalidMissingType = `
{
  "drain_id": "a.bad.drain.with.missing.filter.type",
  "filters": [
     {"field": "process", "param": "python"}
  ]
}`

	TestSessionRequest_InvalidMissingParam = `
{
  "drain_id": "a.bad.drain.with.missing.filter.param",
  "filters": [
     {"field": "process", "type": "regexp"}
  ]
}`

	TestSessionRequest_InvalidMissingField = `
{
  "drain_id": "a.bad.drain.with.missing.filter.field",
  "filters": [
     {"param": "process", "type": "regexp"}
  ]
}`

	TestSessionRequest_InvalidMissingDrain = `
{
  "filters": [
     {"param": "process", "type": "regexp"}
  ]
}`

	TestSessionRequest_InvalidRegexp = `
{
  "drain_id": "a.bad.drain.with.invalid.filter.regexp",
  "filters": [
     {"field": "process", "type": "regexp", "param": "((a-z)+"}
  ]
}`

	TestSessionRequest_InvalidJsonBody = `
  "drain_id": "a.bad.drain.with.missing.filter.type",
  "filters": [
     {"field": "process", "type": "regexp", "param": "[a-z]+"}
  ]
}`
)

func TestreadSessionRequest_ValidSingle(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_ValidOne))
	drainId, filter, err := readSessionRequest(body)
	if err != nil {
		t.Errorf("unexpected error (%s)", err)
	}
	if drainId != "a.good.drain.id.with.single.filter" {
		t.Errorf("unexpected drain id: (%s)", drainId)
	}

	switch filter.(type) {
	case ContainsFilter:
		break
	default:
		t.Errorf("Expected ContainsFilter, found %s", filter)
	}
}

func TestreadSessionRequest_ValidMultiple(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_ValidMultiple))
	drainId, filter, err := readSessionRequest(body)
	if err != nil {
		t.Errorf("unexpected error (%s)", err)
	}
	if drainId != "a.good.drain.id.with.multiple.filters" {
		t.Errorf("unexpected drain id: (%s)", drainId)
	}

	switch filter.(type) {
	case ComboFilter:
		break
	default:
		t.Errorf("Expected ComboFilter, found %s", filter)
	}
}

func TestreadSessionRequest_InvalidMissingField(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidMissingField))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidFilterField {
		t.Errorf("unexpected error (%s)", err)
	}
}

func TestreadSessionRequest_InvalidMissingType(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidMissingType))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidFilterType {
		t.Errorf("unexpected error (%s)", err)
	}
}

func TestreadSessionRequest_InvalidMissingParam(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidMissingParam))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidFilterParam {
		t.Errorf("unexpected error (%s)", err)
	}
}

func TestreadSessionRequest_InvalidMissingDrain(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidMissingDrain))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidRequest {
		t.Errorf("unexpected error (%s)", err)
	}
}

func TestreadSessionRequest_InvalidRegexp(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidRegexp))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidFilterParam {
		t.Errorf("unexpected error (%s)", err)
	}
}

func TestreadSessionRequest_InvalidJsonBody(t *testing.T) {
	body := bytes.NewReader([]byte(TestSessionRequest_InvalidJsonBody))
	_, _, err := readSessionRequest(body)
	if err != ErrInvalidRequest {
		t.Errorf("unexpected error (%s)", err)
	}
}
