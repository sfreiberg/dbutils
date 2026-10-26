package dbutils

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInterfaceScanner(t *testing.T) {
	now := time.Now()
	tests := []struct {
		src  any
		want any
	}{
		{[]byte("hello"), "hello"},
		{int64(42), int64(42)},
		{3.5, 3.5},
		{true, true},
		{nil, nil},
		{now, now},
	}
	for _, tt := range tests {
		s := &InterfaceScanner{}
		if err := s.Scan(tt.src); err != nil {
			t.Fatalf("Scan(%v): %v", tt.src, err)
		}
		if s.Value != tt.want {
			t.Errorf("Scan(%#v) = %#v, want %#v", tt.src, s.Value, tt.want)
		}
	}
}

func TestInterfaceScannerCopiesBytes(t *testing.T) {
	// Drivers may reuse the []byte passed to Scan, so the value must not
	// alias it.
	src := []byte("abc")
	s := &InterfaceScanner{}
	s.Scan(src)
	src[0] = 'z'
	if s.Value != "abc" {
		t.Fatalf("Value = %#v, want %q", s.Value, "abc")
	}
}

func TestConvertRowToMap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "score", "deleted"}).
			AddRow(int64(1), []byte("alice"), 9.5, nil).
			AddRow(int64(2), []byte("bob"), 7.0, nil),
	)

	rows, err := db.Query("SELECT id, name, score, deleted FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []map[string]any
	for rows.Next() {
		m, err := ConvertRowToMap(rows)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, m)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2", len(got))
	}
	want := map[string]any{"id": int64(1), "name": "alice", "score": 9.5, "deleted": nil}
	for k, v := range want {
		if got[0][k] != v {
			t.Errorf("row 0 %s = %#v, want %#v", k, got[0][k], v)
		}
	}
	if got[1]["name"] != "bob" {
		t.Errorf("row 1 name = %#v, want bob", got[1]["name"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConvertRowToMapScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	rows, err := db.Query("SELECT id FROM t")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// rows.Next() has not been called, so Scan fails; the error must be
	// returned instead of a map full of nils.
	m, err := ConvertRowToMap(rows)
	if err == nil {
		t.Fatalf("expected an error, got map %v", m)
	}
}
