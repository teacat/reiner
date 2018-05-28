package reiner

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rb *Builder

func TestRealRealMain(t *testing.T) {
	var err error
	assert := assert.New(t)

	rb, err = New("root:root@/test?charset=utf8")
	assert.NoError(err)

	migration = rb.Migration()

	err = migration.Drop("Users", "Posts", "Products", "NullAllowed")
	assert.NoError(err)

	err = migration.Table("Users").
		Column("Username").Varchar(32).Primary().
		Column("Password").Varchar(32).
		Column("Age").Int(2).
		Charset("utf8").
		Create()
	assert.NoError(err)

	err = migration.Table("Posts").
		Column("ID").Int(32).Primary().
		Column("Username").Varchar(32).
		Column("Title").Varchar(32).
		Charset("utf8").
		Create()
	assert.NoError(err)

	err = migration.Table("Products").
		Column("ID").Int(32).Primary().
		Column("Username").Varchar(32).
		Column("PostID").Int(32).
		Charset("utf8").
		Create()
	assert.NoError(err)

	err = migration.Table("NullAllowed").
		Column("ID").Int(32).Primary().
		Column("Username").Varchar(32).Nullable().
		Charset("utf8").
		Create()
	assert.NoError(err)
}

type user struct {
	Username string
	Password string
	Age      int
}

type post struct {
	ID       int
	Username string
	Title    string
}

type product struct {
	ID       int
	Username string
	PostID   int
}

func TestRealInsert(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      64,
	})
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?)", b.Query())
	assert.Equal(1, b.Count())

	b, err = rb.Table("Products").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"ID":       1,
		"PostID":   1,
	})
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Products (Username, ID, PostID) VALUES (?, ?, ?)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealInsertMulti(t *testing.T) {
	assert := assert.New(t)
	data := []map[string]interface{}{
		{
			"Username": "Shirone",
			"Password": "test",
			"Age":      87,
		}, {
			"Username": "Karisu",
			"Password": "12345",
			"Age":      12,
		},
	}
	b, err := rb.Table("Users").InsertMulti(data)
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?), (?, ?, ?)", b.Query())
	assert.Equal(2, b.Count())
}

func TestRealReplace(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Replace(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "davai",
		"Age":      85,
	})
	assert.NoError(err)
	assertEqual(assert, "REPLACE INTO Users (Username, Password, Age) VALUES (?, ?, ?)", b.Query())
	assert.Equal(2, b.Count()) // Why 2? Check https://blog.xupeng.me/2013/10/11/mysql-replace-into-trap/
}

func TestRealInsertFunc(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Insert(map[string]interface{}{
		"Username": rb.Now("+1Y"),
		"Password": rb.Func("MD5(?)", "secretpassword+salt"),
		"Age":      rb.Func("1 + 1"),
	})
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Users (Username, Password, Age) VALUES (NOW() + INTERVAL 1 YEAR, MD5(?), 1 + 1)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealOnDuplicateInsert(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").OnDuplicate([]string{"Age", "Password"}).Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      95,
	})
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE Age = VALUES(Age), Password = VALUES(Password)", b.Query())
	assert.Equal(2, b.Count())
}

func TestRealUpdate(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "123456",
	})
	assert.NoError(err)
	assertEqual(assert, "UPDATE Users SET Username = ?, Password = ? WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealLimitUpdate(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Limit(1).Update(map[string]interface{}{
		"Username": "PiuPiu",
		"Password": "123456",
	})
	assert.NoError(err)
	assertEqual(assert, "UPDATE Users SET Username = ?, Password = ? LIMIT 1", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	b, err := rb.Table("Users").Bind(&u).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users", b.Query())
	assert.Equal(4, b.Count())
	assert.Len(u, 4)
}

func TestRealLimitGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	b, err := rb.Table("Users").Bind(&u).Limit(2).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users LIMIT 2", b.Query())
	assert.Equal(2, b.Count())
	assert.Len(u, 2)
}

func TestRealLimitGetOne(t *testing.T) {
	assert := assert.New(t)
	var u user
	b, err := rb.Table("Users").Bind(&u).GetOne()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users LIMIT 1", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal(u.Username, "Karisu")
}

