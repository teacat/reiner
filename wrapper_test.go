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

func TestInsertMulti(t *testing.T) {
	assert := assert.New(t)
	data := []map[string]interface{}{
		{
			"Username": "admin1",
			"Password": "testtest",
			"Age":      90,
		}, {
			"Username": "admin2",
			"Password": "test",
			"Age":      50,
		}, {
			"Username": "admin3",
			"Password": "test",
			"Age":      30,
		},
	}
	wrapper.InsertMulti("Users", data)
	assert.Equal("INSERT INTO `Users` (`Username`, `Password`, `Age`) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)", wrapper.LastQuery)
}

func TestInsertMultiWithKeys(t *testing.T) {
	assert := assert.New(t)
	data := [][]interface{}{
		{
			"admin4", "test", 10,
		}, {
			"admin5", "test", 60,
		}, {
			"admin6", "test", 40,
		},
	}
	columns := []string{"Username", "Password", "Age"}
	wrapper.InsertMulti("Users", data, columns)
	assert.Equal("INSERT INTO `Users` (`Username`, `Password`, `Age`) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)", wrapper.LastQuery)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Get("Users")
	assert.Equal("SELECT * FROM `Users`", wrapper.LastQuery)
}

func TestGetLimit(t *testing.T) {
	assert := assert.New(t)
	wrapper.Limit(5).Get("Users")
	assert.Equal("SELECT * FROM `Users` LIMIT 5", wrapper.LastQuery)
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

func TestGetWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", "admin").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` = ?", wrapper.LastQuery)
}

func TestGetWhereAnd(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", "admin").Where("Password", "123456").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` = ? AND `Password` = ?", wrapper.LastQuery)
}

func TestGetWhereOr(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", "admin").OrWhere("Password", "123456").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` = ? OR `Password` = ?", wrapper.LastQuery)
}

func TestGetWhereIn(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", []string{"admin", "best", "lovely"}, "IN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` IN (?, ?, ?)", wrapper.LastQuery)
}

func TestGetWhereNotIn(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", []string{"xiaoan", "best", "lovely"}, "NOT IN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` NOT IN (?, ?, ?)", wrapper.LastQuery)
}

func TestGetWhereInAnd(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", []string{"admin", "best", "lovely"}, "IN").Where("Password", []string{"123456", "789012", "345678"}, "IN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` IN (?, ?, ?) AND `Password` IN (?, ?, ?)", wrapper.LastQuery)
}

func TestGetWhereInOr(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", []string{"admin", "best", "lovely"}, "IN").OrWhere("Password", []string{"123456", "789012", "345678"}, "IN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` IN (?, ?, ?) OR `Password` IN (?, ?, ?)", wrapper.LastQuery)
}

func TestGetWhereBetween(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Age", []int{0, 20}, "BETWEEN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Age` BETWEEN ? AND ? ", wrapper.LastQuery)
}

func TestGetWhereBetweenAnd(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Age", []int{0, 20}, "BETWEEN").Where("Username", "admin").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Age` BETWEEN ? AND ? AND `Username` = ?", wrapper.LastQuery)
}

func TestGetWhereBetweenOr(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Age", []int{0, 20}, "BETWEEN").OrWhere("Username", "admin").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Age` BETWEEN ? AND ? OR `Username` = ?", wrapper.LastQuery)
}

func TestGetHaving(t *testing.T) {
	assert := assert.New(t)
	wrapper.Having("Username", "admin").Get("Users")
	assert.Equal("SELECT * FROM `Users` HAVING `Username` = ?", wrapper.LastQuery)
}

func TestGetHavingOr(t *testing.T) {
	assert := assert.New(t)
	wrapper.Having("Username", "admin").OrHaving("Username", "yamiodymel").Get("Users")
	assert.Equal("SELECT * FROM `Users` HAVING `Username` = ? OR HAVING `Username` = ?", wrapper.LastQuery)
}

func TestGetColumnCompare(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("`Username` = `Password`").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` = `Password`", wrapper.LastQuery)
}

func TestGetWhereNull(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", nil, "IS NOT").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` IS NOT NULL", wrapper.LastQuery)
}

func TestGetWhereCustomVars(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("(`Username` = ? OR `Password` = ?)", []string{"admin", "123456"}).Where("Age", []int{19}).Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE (`Username` = ? OR `Password` = ?) AND `Age` = ?", wrapper.LastQuery)
}

