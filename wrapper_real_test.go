package reiner

import (
	"database/sql"
	"regexp"
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

func TestRealGet(t *testing.T) {
	assert := assert.New(t)
	var u []User
	err := realWrapper.Bind(&u).Table("Users").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", realWrapper.LastQuery)
	assert.Len(u, 3)
}

func TestRealGetOne(t *testing.T) {
	assert := assert.New(t)
	var u User
	err := realWrapper.Bind(&u).Table("Users").GetOne()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 1", realWrapper.LastQuery)
	assert.Equal("Karisu", u.Username)
	assert.Equal("ka_ri_su", u.Password)
	assert.Equal(17, u.Age)
}

func TestRealGetOneValue(t *testing.T) {
	assert := assert.New(t)
	var u string
	err := realWrapper.Bind(&u).Table("Users").Limit(1).GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users LIMIT 1", realWrapper.LastQuery)
	assert.Equal("Karisu", u)
}

func TestRealGetValue(t *testing.T) {
	assert := assert.New(t)
	var u []string
	err := realWrapper.Bind(&u).Table("Users").GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)
	assert.Equal([]string{"Karisu", "Shirone", "YamiOdymel"}, u)
	assert.Len(u, 3)
}

func TestRealScan(t *testing.T) {
	assert := assert.New(t)
	var result string
	err := realWrapper.Scan(func(r *sql.Rows) {
		//var username string
		//r.Scan(&username)
		//result += username
	}).Table("Users").GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)
	assert.Equal("KarisuShironeYamiOdymel", result)
}

func TestRealLastRows(t *testing.T) {
	assert := assert.New(t)
	var result string
	err := realWrapper.Table("Users").GetValue("Username")
	//r := realWrapper.LastRows
	//for r.Next() {
	//	var username string
	//	r.Scan(&username)
	//	result += username
	//}
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)
	assert.Equal("KarisuShironeYamiOdymel", result)
}

func TestRealFunc(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Insert(map[string]interface{}{
		"Username": "Kadeon",
		"Password": realWrapper.Func("MD5(?)", 12345),
		"Age":      58,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, MD5(?), ?)", realWrapper.LastQuery)
	var p string
	err = realWrapper.Bind(&p).Table("Users").GetValue("Password")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users", realWrapper.LastQuery)
	assert.Equal("827ccb0eea8a706c4c34a16891f84e7b", p)
}

func TestRealNow(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Insert(map[string]interface{}{
		"Username": realWrapper.Now("+1Y"),
		"Password": "Wow, So Spring!",
		"Age":      58,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (NOW() + INTERVAL 1 YEAR, ?, ?)", realWrapper.LastQuery)
	var u string
	err = realWrapper.Bind(&u).Table("Users").GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)

	match, _ := regexp.Match("[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}", []byte(u))
	assert.True(match)
}

func TestRealOnDuplicate(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").OnDuplicate([]string{"Username"}).Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "Wow, So Spring!",
		"Age":      58,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE Username = VALUE(Username)", realWrapper.LastQuery)
	var p string
	err = realWrapper.Bind(&p).Table("Users").Where("Username", "YamiOdymel").GetValue("Password")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users WHERE Username = ?", realWrapper.LastQuery)
	assert.Equal(p, "Wow, So Spring!")
}

func TestRealUpdate(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Age": 72,
	})
	assert.NoError(err)
	assert.Equal("UPDATE Users SET Age = ? WHERE Username = ?", realWrapper.LastQuery)
	var a int
	err = realWrapper.Bind(&a).Table("Users").Where("Username", "YamiOdymel").GetValue("Age")
	assert.NoError(err)
	assert.Equal("SELECT Age FROM Users", realWrapper.LastQuery)
	assert.Equal(a, 72)
}

func TestRealLimit(t *testing.T) {
	assert := assert.New(t)
	err := realWrapper.Table("Users").Limit(2).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 2", realWrapper.LastQuery)
	assert.Equal(2, realWrapper.Count)
}

func TestRealPaginate(t *testing.T) {
	assert := assert.New(t)
	realWrapper.PageLimit = 2
	err := realWrapper.Table("Users").Paginate(1)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 0, 2", realWrapper.LastQuery)
	assert.Equal(2, realWrapper.Count)

	err = realWrapper.Table("Users").Paginate(2)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 2, 2", realWrapper.LastQuery)
	assert.Equal(2, realWrapper.Count)
}

func TestRealRawQuery(t *testing.T) {
	assert := assert.New(t)
	var u []User
	err := realWrapper.Bind(&u).RawQuery("SELECT * FROM Users")
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", realWrapper.LastQuery)
	assert.Len(u, 4)
}

func TestRealRawQueryCondition(t *testing.T) {
	assert := assert.New(t)
	var u []User
	err := realWrapper.Bind(&u).RawQuery("SELECT * FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", realWrapper.LastQuery)
	assert.Len(u, 1)
}

func TestRealRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	var u User
	err := realWrapper.Bind(&u).RawQueryOne("SELECT * FROM Users")
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", realWrapper.LastQuery)
	assert.Equal("", u.Username)
}

func TestRealRawQueryValues(t *testing.T) {
	assert := assert.New(t)
	var u []string
	err := realWrapper.Bind(&u).RawQueryValue("SELECT Username FROM Users")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)
	assert.Len(u, 4)
}

func TestRealRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	var u string
	err := realWrapper.Bind(&u).RawQueryValue("SELECT Username FROM Users LIMIT 1")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", realWrapper.LastQuery)
	assert.Equal("", u)
}

func TestRealWhere(t *testing.T) {
	assert := assert.New(t)
	var u User
	err := realWrapper.Bind(&u).Where("Username", "YamiOdymel").Where("Age", 72).Get()
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users WHERE Username = ? AND Age = ?", realWrapper.LastQuery)
	assert.Equal("", u)
}
