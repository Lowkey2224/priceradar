package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type StringArray []string

// encode strings to pgsql format
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil
	}

	elements := make([]string, len(a))
	for i, s := range a {

		escaped := strings.ReplaceAll(s, `"`, `\"`)
		elements[i] = `"` + escaped + `"`
	}

	return "{" + strings.Join(elements, ",") + "}", nil
}

// decode pgsql TEXT[] to array
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}

	var strVal string
	switch v := src.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		return fmt.Errorf("unsupported Scan type for StringArray: %T", src)
	}

	strVal = strings.TrimPrefix(strVal, "{")
	strVal = strings.TrimSuffix(strVal, "}")

	if strVal == "" {
		*a = StringArray{}
		return nil
	}

	rawItems := strings.Split(strVal, ",")
	items := make([]string, 0, len(rawItems))

	for _, item := range rawItems {

		item = strings.Trim(item, `"`)
		items = append(items, item)
	}

	*a = items
	return nil
}
