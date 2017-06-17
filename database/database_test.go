package database

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

type core struct {
	readConnections  []*connection
	writeConnections []*connection
	checkInterval    int
	hasSlaves        bool
	hasMasters       bool
	lastReadIndex    int
	lastWriteIndex   int
}

func (c *core) roundRobin(pool []*connection, currentIndex int) (index int) {
	length := len(pool)
	index = currentIndex + 1
	if index > length {
		index = 0
	}
	return
}

func (c *core) getReadConnetion() (db *sql.DB) {
	index := c.roundRobin(c.readConnections, c.lastReadIndex)
	db = c.readConnections[index].db
	// Set the last index.
	c.lastReadIndex = index
	return
}

func (c *core) getWriteConnetion() (db *sql.DB) {
	index := c.roundRobin(c.writeConnections, c.lastWriteIndex)
	db = c.writeConnections[index].db
	// Set the last index.
	c.lastWriteIndex = index
	return
}

func (c *core) getDB(query string) (db *sql.DB) {
	isInsert := strings.Split(query, " ")[0] == "INSERT"
	if isInsert {
		db = c.getWriteConnetion()
	} else {
		db = c.getReadConnetion()
	}
	return
}

func (c *core) Exec(query string, args ...interface{}) (sql.Result, error) {
	db := c.getDB(query)
	return db.Exec(query, args)
}

func (c *core) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db := c.getDB(query)
	return db.Query(query, args)
}

func (c *core) QueryRow(query string, args ...interface{}) *sql.Row {
	db := c.getDB(query)
	return db.QueryRow(query, args)
}
