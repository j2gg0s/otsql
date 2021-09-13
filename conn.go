package otsql

import (
	"context"
	"database/sql/driver"
	"reflect"
)

// driver.Conn
type otConn struct {
	driver.Conn
	*Options

	connID string
}

func (c otConn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	evt := newEvent(c.Options, c.connID, MethodExec, query, args)

	ctx := before(c.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	execer, ok := c.Conn.(driver.Execer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	if res, err = execer.Exec(query, args); err != nil {
		return nil, err
	}

	return wrapResult(ctx, c.connID, res, c.Options), nil
}

func (c otConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	evt := newEvent(c.Options, c.connID, MethodExec, query, args)

	ctx = before(c.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	if res, err = execer.ExecContext(ctx, query, args); err != nil {
		return nil, err
	}
	return wrapResult(ctx, c.connID, res, c.Options), nil
}

func (c otConn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	evt := newEvent(c.Options, c.connID, MethodQuery, query, args)

	ctx := before(c.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	queryer, ok := c.Conn.(driver.Queryer) // nolint
	if !ok {
		return nil, driver.ErrSkip
	}
	if rows, err = queryer.Query(query, args); err != nil {
		return nil, err
	}
	return wrapRows(ctx, c.connID, rows, c.Options), nil
}

func (c otConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	evt := newEvent(c.Options, c.connID, MethodQuery, query, args)

	ctx = before(c.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	if rows, err = queryer.QueryContext(ctx, query, args); err != nil {
		return nil, err
	}
	return wrapRows(ctx, c.connID, rows, c.Options), nil
}

func (c otConn) Ping(ctx context.Context) (err error) {
	evt := newEvent(c.Options, c.connID, MethodPing, "", nil)

	ctx = before(c.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}

	return pinger.Ping(ctx)
}

func (c otConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	evt := newEvent(c.Options, c.connID, MethodPrepare, query, nil)

	ctx = before(c.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	if prepare, ok := c.Conn.(driver.ConnPrepareContext); ok {
		if stmt, err = prepare.PrepareContext(ctx, query); err != nil {
			return nil, err
		}
	} else {
		if stmt, err = c.Conn.Prepare(query); err != nil {
			return nil, err
		}
	}

	return wrapStmt(c.connID, stmt, query, c.Options), nil
}

func (c otConn) Prepare(query string) (stmt driver.Stmt, err error) {
	evt := newEvent(c.Options, c.connID, MethodPrepare, query, nil)
	ctx := before(c.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	stmt, err = c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	return wrapStmt(c.connID, stmt, query, c.Options), nil
}

func (c otConn) Begin() (tx driver.Tx, err error) {
	evt := newEvent(c.Options, c.connID, MethodBegin, "", nil)
	ctx := before(c.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	tx, err = c.Conn.Begin() // nolint
	if err != nil {
		return nil, err
	}
	return wrapTx(ctx, c.connID, tx, c.Options), nil
}

func (c otConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	evt := newEvent(c.Options, c.connID, MethodBegin, "", nil)
	ctx = before(c.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	if beginTx, ok := c.Conn.(driver.ConnBeginTx); ok {
		if tx, err = beginTx.BeginTx(ctx, opts); err != nil {
			return nil, err
		}
	} else {
		if tx, err = c.Conn.Begin(); err != nil { // nolint
			return nil, err
		}
	}
	return wrapTx(ctx, c.connID, tx, c.Options), nil
}

func (c otConn) Close() (err error) {
	evt := newEvent(c.Options, c.connID, MethodCloseConn, "", nil)
	ctx := before(c.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(c.Hooks, ctx, evt)
	}()

	return c.Conn.Close()
}

func (c otConn) ResetSession(ctx context.Context) (err error) {
	if c.ResetSessionB {
		evt := newEvent(c.Options, c.connID, MethodResetSession, "", nil)
		ctx = before(c.Hooks, ctx, evt)
		defer func() {
			evt.Err = err
			after(c.Hooks, ctx, evt)
		}()
	}

	if cr, ok := c.Conn.(driver.SessionResetter); ok {
		return cr.ResetSession(ctx)
	}

	return nil
}

func wrapConn(connID string, conn driver.Conn, o *Options) driver.Conn {
	return otConn{
		Conn:    conn,
		Options: o,
		connID:  connID,
	}
}

// driver.Result
type otResult struct {
	driver.Result
	*Options
	ctx    context.Context
	connID string
}

func (r otResult) LastInsertId() (id int64, err error) {
	if !r.LastInsertIdB {
		return r.Result.LastInsertId()
	}

	evt := newEvent(r.Options, r.connID, MethodLastInsertId, "", nil)
	r.ctx = before(r.Hooks, r.ctx, evt)
	defer func() {
		evt.Err = err
		after(r.Hooks, r.ctx, evt)
	}()

	id, err = r.Result.LastInsertId()
	return
}

func (r otResult) RowsAffected() (cnt int64, err error) {
	if !r.RowsAffectedB {
		return r.Result.RowsAffected()
	}

	evt := newEvent(r.Options, r.connID, MethodRowsAffected, "", nil)
	r.ctx = before(r.Hooks, r.ctx, evt)
	defer func() {
		evt.Err = err
		after(r.Hooks, r.ctx, evt)
	}()

	cnt, err = r.Result.RowsAffected()
	return
}

func wrapResult(ctx context.Context, connID string, parent driver.Result, o *Options) driver.Result {
	return &otResult{
		Result:  parent,
		ctx:     ctx,
		connID:  connID,
		Options: o,
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
	*Options
	ctx    context.Context
	connID string
}

func (r otRows) Columns() []string {
	return r.Rows.Columns()
}

func (r otRows) Close() (err error) {
	if !r.RowsCloseB {
		return r.Rows.Close()
	}

	evt := newEvent(r.Options, r.connID, MethodRowsClose, "", nil)
	r.ctx = before(r.Hooks, r.ctx, evt)
	defer func() {
		evt.Err = err
		after(r.Hooks, r.ctx, evt)
	}()

	return r.Rows.Close()
}

func (r otRows) Next(dest []driver.Value) (err error) {
	if !r.RowsNextB {
		return r.Rows.Next(dest)
	}

	evt := newEvent(r.Options, r.connID, MethodRowsNext, "", nil)
	r.ctx = before(r.Hooks, r.ctx, evt)
	defer func() {
		evt.Err = err
		after(r.Hooks, r.ctx, evt)
	}()

	err = r.Rows.Next(dest)
	return
}

func wrapRows(ctx context.Context, connID string, parent driver.Rows, o *Options) driver.Rows {
	ts, isColumnTypeScan := parent.(driver.RowsColumnTypeScanType)
	r := otRows{
		Rows:    parent,
		ctx:     ctx,
		connID:  connID,
		Options: o,
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
	query  string
	connID string
	*Options
}

func (s otStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	evt := newEvent(s.Options, s.connID, MethodExec, s.query, args)
	ctx := before(s.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(s.Hooks, ctx, evt)
	}()

	res, err = s.Stmt.Exec(args) // nolint
	if err != nil {
		return nil, err
	}
	return wrapResult(ctx, s.connID, res, s.Options), nil
}

func (s otStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	evt := newEvent(s.Options, s.connID, MethodExec, s.query, args)
	ctx = before(s.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(s.Hooks, ctx, evt)
	}()

	// we already tested driver when wrap stmt
	res, err = s.Stmt.(driver.StmtExecContext).ExecContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return wrapResult(ctx, s.connID, res, s.Options), nil
}

func (s otStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	evt := newEvent(s.Options, s.connID, MethodQuery, s.query, args)
	ctx := before(s.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(s.Hooks, ctx, evt)
	}()

	rows, err = s.Stmt.Query(args) // nolint
	if err != nil {
		return nil, err
	}
	return wrapRows(ctx, s.connID, rows, s.Options), nil
}

func (s otStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	evt := newEvent(s.Options, s.connID, MethodQuery, s.query, args)
	ctx = before(s.Hooks, ctx, evt)
	defer func() {
		evt.Err = err
		after(s.Hooks, ctx, evt)
	}()

	// we already tested driver when wrap stmt
	rows, err = s.Stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	return wrapRows(ctx, s.connID, rows, s.Options), nil
}

func wrapStmt(connID string, stmt driver.Stmt, query string, o *Options) driver.Stmt {
	_, isExecCtx := stmt.(driver.StmtExecContext)
	_, isQueryCtx := stmt.(driver.StmtQueryContext)
	cc, isColumnConverter := stmt.(driver.ColumnConverter) // nolint
	nvc, isNamedValueChecker := stmt.(driver.NamedValueChecker)

	s := otStmt{
		connID:  connID,
		Stmt:    stmt,
		query:   query,
		Options: o,
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
	ctx    context.Context
	connID string
	*Options
}

func (t otTx) Commit() (err error) {
	evt := newEvent(t.Options, t.connID, MethodCommit, "", nil)
	ctx := before(t.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(t.Hooks, ctx, evt)
	}()

	return t.Tx.Commit()
}

func (t otTx) Rollback() (err error) {
	evt := newEvent(t.Options, t.connID, MethodRollback, "", nil)
	ctx := before(t.Hooks, context.Background(), evt)
	defer func() {
		evt.Err = err
		after(t.Hooks, ctx, evt)
	}()

	return t.Tx.Rollback()
}

func wrapTx(ctx context.Context, connID string, tx driver.Tx, o *Options) driver.Tx {
	return otTx{
		Tx:      tx,
		ctx:     ctx,
		connID:  connID,
		Options: o,
	}
}