func TestRealGetColumns(t *testing.T) {
	assert := assert.New(t)
	var u []user
	b, err := rb.Table("Users").Bind(&u).Get("Username", "Age")
	assert.NoError(err)
	assertEqual(assert, "SELECT Username, Age FROM Users", b.Query())
	assert.Equal(4, b.Count())
	assert.Len(u, 4)

	var c int
	b, err = rb.Table("Users").Bind(&c).Get("COUNT(*) AS Count")
	assert.NoError(err)
	assertEqual(assert, "SELECT COUNT(*) AS Count FROM Users", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal(4, c)
}

func TestRealGetOne(t *testing.T) {
	assert := assert.New(t)

	var u user
	b, err := rb.Table("Users").Bind(&u).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal(12, u.Age)
	assert.Equal("12345", u.Password)
	assert.Equal("Karisu", u.Username)

	b, err = rb.Table("Users").Bind(&u).Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal(95, u.Age)
	assert.Equal("123456", u.Password)
	assert.Equal("YamiOdymel", u.Username)

	var i struct {
		Sum   int
		Count int
	}
	b, err = rb.Table("Users").Bind(&i).Get("SUM(Age) AS Sum", "COUNT(*) AS Count")
	assert.NoError(err)
	assertEqual(assert, "SELECT SUM(Age) AS Sum, COUNT(*) AS Count FROM Users", b.Query())
	assert.Equal(196, i.Sum)
	assert.Equal(4, i.Count)

	var m map[string]interface{}
	b, err = rb.Table("Users").Bind(&m).Get("SUM(Age) AS Sum", "COUNT(*) AS Count")
	assert.NoError(err)
	assertEqual(assert, "SELECT SUM(Age) AS Sum, COUNT(*) AS Count FROM Users", b.Query())
	assert.Equal("196", string(m["Sum"].([]uint8)))
	assert.Equal(4, int(m["Count"].(int64)))
}

func TestRealGetValue(t *testing.T) {
	assert := assert.New(t)
	var u []string
	b, err := rb.Table("Users").Bind(&u).Get("Username")
	assert.NoError(err)
	assertEqual(assert, "SELECT Username FROM Users", b.Query())
	assert.Len(u, 4)

	b, err = rb.Table("Users").Bind(&u).Limit(2).Get("Username")
	assert.NoError(err)
	assertEqual(assert, "SELECT Username FROM Users LIMIT 2", b.Query())
	assert.Len(u, 2)
	assert.Equal(2, b.Count())

	var c int
	b, err = rb.Table("Users").Bind(&c).Get("COUNT(*)")
	assert.NoError(err)
	assertEqual(assert, "SELECT COUNT(*) FROM Users", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal(4, c)
}

func TestRealPaginate(t *testing.T) {
	assert := assert.New(t)

	b, err := rb.Table("Users").WithTotalCount().Get()
	assertEqual(assert, "SELECT SQL_CALC_FOUND_ROWS * FROM Users", b.Query())
	assert.Equal(4, b.Count())
	assert.Equal(4, b.TotalCount)

	rb.PageLimit = 2
	b, err = rb.Table("Users").Paginate(1)

	assert.NoError(err)
	assertEqual(assert, "SELECT SQL_CALC_FOUND_ROWS * FROM Users LIMIT 0, 2", b.Query())
	assert.Equal(2, b.Count())
	assert.Equal(4, b.TotalCount)
	assert.Equal(2, b.TotalPage)

	b, err = rb.Table("Users").Paginate(2)
	assert.NoError(err)
	assertEqual(assert, "SELECT SQL_CALC_FOUND_ROWS * FROM Users LIMIT 2, 2", b.Query())
	assert.Equal(2, b.Count())
	assert.Equal(4, b.TotalCount)
	assert.Equal(2, b.TotalPage)
}

func TestRealRawQuery(t *testing.T) {
	assert := assert.New(t)
	var u []user
	b, err := rb.Bind(&u).RawQuery("SELECT * FROM Users WHERE Age >= ?", 80)
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Age >= ?", b.Query())
	assert.Equal(2, b.Count())
	assert.Len(u, 2)
}

func TestRealRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	var u user
	b, err := rb.Bind(&u).RawQuery("SELECT * FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal("YamiOdymel", u.Username)
}

func TestRealRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	var p string
	b, err := rb.Bind(&p).RawQuery("SELECT Password FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assertEqual(assert, "SELECT Password FROM Users WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal("123456", p)

	b, err = rb.Bind(&p).RawQuery("SELECT Password FROM Users WHERE Username = ? LIMIT 1", "YamiOdymel")
	assert.NoError(err)
	assertEqual(assert, "SELECT Password FROM Users WHERE Username = ? LIMIT 1", b.Query())
	assert.Equal(1, b.Count())
	assert.Equal("123456", p)

	var ps []string
	b, err = rb.Bind(&ps).RawQuery("SELECT Password FROM Users")
	assert.NoError(err)
	assertEqual(assert, "SELECT Password FROM Users", b.Query())
	assert.Equal(4, b.Count())
	assert.Len(ps, 4)
}

func TestRealWhere(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealWhereHaving(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Having("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users HAVING Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealWhereColumns(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username = Password").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = Password", b.Query())
	assert.Equal(0, b.Count())
}

func TestRealWhereOperator(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Age", "<=", 80).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Age <= ?", b.Query())
	assert.Equal(2, b.Count())
}

func TestRealWhereBetween(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Age", "BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Age BETWEEN ? AND ?", b.Query())
	assert.Equal(2, b.Count())

	b, err = rb.Table("Users").Where("Age", "NOT BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Age NOT BETWEEN ? AND ?", b.Query())
	assert.Equal(2, b.Count())
}

func TestRealWhereIn(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "IN", "YamiOdymel", "Karisu", 27, -1, "d").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username IN (?, ?, ?, ?, ?)", b.Query())
	assert.Equal(2, b.Count())

	b, err = rb.Table("Users").Where("Username", "NOT IN", 1, 5, 27, -1, "d").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username NOT IN (?, ?, ?, ?, ?)", b.Query())
	assert.Equal(4, b.Count())
}

func TestRealOrWhere(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "Dave").OrWhere("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ? OR Username = ?", b.Query())
	assert.Equal(1, b.Count())

	b, err = rb.Table("Users").Where("Username = ?", 123456).OrWhere("(Username = ? OR Username = ?)", "xxx", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ? OR (Username = ? OR Username = ?)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealWhereNull(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "IS", nil).Get()
	assert.NoError(err)
	//panic(rb.Params())
	assertEqual(assert, "SELECT * FROM Users WHERE Username IS NULL", b.Query())
	assert.Equal(0, b.Count())

	b, err = rb.Table("Users").Where("Username", "IS NOT", nil).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username IS NOT NULL", b.Query())
	assert.Equal(4, b.Count())
}

func TestRealTimestampDate(t *testing.T) {
	//assert := assert.New(t)
	//ts := rb.Timestamp
	//err := rb.Table("Users").Where("Age", ts.IsDate("2017-07-13")).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsYear(2017)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMonth(1)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMonth("January")).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsDay(16)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday("Friday")).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
}

func TestRealTimestampTime(t *testing.T) {
	//assert := assert.New(t)
	//ts := rb.Timestamp
	//err := rb.Table("Users").Where("CreatedAt", ts.IsHour(18)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMinute(25)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsSecond(16)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assertEqual(assert, "SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", b.Query())
}

func TestRealRawWhere(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username != Password").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username != Password", b.Query())
	assert.Equal(4, b.Count())
}

