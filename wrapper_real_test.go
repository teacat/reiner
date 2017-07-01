package reiner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var realWrapper *Wrapper

type User struct {
	Username string
	Password string
	Age      int
}

func TestRealMain(t *testing.T) {
	var err error
	assert := assert.New(t)

	realWrapper, err = New("root:root@/test?charset=utf8")
	assert.NoError(err)

	migration = realWrapper.Migration()

	err = migration.Drop("Users")
	assert.NoError(err)

	err = migration.Table("Users").
		Column("Username").Varchar(32).Primary().
		Column("Password").Varchar(32).
		Column("Age").Int(2).
		Create()
	assert.NoError(err)
}

func TestRealInsert(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      32,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?)", realWrapper.LastQuery)
}

func TestRealInsertMulti(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").InsertMulti([]map[string]interface{}{
		{
			"Username": "Karisu",
			"Password": "ka_ri_su",
			"Age":      17,
		},
		{
			"Username": "Shirone",
			"Password": "shiroi",
			"Age":      18,
		},
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?), (?, ?, ?)", realWrapper.LastQuery)
}

func TestRealGetOne(t *testing.T) {
	assert := assert.New(t)
	var u User
	err := realWrapper.Bind(&u).Table("Users").GetOne()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", realWrapper.LastQuery)

	assert.Equal("YamiOdymel", u.Username)
	assert.Equal("test", u.Password)
	assert.Equal(32, u.Age)
}
