package reiner

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
		Column("password").Varchar(64).
		Column("age").Int(2).Create("users")
	assert.NoError(err)
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	id, err := db.Insert("users", map[string]interface{}{
		"username": "admin",
		"password": "test",
		"age":      19,
	})
	assert.NoError(err)
	assert.Equal(1, id)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	var users []struct {
		username, password string
		age                int
	}
	err := db.Bind(&users).Get("users")
	assert.NoError(err)
	assert.Len(users, 1)
	assert.Equal("admin", users[0].username)
	assert.Equal("test", users[0].password)
	assert.Equal(19, users[0].age)
}
