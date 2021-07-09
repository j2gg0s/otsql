package otsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
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
// identified by its driverName and using provided Options. On success it
// returns the generated driverName to use when calling sql.Open.
// It is possible to register multiple wrappers for the same database driver if
// needing different Options for different connections.
func Register(driverName string, options ...Option) (string, error) {
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
	for i := 0; i < 100; i++ {
		exist := false
		regName := driverName + strconv.Itoa(i)
		for _, name := range sql.Drivers() {
			if name == regName {
				exist = true
				break
			}
		}

		if !exist {
			sql.Register(regName, Wrap(dri, options...))
			return regName, nil
		}
	}
	return "", errors.New("unable to register driver, all slots have been taken")
}

// Wrap takes a SQL driver and wraps it with hook enabled.
func Wrap(dri driver.Driver, opts ...Option) driver.Driver {
	return wrapDriver(dri, newOptions(opts))
}

type otConnector struct {
	dc  driver.Connector
	dri driver.Driver
	*Options
}

func (oc otConnector) Connect(ctx context.Context) (conn driver.Conn, err error) {
	evt := newEvent(oc.Options, "", MethodCreateConn, "", nil)
	before(oc.Hooks, ctx, evt)

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	defer func() {
		evt.Err = err
		evt.Conn = id
		after(oc.Hooks, ctx, evt)
	}()

	conn, err = oc.dc.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return wrapConn(id, conn, oc.Options), nil

}

func (oc otConnector) Driver() driver.Driver {
	return oc.dri
}

// WrapConnector allows wrapping a database driver.Connector which eliminates
// the need to register otsql as an available driver.Driver.
func WrapConnector(dc driver.Connector, opts ...Option) driver.Connector {
	o := newOptions(opts)
	return &otConnector{
		dc:      dc,
		dri:     wrapDriver(dc.Driver(), o),
		Options: o,
	}
}

// WrapConn allows an existing driver.Conn to be wrapped by otsql.
func WrapConn(c driver.Conn, opts ...Option) driver.Conn {
	return wrapConn(fmt.Sprintf("%d", time.Now().UnixNano()), c, newOptions(opts))
}

func wrapDriver(dri driver.Driver, o *Options) driver.Driver {
	if _, ok := dri.(driver.DriverContext); ok {
		return otDriver{Driver: dri, Options: o}
	}
	return struct{ driver.Driver }{otDriver{Driver: dri, Options: o}}
}

type otDriver struct {
	driver.Driver

	*Options
}

func (d otDriver) Open(name string) (conn driver.Conn, err error) {
	ctx := context.Background()

	evt := newEvent(d.Options, "", MethodCreateConn, "", nil)
	before(d.Hooks, ctx, evt)

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	defer func() {
		evt.Err = err
		evt.Conn = id
		after(d.Hooks, ctx, evt)
	}()

	conn, err = d.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	return wrapConn(id, conn, addInstance(d.Options, name)), nil
}

func (d otDriver) OpenConnector(name string) (driver.Connector, error) {
	connector, err := d.Driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &otConnector{
		dc:      connector,
		dri:     d,
		Options: addInstance(d.Options, name),
	}, nil
}

func addInstance(o *Options, dsn string) *Options {
	instance, database := parseDSN(dsn)
	if o.Instance == "" && instance != "" {
		o.Instance = instance
	}
	if database != "" {
		o.Database = database
	}
	return o
}
