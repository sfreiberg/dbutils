package dbutils

import (
	"database/sql"
	"fmt"
)

type InterfaceScanner struct {
	Name  string
	Value any
}

// InterfaceScanner implements the sql.Scan interface
func (s *InterfaceScanner) Scan(src any) error {
	switch src.(type) {
	default:
		s.Value = src
	case []uint8:
		s.Value = fmt.Sprintf("%s", src)
	}

	return nil
}

// Convert a row to a map. Expects rows.Next() to have been already called.
func ConvertRowToMap(r *sql.Rows) (map[string]any, error) {
	result := make(map[string]any)

	cols, err := r.Columns()
	if err != nil {
		return result, err
	}

	vals := make([]any, len(cols))
	for i := range vals {
		vals[i] = &InterfaceScanner{Name: cols[i]}
	}

	if err := r.Scan(vals...); err != nil {
		return nil, err
	}

	for i := range vals {
		scanner := *(vals[i].(*InterfaceScanner))
		result[scanner.Name] = scanner.Value
	}

	return result, nil
}
