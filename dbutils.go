package dbutils

import "database/sql"

type InterfaceScanner struct {
	Name  string
	Value interface{}
}

// InterfaceScanner implements the sql.Scan interface
func (s *InterfaceScanner) Scan(src interface{}) error {
	s.Value = src
	return nil
}

// Convert a row to a map. Expects rows.Next() to have been already called.
func ConvertRowToMap(r *sql.Rows) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	cols, err := r.Columns()
	if err != nil {
		return result, err
	}

	vals := make([]interface{}, len(cols))
	for i := range vals {
		vals[i] = &InterfaceScanner{Name: cols[i]}
	}

	r.Scan(vals...)

	for i := range vals {
		scanner := *(vals[i].(*InterfaceScanner))
		result[scanner.Name] = scanner.Value
	}

	return result, nil
}