func TestRealRawWhereParams(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("(Age >= ? OR Age <= ?)", 0, 90).Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE (Age >= ? OR Age <= ?) AND Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealDelete(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").Where("Username", "Shirone").Delete()
	assert.NoError(err)
	assertEqual(assert, "DELETE FROM Users WHERE Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealOrderBy(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").OrderBy("Age", "DESC").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users ORDER BY Age DESC", b.Query())
	assert.Equal(3, b.Count())

	b, err = rb.Table("Users").OrderBy("Age", "DESC").Limit(0, 12).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users ORDER BY Age DESC LIMIT 0, 12", b.Query())
	assert.Equal(3, b.Count())
}

func TestRealOrderByField(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").OrderBy("Username", "DESC", "YamiOdymel", "Karisu", "Dave").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users ORDER BY FIELD (Username, ?, ?, ?) DESC", b.Query())
	assert.Equal(3, b.Count())
}

func TestRealGroupBy(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("Users").GroupBy("Username").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users GROUP BY Username", b.Query())
	assert.Equal(3, b.Count())

	b, err = rb.Table("Users").GroupBy("Username", "Age").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users GROUP BY Username, Age", b.Query())
	assert.Equal(3, b.Count())

	b, err = rb.Table("Users").GroupBy("Username").Limit(0, 12).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users GROUP BY Username LIMIT 0, 12", b.Query())
	assert.Equal(3, b.Count())
}

func TestRealJoin(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assertEqual(assert, "SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username) WHERE Users.Username = ?", b.Query())
	assert.Equal(1, b.Count())

	b, err = rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		RightJoin("Products", "Products.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assertEqual(assert, "SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username) RIGHT JOIN Products ON (Products.Username = Users.Username) WHERE Users.Username = ?", b.Query())
	assert.Equal(1, b.Count())

}

