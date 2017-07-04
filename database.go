package reiner

import (
	"database/sql"
	"strings"
)

// connection represents a single database connection.
type connection struct {
	db             *sql.DB
	tx             *sql.Tx
	lastCheck      int
	isHealth       bool
	dataSourceName string
}

// DB represents the main database with the connections,
// a database can have a lot of the connections.
type DB struct {
	slaves         []*connection
	master         *connection
	hasSlave       bool
	lastSlaveIndex int
}

// openDatabase opens a single connection.
func openDatabase(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return db, err
	}
	if err = db.Ping(); err != nil {
		return db, err
	}
	return db, nil
}

// newDatabase creates the new connections if there're multiple masters or the slaves.
// It opens a single main connection if there's only one master and no slaves.
func newDatabase(master string, slaves []string) (*DB, error) {
	d := &DB{}
	// Create the main connection if there's only one master an no slaves.
	if len(slaves) == 0 {
		db, err := openDatabase(master)
		if err != nil {
			return d, err
		}
		d.master = &connection{
			db:             db,
			dataSourceName: master,
		}
		return d, nil
	}
	d.hasSlave = true
	// Connect to the slave databases.
	for _, v := range slaves {
		db, err := openDatabase(v)
		if err != nil {
			return d, err
		}
		d.slaves = append(d.slaves, &connection{
			db:             db,
			dataSourceName: v,
		})
	}
	return d, nil
}

// roundRobin picks the next connection to prevent keep using the same connection.
// It works like a simple load balancer.
func (d *DB) roundRobin(pool []*connection, currentIndex int) (index int) {
	length := len(pool) - 1
	index = currentIndex + 1
	if index > length {
		index = 0
	}
	return
}

// getSlave gets a available slave connection.
func (d *DB) getSlave() (db *sql.DB) {
	index := d.roundRobin(d.slaves, d.lastSlaveIndex)
	db = d.slaves[index].db
	// Set the last index.
	d.lastSlaveIndex = index
	return
}

// getDB gets the database connection based on the query, used for the read/write splitting.
func (d *DB) getDB(query ...string) (db *sql.DB) {
	if len(query) == 0 || !d.hasSlave {
		db = d.master.db
		return
	}
	action := strings.Split(query[0], " ")[0]
	switch action {
	case "SELECT":
		db = d.getSlave()
	default:
		db = d.master.db
	}
	return
}

// Begin begins the transaction of the current database connection.
func (d *DB) Begin() (*sql.Tx, error) {
	return d.master.db.Begin()
}

// Rollback rollbacks the transaction.
func (d *DB) Rollback() error {
	if d.master.tx == nil {
		return ErrUnbegunTransaction
	}
	err := d.master.tx.Rollback()
	if err != nil {
		return err
	}
	d.master.tx = nil
	return nil
}

// Commit commits the transaction.
func (d *DB) Commit() error {
	if d.master.tx == nil {
		return ErrUnbegunTransaction
	}
	err := d.master.tx.Commit()
	if err != nil {
		return err
	}
	d.master.tx = nil
	return nil
}

// Ping pings all the connections, includes the slave connections.
func (d *DB) Ping() error {
	var err error
	err = d.master.db.Ping()
	if err != nil {
		return err
	}
	for _, v := range d.slaves {
		err = v.db.Ping()
		if err != nil {
			return err
		}
	}
	return nil
}

// Disconnect disconnects all the connections, includes the slave connections.
func (d *DB) Disconnect() error {
	var err error
	err = d.master.db.Close()
	if err != nil {
		return err
	}
	for _, v := range d.slaves {
		err = v.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Connect reconnects the database connections, includes the slave connections.
func (d *DB) Connect() error {
	db, err := sql.Open("mysql", d.master.dataSourceName)
	if err != nil {
		return err
	}
	d.master.db = db
	for k, v := range d.slaves {
		db, err := sql.Open("mysql", v.dataSourceName)
		if err != nil {
			return err
		}
		d.slaves[k].db = db
	}
	return nil
}

// Prepare prepares the query.
func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	if d.master.tx != nil {
		return d.master.tx.Prepare(query)
	}
	return d.getDB(query).Prepare(query)
}

// Exec executes the queries and returns the result, not the rows.
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if d.master.tx != nil {
		return d.master.tx.Exec(query, args...)
	}
	return d.getDB(query).Exec(query, args...)
}

// Query executes the SQL queries.
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.master.tx != nil {
		return d.master.tx.Query(query, args...)
	}
	return d.getDB(query).Query(query, args...)
}
