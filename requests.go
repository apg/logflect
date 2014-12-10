package logflect

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

var (
	ErrInvalidRequest     = errors.New("Invalid request")
	ErrInvalidFilterParam = errors.New("Invalid filter parameter")
	ErrInvalidFilterField = errors.New("Invalid filter field")
	ErrInvalidFilterType  = errors.New("Invalid filter type")
)

type sessionRequest struct {
	DrainId string          `json:"drain_id"`
	Filters []sessionFilter `json:"filters,omitempty"`
}

type sessionFilter struct {
	Field string `json:"field,omitempty"`
	Type  string `json:"type,omitempty"`
	Param string `json:"param,omitempty"`
}

func readSessionRequest(body io.Reader) (string, Filter, error) {
	decoder := json.NewDecoder(body)
	request := sessionRequest{}

	if err := decoder.Decode(&request); err != nil {
		return "", nil, ErrInvalidRequest
	}

	if request.DrainId == "" {
		return "", nil, ErrInvalidRequest
	}

	switch len(request.Filters) {
	case 0:
		return request.DrainId, NewNoFilter(), nil
	case 1:
		if filter, err := request.Filters[0].ToFilter(); err != nil {
			return request.DrainId, nil, err
		} else {
			return request.DrainId, filter, nil
		}
	default:
		filters := make([]Filter, len(request.Filters))
		for i := 0; i < len(request.Filters); i++ {
			if f, err := request.Filters[i].ToFilter(); err != nil {
				return request.DrainId, nil, err
			} else {
				filters[i] = f
			}
		}
		return request.DrainId, NewComboFilter(filters...), nil
	}
}

func (sf *sessionFilter) ToFilter() (Filter, error) {
	if sf.Field == "" {
		return nil, ErrInvalidFilterField
	}
	if sf.Param == "" {
		return nil, ErrInvalidFilterParam
	}

	switch sf.Type {
	case "contains":
		return NewContainsFilter(sf.Field, sf.Param), nil
	case "regexp":
		if re, err := regexp.CompilePOSIX(sf.Param); err != nil {
			return nil, ErrInvalidFilterParam
		} else {
			return NewRegexpFilter(sf.Field, re), nil
		}
	default:
		return nil, ErrInvalidFilterType
	}
}