func TestRealJoinWhere(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		JoinWhere("Posts", "Posts.ID", 0).
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assertEqual(assert, "SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username AND Posts.ID = ?) WHERE Users.Username = ?", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	subQuery, _ := rb.SubQuery().Table("Products").Get("Username")
	b, err := rb.Table("Users").Where("Username", "IN", subQuery).Get()
	assert.NoError(err, b.Query())
	assertEqual(assert, "SELECT * FROM Users WHERE Username IN (SELECT Username FROM Products)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	subQuery, _ := rb.SubQuery().Table("Users").Where("Username", "YamiOdymel").Get("Username")
	b, err := rb.Table("Posts").Insert(map[string]interface{}{
		"ID":       1,
		"Title":    "測試商品",
		"Username": subQuery,
	})
	assert.NoError(err)
	assertEqual(assert, "INSERT INTO Posts (ID, Title, Username) VALUES (?, ?, (SELECT Username FROM Users WHERE Username = ?))", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealSubQueryJoin(t *testing.T) {
	assert := assert.New(t)
	subQuery, _ := rb.SubQuery("Users").Table("Users").Where("Username", "YamiOdymel").Get()
	b, err := rb.
		Table("Products").
		LeftJoin(subQuery, "Products.Username = Users.Username").
		Get("Users.Username", "Products.PostID")
	assert.NoError(err)
	assertEqual(assert, "SELECT Users.Username, Products.PostID FROM Products LEFT JOIN (SELECT * FROM Users WHERE Username = ?) AS Users ON (Products.Username = Users.Username)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealSubQueryExist(t *testing.T) {
	assert := assert.New(t)
	subQuery, _ := rb.SubQuery("Users").Table("Users").Where("Username", "YamiOdymel").Get("Username")
	b, err := rb.Table("Products").Where(subQuery, "EXISTS").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Products WHERE EXISTS (SELECT Username FROM Users WHERE Username = ?)", b.Query())
	assert.Equal(1, b.Count())
}

func TestRealHas(t *testing.T) {
	assert := assert.New(t)
	b, has, err := rb.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM Users WHERE Username = ? AND Password = ? LIMIT 1", b.Query())
	assert.True(has)
}

func TestRealGoroutine(t *testing.T) {
	var err error
	assert := assert.New(t)
	done := make(chan bool)
	for i := 0; i < 30; i++ {
		go func(i int) {
			_, errAssert := rb.Table("Users").Insert(map[string]interface{}{
				"Username": i,
				"Password": 12345,
				"Age":      12345,
			})
			if errAssert != nil {
				err = errAssert
			}
			done <- true
		}(i)
	}
	<-done
	assert.NoError(err)
}

func TestRealTx(t *testing.T) {
	assert := assert.New(t)

	//
	tx, err := rb.Begin()
	assert.NoError(err)
	assert.Nil(rb.db.master.tx)

	tx2, err := rb.Begin()
	assert.NoError(err)
	assert.Nil(rb.db.master.tx)

	b, err := tx.Table("Users").Insert(map[string]interface{}{
		"Username": "Petrarca",
		"Password": "yamiodymel",
		"Age":      123456,
	})
	assert.NoError(err)

	b, err = tx2.Table("Users").Insert(map[string]interface{}{
		"Username": "Kadeon",
		"Password": "MoonMoon",
		"Age":      123456,
	})
	assert.NoError(err)

	err = tx2.Commit()
	assert.NoError(err)

	b, err = rb.Table("Users").Insert(map[string]interface{}{
		"Username": "NotInTransaction",
		"Password": "HelloWorld",
		"Age":      123456,
	})
	assert.NoError(err)

	err = tx.Rollback()
	assert.NoError(err)

	b, err = rb.Table("Users").Where("Username", "Petrarca").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(0, b.Count())

	b, err = rb.Table("Users").Where("Username", "NotInTransaction").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(1, b.Count())

	b, err = rb.Table("Users").Where("Username", "Kadeon").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(1, b.Count())
}

func TestRealNull(t *testing.T) {
	assert := assert.New(t)
	b, err := rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID": 123123,
	})
	assert.NoError(err)

	var n struct {
		ID       int
		Username string
	}
	b, err = rb.Table("NullAllowed").Bind(&n).Get()
	assert.Error(err)
	assertEqual(assert, "SELECT * FROM NullAllowed", b.Query())

	var np struct {
		ID       int
		Username *string
	}
	b, err = rb.Table("NullAllowed").Bind(&np).Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM NullAllowed", b.Query())
	//assert.Nil(n.Username)
	assert.Equal("", n.Username)

	b, err = rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID":       456456,
		"Username": "Wow",
	})
	assert.NoError(err)

	b, err = rb.Table("NullAllowed").Bind(&np).Where("Username", "Wow").Get()
	assert.NoError(err)
	assertEqual(assert, "SELECT * FROM NullAllowed WHERE Username = ?", b.Query())
	assert.Equal("Wow", *np.Username)

	var nn struct {
		ID       int
		Username sql.NullString
	}
	nn.ID = 789789

	b, err = rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID":       nn.ID,
		"Username": nn.Username,
	})
	assert.NoError(err)
}
