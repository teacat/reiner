package reiner

import "testing"
import "github.com/stretchr/testify/assert"

var realWrapper *Wrapper

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
		Column("Password").Varchar(32).Create()
	assert.NoError(err)
}

func TestRealInsert(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password) VALUES (?, ?)", realWrapper.LastQuery)
	realWrapper.Table("User").Get
}
