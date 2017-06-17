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
	readConnections  []*connection
	writeConnections []*connection
	checkInterval    int
	lastReadIndex    int
	lastWriteIndex   int
}

func newDatabase(masters []string, slaves []string) (*DB, error) {
	d := &DB{}
	// Connect to the master databases.
	for _, v := range masters {
		db, err := sql.Open("mysql", v)
		if err != nil {
			return d, err
		}
		if err = db.Ping(); err != nil {
			return d, err
		}
		d.writeConnections = append(d.writeConnections, &connection{
			db:             db,
			dataSourceName: v,
		})
	}
	// Connect to the slave databases.
	for _, v := range slaves {
		db, err := sql.Open("mysql", v)
		if err != nil {
			return d, err
		}
		if err = db.Ping(); err != nil {
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
	length := len(pool)
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
	isInsert := strings.Split(query, " ")[0] == "INSERT"
	if isInsert {
		db = d.getWriteConnetion()
	} else {
		db = d.getReadConnetion()
	}
	return
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db := d.getDB(query)
	return db.Exec(query, args)
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db := d.getDB(query)
	return db.Query(query, args)
}

func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	db := d.getDB(query)
	return db.QueryRow(query, args)
}
