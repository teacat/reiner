package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var db *DB

func TestMain(t *testing.T) {
	assert := assert.New(t)
	var err error
	db, err = New("root:root@/test?charset=utf8")
	assert.NoError(err)
	migration := db.Migration()
	err = migration.
		Column("username").Varchar(32).Primary().
		Column("password").Varchar(64).Create("user")
	assert.NoError(err)
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)

	id, err := db.Insert("user", map[string]interface{}{
		"username": "axdmin",
		"password": 123123123,
		//"age":      19,
	})
	assert.NoError(err)
	assert.Equal(1, id)
}
