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
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)

	id, err := db.Insert("test", map[string]interface{}{
		"username": "admin",
		"password": "test",
		"age":      19,
	})
	assert.NoError(err)
	assert.Equal(1, id)
}
