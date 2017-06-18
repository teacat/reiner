package reiner

import (
	"database/sql"
	"strings"
)

type connection struct {
	db             *sql.DB
	lastCheck      int
	isHealth       bool
	dataSourceName string
}

type DB struct {
	readConnections    []*connection
	writeConnections   []*connection
	mainConnection     *connection
	isSingleConnection bool
	checkInterval      int
	lastReadIndex      int
	lastWriteIndex     int
}

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

func (d *DB) roundRobin(pool []*connection, currentIndex int) (index int) {
	length := len(pool) - 1
	index = currentIndex + 1
	if index > length {
		index = 0
	}
	return
}

func (d *DB) getReadConnetion() (db *sql.DB) {
	index := d.roundRobin(d.readConnections, d.lastReadIndex)
	db = d.readConnections[index].db
	// Set the last index.
	d.lastReadIndex = index
	return
}

func (d *DB) getWriteConnetion() (db *sql.DB) {
	index := d.roundRobin(d.writeConnections, d.lastWriteIndex)
	db = d.writeConnections[index].db
	// Set the last index.
	d.lastWriteIndex = index
	return
}

func (d *DB) getDB(query string) (db *sql.DB) {
	firstAction := strings.Split(query, " ")[0]
	isWrite := firstAction == "INSERT" || firstAction == "CREATE"
	if isWrite {
		db = d.getWriteConnetion()
	} else {
		db = d.getReadConnetion()
	}
	return
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	var db *sql.DB
	if d.isSingleConnection {
		db = d.mainConnection.db
	} else {
		db = d.getDB(query)
	}
	return db.Exec(query, args...)
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var db *sql.DB
	if d.isSingleConnection {
		db = d.mainConnection.db
	} else {
		db = d.getDB(query)
	}
	return db.Query(query, args...)
}

func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	var db *sql.DB
	if d.isSingleConnection {
		db = d.mainConnection.db
	} else {
		db = d.getDB(query)
	}
	return db.QueryRow(query, args...)
}
