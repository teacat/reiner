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
	err = migration.Drop("Users")
	assert.NoError(err)

	err = migration.
		Column("Username").Varchar(32).Primary().
		Column("Password").Varchar(64).
		Column("Age").Int(2).Create("Users")
	assert.NoError(err)
}

type user struct {
	Username string
	Password string
	Age      int
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	id, err := wrapper.Insert("Users", map[string]interface{}{
		"Username": "admin",
		"Password": "test",
		"Age":      19,
	})
	assert.NoError(err)
	assert.Equal(0, id)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	var users []*user
	err := wrapper.Bind(&users).Get("Users")
	assert.NoError(err)
	assert.Len(users, 1)
	assert.Equal("admin", users[0].Username)
	assert.Equal("test", users[0].Password)
	assert.Equal(19, users[0].Age)
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)
	var user user
	err := wrapper.Bind(&user).GetOne("Users")
	assert.NoError(err)
	assert.Equal("admin", user.Username)
	assert.Equal("test", user.Password)
	assert.Equal(19, user.Age)
}

func TestGetValue(t *testing.T) {
	assert := assert.New(t)
	var count int
	err := wrapper.Bind(&count).GetValue("Users", "COUNT(*)")
	assert.NoError(err)
	assert.Equal(1, count)
}
