package reiner

import (
	"database/sql"
	"strings"
)

// connection represents a single database connection.
type connection struct {
	db             *sql.DB
	lastCheck      int
	isHealth       bool
	dataSourceName string
}

// DB represents the main database with the connections,
// a database can have a lot of the connections.
type DB struct {
	readConnections    []*connection
	writeConnections   []*connection
	mainConnection     *connection
	isSingleConnection bool
	checkInterval      int
	lastReadIndex      int
	lastWriteIndex     int
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
func newDatabase(masters []string, slaves []string) (*DB, error) {
	d := &DB{}
	// Create the main connection if there's only one master an no slaves.
	if len(masters) == 1 && len(slaves) == 0 {
		db, err := openDatabase(masters[0])
		if err != nil {
			return d, err
		}
		d.isSingleConnection = true
		d.mainConnection = &connection{
			db:             db,
			dataSourceName: masters[0],
		}
		return d, nil
	}

	// Connect to the master databases.
	for _, v := range masters {
		db, err := openDatabase(v)
		if err != nil {
			return d, err
		}
		d.writeConnections = append(d.writeConnections, &connection{
			db:             db,
			dataSourceName: v,
		})
	}
	// Connect to the slave databases.
	for _, v := range slaves {
		db, err := openDatabase(v)
		if err != nil {
			return d, err
		}
		d.readConnections = append(d.readConnections, &connection{
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

// getReadConnetion gets a available read connection.
func (d *DB) getReadConnetion() (db *sql.DB) {
	index := d.roundRobin(d.readConnections, d.lastReadIndex)
	db = d.readConnections[index].db
	// Set the last index.
	d.lastReadIndex = index
	return
}

// getWriteConnection gets a available write connection.
func (d *DB) getWriteConnetion() (db *sql.DB) {
	index := d.roundRobin(d.writeConnections, d.lastWriteIndex)
	db = d.writeConnections[index].db
	// Set the last index.
	d.lastWriteIndex = index
	return
}

// getDB gets the database connection based on the query, used for the read/write splitting.
func (d *DB) getDB(query ...string) (db *sql.DB) {
	if len(query) == 0 || d.isSingleConnection {
		db = d.mainConnection.db
		return
	}

	firstAction := strings.Split(query[0], " ")[0]
	isWrite := firstAction == "INSERT" || firstAction == "CREATE"
	if isWrite {
		db = d.getWriteConnetion()
	} else {
		db = d.getReadConnetion()
	}
	return
}

//
func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	return d.getDB(query).Prepare(query)
}

// Exec executes the queries and returns the result, not the rows.
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.getDB(query).Exec(query, args...)
}

// Query executes the SQL queries.
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.getDB(query).Query(query, args...)
}

// QueryRow executes the query which has only one row as the result.
func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.getDB(query).QueryRow(query, args...)
}
