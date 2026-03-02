package db2

import (
	"context"
	"database/sql/driver"
	"io"
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

// PrepareContext implements driver.ConnPrepareContext.
func (c *db2Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.inner.(driver.ConnPrepareContext); ok {
		stmt, err := pc.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}

		return &db2Stmt{inner: stmt}, nil
	}

	return c.Prepare(query)
}

// Query implements driver.Queryer to preserve the direct query optimization.
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

// ExecContext implements driver.StmtExecContext.
func (s *db2Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := s.inner.(driver.StmtExecContext); ok {
		return ec.ExecContext(ctx, args)
	}

	dargs := make([]driver.Value, len(args))
	for i, arg := range args {
		dargs[i] = arg.Value
	}

	return s.inner.Exec(dargs) //nolint:staticcheck
}

func (s *db2Stmt) Query(args []driver.Value) (driver.Rows, error) {
	rows, err := s.inner.Query(args) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	return newDB2Rows(rows), nil
}

// QueryContext implements driver.StmtQueryContext.
func (s *db2Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := s.inner.(driver.StmtQueryContext); ok {
		rows, err := qc.QueryContext(ctx, args)
		if err != nil {
			return nil, err
		}

		return newDB2Rows(rows), nil
	}

	dargs := make([]driver.Value, len(args))
	for i, arg := range args {
		dargs[i] = arg.Value
	}

	return s.Query(dargs)
}

// CheckNamedValue implements driver.NamedValueChecker.
func (s *db2Stmt) CheckNamedValue(nv *driver.NamedValue) error {
	if checker, ok := s.inner.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nv)
	}

	return driver.ErrSkip
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
		if b, ok := v.([]byte); ok {
			if r.textCols != nil && i < len(r.textCols) && r.textCols[i] {
				dest[i] = string(b)
			}
		}
	}

	return nil
}

// HasNextResultSet implements driver.RowsNextResultSet.
func (r *db2Rows) HasNextResultSet() bool {
	if hs, ok := r.inner.(driver.RowsNextResultSet); ok {
		return hs.HasNextResultSet()
	}

	return false
}

// NextResultSet implements driver.RowsNextResultSet.
func (r *db2Rows) NextResultSet() error {
	if ns, ok := r.inner.(driver.RowsNextResultSet); ok {
		r.inited = false // Reset for new result set

		return ns.NextResultSet()
	}

	return io.EOF
}

// ColumnTypeScanType implements driver.RowsColumnTypeScanType.
func (r *db2Rows) ColumnTypeScanType(index int) reflect.Type {
	if st, ok := r.inner.(driver.RowsColumnTypeScanType); ok {
		return st.ColumnTypeScanType(index)
	}

	return reflect.TypeOf(new(interface{}))
}

// ColumnTypeDatabaseTypeName implements driver.RowsColumnTypeDatabaseTypeName.
func (r *db2Rows) ColumnTypeDatabaseTypeName(index int) string {
	if dt, ok := r.inner.(driver.RowsColumnTypeDatabaseTypeName); ok {
		return dt.ColumnTypeDatabaseTypeName(index)
	}

	return ""
}

// ColumnTypeLength implements driver.RowsColumnTypeLength.
func (r *db2Rows) ColumnTypeLength(index int) (int64, bool) {
	if lt, ok := r.inner.(driver.RowsColumnTypeLength); ok {
		return lt.ColumnTypeLength(index)
	}

	return 0, false
}

// ColumnTypeNullable implements driver.RowsColumnTypeNullable.
func (r *db2Rows) ColumnTypeNullable(index int) (bool, bool) {
	if nt, ok := r.inner.(driver.RowsColumnTypeNullable); ok {
		return nt.ColumnTypeNullable(index)
	}

	return false, false
}

// ColumnTypePrecisionScale implements driver.RowsColumnTypePrecisionScale.
func (r *db2Rows) ColumnTypePrecisionScale(index int) (int64, int64, bool) {
	if ps, ok := r.inner.(driver.RowsColumnTypePrecisionScale); ok {
		return ps.ColumnTypePrecisionScale(index)
	}

	return 0, 0, false
}
