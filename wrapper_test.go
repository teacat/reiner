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
	Username string `db:"Username"`
	Password string `db:"Password"`
	Age      int    `db:"Age"`
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	wrapper.Insert("Users", map[string]interface{}{
		"Username": "admin",
		"Password": "test",
		"Age":      19,
	})
	assert.Equal("INSERT INTO `Users` (`Username`, `Password`, `Age`) VALUES (?, ?, ?)", wrapper.LastQuery)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Get("Users")
	assert.Equal("SELECT * FROM `Users`", wrapper.LastQuery)
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)
	wrapper.GetOne("Users")
	assert.Equal("SELECT * FROM `Users` LIMIT 1", wrapper.LastQuery)
}

func TestGetValue(t *testing.T) {
	assert := assert.New(t)
	wrapper.GetValue("Users", "COUNT(*)")
	assert.Equal("SELECT COUNT(*) as retval FROM `Users`", wrapper.LastQuery)
}
