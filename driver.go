package db2

import (
	"database/sql/driver"
	"reflect"
)

var stringType = reflect.TypeOf("")

// db2Driver wraps the go_ibm_db driver to convert []byte to string
// for text columns, so k6 scripts receive proper strings instead of Uint8Arrays.
type db2Driver struct {
	inner driver.Driver
}

func (d *db2Driver) Open(dsn string) (driver.Conn, error) {
	conn, err := d.inner.Open(dsn)
	if err != nil {
		return nil, err
	}

	return &db2Conn{inner: conn}, nil
}

// db2Conn wraps a driver.Conn to return wrapped statements and rows.
type db2Conn struct {
	inner driver.Conn
}

func (c *db2Conn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.inner.Prepare(query)
	if err != nil {
		return nil, err
	}

	return &db2Stmt{inner: stmt}, nil
}

func (c *db2Conn) Close() error {
	return c.inner.Close()
}

func (c *db2Conn) Begin() (driver.Tx, error) {
	return c.inner.Begin() //nolint:staticcheck
}

// Query implements driver.Queryer to preserve the direct query path
// that go_ibm_db uses for queries without parameters.
func (c *db2Conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	q, ok := c.inner.(driver.Queryer) //nolint:staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}

	rows, err := q.Query(query, args) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	return newDB2Rows(rows), nil
}

// db2Stmt wraps a driver.Stmt to return wrapped rows from queries.
type db2Stmt struct {
	inner driver.Stmt
}

func (s *db2Stmt) Close() error {
	return s.inner.Close()
}

func (s *db2Stmt) NumInput() int {
	return s.inner.NumInput()
}

func (s *db2Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.inner.Exec(args) //nolint:staticcheck
}

func (s *db2Stmt) Query(args []driver.Value) (driver.Rows, error) {
	rows, err := s.inner.Query(args) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	return newDB2Rows(rows), nil
}

// db2Rows wraps driver.Rows to convert []byte to string for text columns.
type db2Rows struct {
	inner    driver.Rows
	textCols []bool
	inited   bool
}

func newDB2Rows(inner driver.Rows) *db2Rows {
	return &db2Rows{inner: inner}
}

// initTextCols detects which columns should have []byte converted to string
// by checking ColumnTypeScanType from the underlying driver.
func (r *db2Rows) initTextCols() {
	if r.inited {
		return
	}

	r.inited = true

	scanner, ok := r.inner.(driver.RowsColumnTypeScanType)
	if !ok {
		return
	}

	cols := r.inner.Columns()
	r.textCols = make([]bool, len(cols))

	for i := range cols {
		if scanner.ColumnTypeScanType(i) == stringType {
			r.textCols[i] = true
		}
	}
}

func (r *db2Rows) Columns() []string {
	return r.inner.Columns()
}

func (r *db2Rows) Close() error {
	return r.inner.Close()
}

// Next fetches the next row and converts []byte to string for text columns.
func (r *db2Rows) Next(dest []driver.Value) error {
	r.initTextCols()

	err := r.inner.Next(dest)
	if err != nil {
		return err
	}

	for i, v := range dest {
		if b, ok := v.([]byte); ok && r.textCols != nil && i < len(r.textCols) && r.textCols[i] {
			dest[i] = string(b)
		}
	}

	return nil
}
