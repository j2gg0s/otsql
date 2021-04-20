package otsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strconv"
	"sync"
)

var regMu sync.Mutex

var (
	// Compile time assertions
	_ driver.Driver           = &otDriver{}
	_ driver.Result           = &otResult{}
	_ driver.Stmt             = &otStmt{}
	_ driver.StmtExecContext  = &otStmt{}
	_ driver.StmtQueryContext = &otStmt{}
	_ driver.Rows             = &otRows{}
)

// Register initializes and registers our otsql wrapped database driver
// identified by its driverName and using provided TraceOptions. On success it
// returns the generated driverName to use when calling sql.Open.
// It is possible to register multiple wrappers for the same database driver if
// needing different TraceOptions for different connections.
func Register(driverName string, options ...TraceOption) (string, error) {
	db, err := sql.Open(driverName, "")
	if err != nil {
		return "", err
	}

	dri := db.Driver()
	if err = db.Close(); err != nil {
		return "", err
	}

	regMu.Lock()
	defer regMu.Unlock()

	driverName = driverName + "-j2gg0s-otsql-"
NewDriverName:
	for i := 0; i < 100; i++ {
		regName := driverName + strconv.Itoa(i)
		for _, name := range sql.Drivers() {
			if name == regName {
				continue NewDriverName
			}
		}

		sql.Register(regName, Wrap(dri, options...))
		return regName, nil
	}
	return "", errors.New("unable to register driver, all slots have been taken")
}

// Wrap takes a SQL driver and wraps it with OpenCensus instrumentation.
func Wrap(dri driver.Driver, options ...TraceOption) driver.Driver {
	return wrapDriver(dri, newTraceOptions(options...))
}

type otConnector struct {
	dc      driver.Connector
	dri     driver.Driver
	options TraceOptions
}

func (oc otConnector) Connect(ctx context.Context) (conn driver.Conn, err error) {
	ctx, _, endTrace := startTrace(ctx, oc.options, methodCreateConn, "", nil)
	defer func() {
		endTrace(ctx, err)
	}()

	conn, err = oc.dc.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return wrapConn(conn, oc.options), nil
}

func (oc otConnector) Driver() driver.Driver {
	return oc.dri
}

// WrapConnector allows wrapping a database driver.Connector which eliminates
// the need to register otsql as an available driver.Driver.
func WrapConnector(dc driver.Connector, options ...TraceOption) driver.Connector {
	return &otConnector{
		dc:      dc,
		dri:     wrapDriver(dc.Driver(), newTraceOptions(options...)),
		options: newTraceOptions(options...),
	}
}

// WrapConn allows an existing driver.Conn to be wrapped by otsql.
func WrapConn(c driver.Conn, options ...TraceOption) driver.Conn {
	return wrapConn(c, newTraceOptions(options...))
}

func wrapDriver(dri driver.Driver, options TraceOptions) driver.Driver {
	if _, ok := dri.(driver.DriverContext); ok {
		return otDriver{Driver: dri, options: options}
	}
	return struct{ driver.Driver }{otDriver{Driver: dri, options: options}}
}

type otDriver struct {
	driver.Driver
	options TraceOptions
}

func (d otDriver) Open(name string) (conn driver.Conn, err error) {
	ctx, span, endTrace := startTrace(context.Background(), d.options, methodCreateConn, "", nil)
	span.SetAttributes(attributeMissingContext)
	defer func() {
		endTrace(ctx, err)
	}()

	conn, err = d.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	return wrapConn(conn, d.options), nil
}

func (d otDriver) OpenConnector(name string) (driver.Connector, error) {
	connector, err := d.Driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &otConnector{
		dc:      connector,
		dri:     d,
		options: d.options,
	}, nil
}