func TestReplace(t *testing.T) {
	assert := assert.New(t)

}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	data := map[string]interface{}{
		"Password": "456789",
		"Age":      wrapper.Increment(5),
	}
	wrapper.Where("Username", "admin3").Update("Users", data)
	assert.Equal("UPDATE `Users` SET `Password` = ?, `Age` = `Age` + 5 WHERE `Username` = ?", wrapper.LastQuery)
}

func TestUpdateLimit(t *testing.T) {
	assert := assert.New(t)
	data := map[string]interface{}{
		"Age": wrapper.Increment(5),
	}
	wrapper.Limit(3).Update("Users", data)
	assert.Equal("UPDATE `Users` SET `Age` = `Age` + 5 LIMIT 3", wrapper.LastQuery)
}

func TestPaginate(t *testing.T) {
	assert := assert.New(t)
	page := 1
	wrapper.PageLimit = 2
	wrapper.Paginate("Users", page)
	assert.Equal("SELECT * FROM `Users` LIMIT 1, 2", wrapper.LastQuery)
}

func TestSetQueryOption(t *testing.T) {
	assert := assert.New(t)
	wrapper.SetQueryOption("SQL_NO_CACHE").Get("Users")
	assert.Equal("SELECT SQL_NO_CACHE * FROM `Users`", wrapper.LastQuery)
}

func TestSetQueryOptionMulti(t *testing.T) {
	assert := assert.New(t)
	wrapper.SetQueryOption("LOW_PRIORITY", "IGNORE").Insert("Users", map[string]interface{}{
		"Username": "admin100",
		"Password": "123456",
		"Age":      200,
	})
	assert.Equal("INSERT LOW_PRIORITY IGNORE INTO `Users` (`Username`, `Password`, `Age`) VALUES (?, ?, ?)", wrapper.LastQuery)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	wrapper.Where("Username", "admin100").Delete("Users")
	assert.Equal("DELETE FROM `Users` WHERE `Username` = ?", wrapper.LastQuery)
}

func TestOrderBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.OrderBy("Username", "ASC").OrderBy("Password", "DESC").OrderBy("RAND()").Get("Users")
	assert.Equal("SELECT * FROM `Users` ORDER BY `Username` ASC, `PASSWORD` DESC, RAND()", wrapper.LastQuery)
}

func TestOrderByFields(t *testing.T) {
	assert := assert.New(t)
	wrapper.OrderBy("Username", "DESC", []string{"admin", "admin10", "admin5"}).Get("Users")
	assert.Equal("SELECT * FROM `Users` ORDER BY FIELD (`Username`, ?, ?, ?) DESC", wrapper.LastQuery)
}

func TestGroupBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.GroupBy("Username").Get("Users")
	assert.Equal("SELECT * FROM `Users` GROUP BY `Username`", wrapper.LastQuery)
}

func TestLeftJoin(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		LeftJoin("Posts", "Posts.AuthorUsername = Users.Username").
		Where("Users.Username", "admin").
		Columns("Users.Age", "Posts.Title").
		Get("Users")
	assert.Equal("SELECT `Users`.Age, `Posts`.Title WHERE `Users`.Username = ? FROM `Users` AS `Users` LEFT JOIN `Posts` AS `Posts` ON (`Posts`.AuthorUsername = `Users`.Username)", wrapper.LastQuery)
}

func TestRightJoin(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		RightJoin("Posts", "Posts.AuthorUsername = Users.Username").
		Where("Users.Username", "admin").
		Columns("Users.Age", "Posts.Title").
		Get("Users")
	assert.Equal("SELECT `Users`.Age, `Posts`.Title WHERE `Users`.Username = ? FROM `Users` AS `Users` RIGHT JOIN `Posts` AS `Posts` ON (`Posts`.AuthorUsername = `Users`.Username)", wrapper.LastQuery)
}

func TestJoinWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		LeftJoin("Posts", "Posts.AuthorUsername = Users.Username").
		JoinWhere("Users.Username", "admin").
		Where("Users.Username", "admin").
		Columns("Users.Age", "Posts.Title").
		Get("Users")
	assert.Equal("SELECT `Users`.Age, `Posts`.Title WHERE `Users`.Username = ? FROM `Users` AS `Users` LEFT JOIN `Posts` AS `Posts` ON (`Posts`.AuthorUsername = `Users`.Username AND `Users`.Username = ?)", wrapper.LastQuery)
}

func TestJoinOrWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		LeftJoin("Posts", "Posts.AuthorUsername = Users.Username").
		JoinOrWhere("Users.Username", "admin").
		Where("Users.Username", "admin").
		Columns("Users.Age", "Posts.Title").
		Get("Users")
	assert.Equal("SELECT `Users`.Age, `Posts`.Title WHERE `Users`.Username = ? FROM `Users` AS `Users` LEFT JOIN `Posts` AS `Posts` ON (`Posts`.AuthorUsername = `Users`.Username OR `Users`.Username = ?)", wrapper.LastQuery)
}

func TestJoinOrWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		LeftJoin("Posts", "Posts.AuthorUsername = Users.Username").
		JoinOrWhere("Users.Username", "admin").
		Where("Users.Username", "admin").
		Columns("Users.Age", "Posts.Title").
		Get("Users")
	assert.Equal("SELECT `Users`.Age, `Posts`.Title WHERE `Users`.Username = ? FROM `Users` AS `Users` LEFT JOIN `Posts` AS `Posts` ON (`Posts`.AuthorUsername = `Users`.Username OR `Users`.Username = ?)", wrapper.LastQuery)
}

func TestCopy(t *testing.T) {
	assert := assert.New(t)
	a := wrapper.Where("Username", "admin")
	b := a.Copy().Get("Users")
	a.Get("Posts")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` = ?", b.LastQuery)
	assert.Equal("SELECT * FROM `Posts` WHERE `Username` = ?", a.LastQuery)
}
func TestSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	ids := wrapper.SubQuery().Where("ID", 1, ">").Columns("AuthorUsername").Get("Products")
	wrapper.Where("Username", ids, "IN").Get("Users")
	assert.Equal("SELECT * FROM `Users` WHERE `Username` IN (SELECT `AuthorUsername` FROM `Products` WHERE `ID` > 1)", wrapper.LastQuery)
}

func TestSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	usernameQ := wrapper.SubQuery()
	usernameQ.Where("Username", "admin").Columns("Username").GetOne("Users")
	wrapper.Insert("Posts", map[string]interface{}{
		"Title":          "Wow, Amazing!",
		"AuthorUsername": usernameQ,
	})
	assert.Equal("INSERT INTO `Posts` (`Title`, `AuthorUsername`) VALUES (?, (SELECT `Username` FROM `Users` WHERE `Username` = ?))", wrapper.LastQuery)
}

func TestSubQueryJoin(t *testing.T) {
	assert := assert.New(t)
	usersQ := wrapper.SubQuery("Users").Where("Username", "admin").Get("Users")
	wrapper.
		LeftJoin(usersQ, "Products.AuthorUsername = Users.Username").
		Columns("Users.Username", "Posts.Title").
		Get("Posts")
	assert.Equal("SELECT `Users`.Username, `Posts`.Title FROM `Posts` AS `Posts` LEFT JOIN (SELECT * FROM `Users` WHERE `Username` = ?) AS `Users` ON `Products`.AuthorUsername = `Users`.Username", wrapper.LastQuery)
}

func TestSubQueryExist(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery().Where("Username", "admin").Columns("Username").Get("Users")
	wrapper.Where("", subQuery, "EXISTS").Get("Posts")
	assert.Equal("SELECT * FROM `Posts` WHERE EXISTS (SELECT `Username` FROM `Users` WHERE `Username` = ?)", wrapper.LastQuery)
}

func TestHas(t *testing.T) {
	//assert := assert.New(t)
	//wrapper.Where("Username", "admin").Has("Users")
	//assert.Equal("SELECT * FROM `Posts` WHERE EXISTS (SELECT `Username` FROM `Users` WHERE `Username` = ?)", wrapper.LastQuery)
}

func TestTransaction(t *testing.T) {
	//assert := assert.New(t)
	//wrapper.Begin()
	//assert.Equal("SELECT * FROM `Posts` WHERE EXISTS (SELECT `Username` FROM `Users` WHERE `Username` = ?)", wrapper.LastQuery)
}
