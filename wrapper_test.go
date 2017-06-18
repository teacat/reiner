package reiner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var db *DB
var wrapper *Wrapper

func TestMain(t *testing.T) {
	assert := assert.New(t)
	var err error
	wrapper, err = New("root:root@/test?charset=utf8")
	assert.NoError(err)
	migration := wrapper.Migration()
	err = migration.
		Column("username").Varchar(32).Primary().
		Column("password").Varchar(64).
		Column("age").Int(2).Create("users")
	assert.NoError(err)
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	id, err := wrapper.Insert("users", map[string]interface{}{
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
	err := wrapper.Bind(&users).Get("users")
	assert.NoError(err)
	assert.Len(users, 1)
	assert.Equal("admin", users[0].username)
	assert.Equal("test", users[0].password)
	assert.Equal(19, users[0].age)
}
