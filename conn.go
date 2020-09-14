package otsql

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"

	"go.opentelemetry.io/otel/label"
)

var (
	labelMissingContext = label.String("otsql.warning", "missing upstream context")
	labelDeprecated     = label.String("otsql.warning", "database driver uses deprecated features")
	labelUnknownArgs    = label.String("otsql.warning", "unknown args type")
)

// driver.Conn
type otConn struct {
	driver.Conn
	options TraceOptions
}

func (c otConn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	ctx, span, endTrace := startTrace(context.Background(), c.options, methodExec, query, args)
	span.SetAttributes(labelDeprecated, labelMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()

	execer, ok := c.Conn.(driver.Execer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	if res, err = execer.Exec(query, args); err != nil {
		return nil, err
	}
	return wrapResult(ctx, res, c.options), nil
}

func (c otConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	ctx, _, endTrace := startTrace(ctx, c.options, methodExec, query, args)
	defer func() {
		endTrace(ctx, err)
	}()

	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	if res, err = execer.ExecContext(ctx, query, args); err != nil {
		return nil, err
	}
	return wrapResult(ctx, res, c.options), nil
}

func (c otConn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	ctx, span, endTrace := startTrace(context.Background(), c.options, methodQuery, query, args)
	span.SetAttributes(labelDeprecated, labelMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()
	queryer, ok := c.Conn.(driver.Queryer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	if rows, err = queryer.Query(query, args); err != nil {
		return nil, err
	}
	return wrapRows(ctx, rows, c.options), nil
}

func (c otConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	ctx, _, endTrace := startTrace(ctx, c.options, methodQuery, query, args)
	defer func() {
		endTrace(ctx, err)
	}()

	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	if rows, err = queryer.QueryContext(ctx, query, args); err != nil {
		return nil, err
	}
	return wrapRows(ctx, rows, c.options), nil
}

func (c otConn) Ping(ctx context.Context) (err error) {
	ctx, _, traceFunc := startTrace(ctx, c.options, methodPing, "", nil)
	defer func() {
		traceFunc(ctx, err)
	}()

	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}

	return pinger.Ping(ctx)
}

func (c otConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	ctx, span, endTrace := startTrace(ctx, c.options, methodPrepare, query, nil)
	defer func() {
		endTrace(ctx, err)
	}()

	if prepare, ok := c.Conn.(driver.ConnPrepareContext); ok {
		if stmt, err = prepare.PrepareContext(ctx, query); err != nil {
			return nil, err
		}
	} else {
		span.SetAttributes(labelMissingContext)
		if stmt, err = c.Conn.Prepare(query); err != nil {
			return nil, err
		}
	}

	return wrapStmt(stmt, query, c.options), nil
}

func (c otConn) Prepare(query string) (stmt driver.Stmt, err error) {
	ctx, span, endTrace := startTrace(context.Background(), c.options, methodPrepare, query, nil)
	span.SetAttributes(labelMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()

	stmt, err = c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	return wrapStmt(stmt, query, c.options), nil
}

func (c otConn) Begin() (tx driver.Tx, err error) {
	ctx, span, endTrace := startTrace(context.Background(), c.options, methodBegin, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	span.SetAttributes(labelDeprecated, labelMissingContext)
	tx, err = c.Conn.Begin() // nolint
	if err != nil {
		return nil, err
	}
	return wrapTx(ctx, tx, c.options), nil
}

func (c otConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	ctx, span, endTrace := startTrace(ctx, c.options, methodBegin, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	if beginTx, ok := c.Conn.(driver.ConnBeginTx); ok {
		if tx, err = beginTx.BeginTx(ctx, opts); err != nil {
			return nil, err
		}
	} else {
		span.SetAttributes(labelDeprecated, labelMissingContext)
		if tx, err = c.Conn.Begin(); err != nil { // nolint
			return nil, err
		}
	}
	return wrapTx(ctx, tx, c.options), nil
}

func (c otConn) Close() error {
	return c.Conn.Close()
}

func wrapConn(conn driver.Conn, options TraceOptions) driver.Conn {
	return otConn{
		Conn:    conn,
		options: options,
	}

}

// driver.Result
type otResult struct {
	driver.Result
	ctx     context.Context
	options TraceOptions
}

func (r otResult) LastInsertId() (id int64, err error) {
	if !r.options.LastInsertID {
		return r.Result.LastInsertId()
	}

	ctx, span, endTrace := startTrace(r.ctx, r.options, methodLastInsertID, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()
	r.ctx = ctx
	id, err = r.Result.LastInsertId()
	span.SetAttributes(label.Int64("sql.last_insert_id", id))
	return
}

func (r otResult) RowsAffected() (cnt int64, err error) {
	if !r.options.RowsAffected {
		return r.Result.RowsAffected()
	}

	ctx, span, endTrace := startTrace(r.ctx, r.options, methodRowsAffected, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()
	r.ctx = ctx
	cnt, err = r.Result.RowsAffected()
	span.SetAttributes(label.Int64("sql.rows_affected", cnt))
	return
}

func wrapResult(ctx context.Context, parent driver.Result, options TraceOptions) driver.Result {
	return &otResult{
		Result:  parent,
		ctx:     ctx,
		options: options,
	}
}

// withRowsColumnTypeScanType is the same as the driver.RowsColumnTypeScanType
// interface except it omits the driver.Rows embedded interface.
// If the original driver.Rows implementation wrapped by ocsql supports
// RowsColumnTypeScanType we enable the original method implementation in the
// returned driver.Rows from wrapRows by doing a composition with ocRows.
type withRowsColumnTypeScanType interface {
	ColumnTypeScanType(index int) reflect.Type
}

// driver.Rows
type otRows struct {
	driver.Rows
	ctx     context.Context
	options TraceOptions
}

func (r otRows) Columns() []string {
	return r.Rows.Columns()
}

func (r otRows) Close() (err error) {
	if !r.options.RowsClose {
		return r.Rows.Close()
	}

	ctx, _, endTrace := startTrace(r.ctx, r.options, methodRowsClose, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	return r.Rows.Close()
}

func (r otRows) Next(dest []driver.Value) (err error) {
	if !r.options.RowsNext {
		return r.Rows.Next(dest)
	}

	ctx, _, endTrace := startTrace(r.ctx, r.options, methodRowsNext, "", nil)
	defer func() {
		if err == io.EOF {
			endTrace(ctx, nil)
		} else {
			endTrace(ctx, err)
		}
	}()

	err = r.Rows.Next(dest)
	return
}

func wrapRows(ctx context.Context, parent driver.Rows, options TraceOptions) driver.Rows {
	ts, isColumnTypeScan := parent.(driver.RowsColumnTypeScanType)
	r := otRows{
		Rows:    parent,
		ctx:     ctx,
		options: options,
	}
	if isColumnTypeScan {
		return struct {
			otRows
			withRowsColumnTypeScanType
		}{r, ts}
	}

	return r
}

type otStmt struct {
	driver.Stmt
	query   string
	options TraceOptions
}

func (s otStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	ctx, span, endTrace := startTrace(context.Background(), s.options, methodExec, s.query, args)
	span.SetAttributes(labelDeprecated, labelMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()

	res, err = s.Stmt.Exec(args) // nolint
	if err != nil {
		return nil, err
	}
	return wrapResult(ctx, res, s.options), nil
}

func (s otStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	ctx, _, endTrace := startTrace(ctx, s.options, methodExec, s.query, args)
	defer func() {
		endTrace(ctx, err)
	}()

	// we already tested driver when wrap stmt
	res, err = s.Stmt.(driver.StmtExecContext).ExecContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return wrapResult(ctx, res, s.options), nil
}

func (s otStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	ctx, span, endTrace := startTrace(context.Background(), s.options, methodQuery, s.query, args)
	span.SetAttributes(labelDeprecated, labelMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()

	rows, err = s.Stmt.Query(args) // nolint
	if err != nil {
		return nil, err
	}
	return wrapRows(ctx, rows, s.options), nil
}

func (s otStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	ctx, _, endTrace := startTrace(ctx, s.options, methodExec, s.query, args)
	defer func() {
		endTrace(ctx, err)
	}()

	// we already tested driver when wrap stmt
	rows, err = s.Stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return wrapRows(ctx, rows, s.options), nil
}

func wrapStmt(stmt driver.Stmt, query string, options TraceOptions) driver.Stmt {
	_, isExecCtx := stmt.(driver.StmtExecContext)
	_, isQueryCtx := stmt.(driver.StmtQueryContext)
	cc, isColumnConverter := stmt.(driver.ColumnConverter) // nolint
	nvc, isNamedValueChecker := stmt.(driver.NamedValueChecker)

	s := otStmt{
		Stmt:    stmt,
		query:   query,
		options: options,
	}

	switch {
	case !isExecCtx && !isQueryCtx && !isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
		}{s}

	case isExecCtx && !isQueryCtx && !isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
		}{s, s}
	case !isExecCtx && isQueryCtx && !isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
		}{s, s}
	case !isExecCtx && !isQueryCtx && isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.ColumnConverter
		}{s, cc}
	case !isExecCtx && !isQueryCtx && !isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.NamedValueChecker
		}{s, nvc}

	case isExecCtx && isQueryCtx && !isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
		}{s, s, s}
	case isExecCtx && !isQueryCtx && isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
		}{s, s, cc}
	case isExecCtx && !isQueryCtx && !isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.NamedValueChecker
		}{s, s, nvc}
	case !isExecCtx && isQueryCtx && isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, cc}
	case !isExecCtx && isQueryCtx && !isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, nvc}
	case !isExecCtx && !isQueryCtx && isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, cc, nvc}

	case isExecCtx && isQueryCtx && isColumnConverter && !isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, s, cc}
	case isExecCtx && isQueryCtx && !isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, s, nvc}
	case isExecCtx && !isQueryCtx && isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, cc, nvc}
	case !isExecCtx && isQueryCtx && isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, cc, nvc}

	case isExecCtx && isQueryCtx && isColumnConverter && isNamedValueChecker:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, s, cc, nvc}
	}

	panic("unreachable")
}

type otTx struct {
	driver.Tx
	ctx     context.Context
	options TraceOptions
}

func (t otTx) Commit() (err error) {
	ctx, _, endTrace := startTrace(t.ctx, t.options, methodCommit, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	return t.Tx.Commit()
}
func (t otTx) Rollback() (err error) {
	ctx, _, endTrace := startTrace(t.ctx, t.options, methodRollback, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	return t.Tx.Rollback()
}

func wrapTx(ctx context.Context, tx driver.Tx, options TraceOptions) driver.Tx {
	return otTx{
		Tx:      tx,
		ctx:     ctx,
		options: options,
	}
}
