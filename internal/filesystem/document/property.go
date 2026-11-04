package document

import "fmt"

type Property interface {
	isProperty()
}

type Properties map[string]Property

type Text string

func (Text) isProperty() {}

type Float float64

func (Float) isProperty() {}

type List []string

func (List) isProperty() {}

type Integer int

func (Integer) isProperty() {}

func Normalize(v any) (Property, error) {
	switch v := v.(type) {
	case nil:
		return nil, nil

	case int:
		return Integer(v), nil

	case float64:
		return Float(v), nil

	case string:
		return Text(v), nil

	case []string:
		list := make(List, 0, len(v))
		for _, item := range v {
			list = append(list, fmt.Sprint(item))
		}
		return list, nil

	default:
		return nil, fmt.Errorf("unsupported type '%T'", v)
	}
}
