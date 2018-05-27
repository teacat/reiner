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
	_, err := rb.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      64,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?)", rb.Query())
	assert.Equal(1, rb.Count())

	_, err = rb.Table("Products").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"ID":       1,
		"PostID":   1,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Products (Username, ID, PostID) VALUES (?, ?, ?)", rb.Query())
	assert.Equal(1, rb.Count())
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
	_, err := rb.Table("Users").InsertMulti(data)
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?), (?, ?, ?)", rb.Query())
	assert.Equal(2, rb.Count())
}

func TestRealReplace(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Replace(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "davai",
		"Age":      85,
	})
	assert.NoError(err)
	assert.Equal("REPLACE INTO Users (Username, Password, Age) VALUES (?, ?, ?)", rb.Query())
	assert.Equal(2, rb.Count()) // Why 2? Check https://blog.xupeng.me/2013/10/11/mysql-replace-into-trap/
}

func TestRealInsertFunc(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Insert(map[string]interface{}{
		"Username": rb.Now("+1Y"),
		"Password": rb.Func("MD5(?)", "secretpassword+salt"),
		"Age":      rb.Func("1 + 1"),
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (NOW() + INTERVAL 1 YEAR, MD5(?), 1 + 1)", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealOnDuplicateInsert(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").OnDuplicate([]string{"Age", "Password"}).Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      95,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE Age = VALUES(Age), Password = VALUES(Password)", rb.Query())
	assert.Equal(2, rb.Count())
}

func TestRealUpdate(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "123456",
	})
	assert.NoError(err)
	assert.Equal("UPDATE Users SET Username = ?, Password = ? WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealLimitUpdate(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Limit(1).Update(map[string]interface{}{
		"Username": "PiuPiu",
		"Password": "123456",
	})
	assert.NoError(err)
	assert.Equal("UPDATE Users SET Username = ?, Password = ? LIMIT 1", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	_, err := rb.Table("Users").Bind(&u).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", rb.Query())
	assert.Equal(4, rb.Count())
	assert.Len(u, 4)
}

func TestRealLimitGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	_, err := rb.Table("Users").Bind(&u).Limit(2).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 2", rb.Query())
	assert.Equal(2, rb.Count())
	assert.Len(u, 2)
}

func TestRealLimitGetOne(t *testing.T) {
	assert := assert.New(t)
	var u user
	_, err := rb.Table("Users").Bind(&u).GetOne()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 1", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal(u.Username, "Karisu")
}

func TestRealGetColumns(t *testing.T) {
	assert := assert.New(t)
	var u []user
	_, err := rb.Table("Users").Bind(&u).Get("Username", "Age")
	assert.NoError(err)
	assert.Equal("SELECT Username, Age FROM Users", rb.Query())
	assert.Equal(4, rb.Count())
	assert.Len(u, 4)

	var c int
	_, err = rb.Table("Users").Bind(&c).Get("COUNT(*) AS Count")
	assert.NoError(err)
	assert.Equal("SELECT COUNT(*) AS Count FROM Users", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal(4, c)
}

func TestRealGetOne(t *testing.T) {
	assert := assert.New(t)

	var u user
	_, err := rb.Table("Users").Bind(&u).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal(12, u.Age)
	assert.Equal("12345", u.Password)
	assert.Equal("Karisu", u.Username)

	_, err = rb.Table("Users").Bind(&u).Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal(95, u.Age)
	assert.Equal("123456", u.Password)
	assert.Equal("YamiOdymel", u.Username)

	var i struct {
		Sum   int
		Count int
	}
	_, err = rb.Table("Users").Bind(&i).Get("SUM(Age) AS Sum", "COUNT(*) AS Count")
	assert.NoError(err)
	assert.Equal("SELECT SUM(Age) AS Sum, COUNT(*) AS Count FROM Users", rb.Query())
	assert.Equal(196, i.Sum)
	assert.Equal(4, i.Count)

	var m map[string]interface{}
	_, err = rb.Table("Users").Bind(&m).Get("SUM(Age) AS Sum", "COUNT(*) AS Count")
	assert.NoError(err)
	assert.Equal("SELECT SUM(Age) AS Sum, COUNT(*) AS Count FROM Users", rb.Query())
	assert.Equal("196", string(m["Sum"].([]uint8)))
	assert.Equal(4, int(m["Count"].(int64)))
}

func TestRealGetValue(t *testing.T) {
	assert := assert.New(t)
	var u []string
	_, err := rb.Table("Users").Bind(&u).Get("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users", rb.Query())
	assert.Len(u, 4)

	_, err = rb.Table("Users").Bind(&u).Limit(2).Get("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username FROM Users LIMIT 2", rb.Query())
	assert.Len(u, 2)
	assert.Equal(2, rb.Count())

	var c int
	_, err = rb.Table("Users").Bind(&c).Get("COUNT(*)")
	assert.NoError(err)
	assert.Equal("SELECT COUNT(*) FROM Users", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal(4, c)
}

func TestRealPaginate(t *testing.T) {
	assert := assert.New(t)

	_, err := rb.Table("Users").WithTotalCount().Get()
	assert.Equal("SELECT SQL_CALC_FOUND_ROWS * FROM Users", rb.Query())
	assert.Equal(4, rb.Count())
	assert.Equal(4, rb.TotalCount)

	rb.PageLimit = 2
	_, err = rb.Table("Users").Paginate(1)

	assert.NoError(err)
	assert.Equal("SELECT SQL_CALC_FOUND_ROWS * FROM Users LIMIT 0, 2", rb.Query())
	assert.Equal(2, rb.Count())
	assert.Equal(4, rb.TotalCount)
	assert.Equal(2, rb.TotalPage)

	_, err = rb.Table("Users").Paginate(2)
	assert.NoError(err)
	assert.Equal("SELECT SQL_CALC_FOUND_ROWS * FROM Users LIMIT 2, 2", rb.Query())
	assert.Equal(2, rb.Count())
	assert.Equal(4, rb.TotalCount)
	assert.Equal(2, rb.TotalPage)
}

func TestRealRawQuery(t *testing.T) {
	assert := assert.New(t)
	var u []user
	_, err := rb.Bind(&u).RawQuery("SELECT * FROM Users WHERE Age >= ?", 80)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age >= ?", rb.Query())
	assert.Equal(2, rb.Count())
	assert.Len(u, 2)
}

func TestRealRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	var u user
	_, err := rb.Bind(&u).RawQuery("SELECT * FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal("YamiOdymel", u.Username)
}

func TestRealRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	var p string
	_, err := rb.Bind(&p).RawQuery("SELECT Password FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal("123456", p)

	_, err = rb.Bind(&p).RawQuery("SELECT Password FROM Users WHERE Username = ? LIMIT 1", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users WHERE Username = ? LIMIT 1", rb.Query())
	assert.Equal(1, rb.Count())
	assert.Equal("123456", p)

	var ps []string
	_, err = rb.Bind(&ps).RawQuery("SELECT Password FROM Users")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users", rb.Query())
	assert.Equal(4, rb.Count())
	assert.Len(ps, 4)
}

func TestRealWhere(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealWhereHaving(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Having("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users HAVING Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealWhereColumns(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username = Password").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = Password", rb.Query())
	assert.Equal(0, rb.Count())
}

func TestRealWhereOperator(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Age", "<=", 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age <= ?", rb.Query())
	assert.Equal(2, rb.Count())
}

func TestRealWhereBetween(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Age", "BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age BETWEEN ? AND ?", rb.Query())
	assert.Equal(2, rb.Count())

	_, err = rb.Table("Users").Where("Age", "NOT BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age NOT BETWEEN ? AND ?", rb.Query())
	assert.Equal(2, rb.Count())
}

func TestRealWhereIn(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "IN", "YamiOdymel", "Karisu", 27, -1, "d").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IN (?, ?, ?, ?, ?)", rb.Query())
	assert.Equal(2, rb.Count())

	_, err = rb.Table("Users").Where("Username", "NOT IN", 1, 5, 27, -1, "d").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username NOT IN (?, ?, ?, ?, ?)", rb.Query())
	assert.Equal(4, rb.Count())
}

func TestRealOrWhere(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "Dave").OrWhere("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ? OR Username = ?", rb.Query())
	assert.Equal(1, rb.Count())

	_, err = rb.Table("Users").Where("Username = ?", 123456).OrWhere("(Username = ? OR Username = ?)", "xxx", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ? OR (Username = ? OR Username = ?)", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealWhereNull(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "IS", nil).Get()
	assert.NoError(err)
	//panic(rb.Params())
	assert.Equal("SELECT * FROM Users WHERE Username IS NULL", rb.Query())
	assert.Equal(0, rb.Count())

	_, err = rb.Table("Users").Where("Username", "IS NOT", nil).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IS NOT NULL", rb.Query())
	assert.Equal(4, rb.Count())
}

func TestRealTimestampDate(t *testing.T) {
	//assert := assert.New(t)
	//ts := rb.Timestamp
	//err := rb.Table("Users").Where("Age", ts.IsDate("2017-07-13")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsYear(2017)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMonth(1)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMonth("January")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsDay(16)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday("Friday")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
}

func TestRealTimestampTime(t *testing.T) {
	//assert := assert.New(t)
	//ts := rb.Timestamp
	//err := rb.Table("Users").Where("CreatedAt", ts.IsHour(18)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsMinute(25)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsSecond(16)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
	//
	//err = rb.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rb.Query())
}

func TestRealRawWhere(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username != Password").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username != Password", rb.Query())
	assert.Equal(4, rb.Count())
}

func TestRealRawWhereParams(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("(Age >= ? OR Age <= ?)", 0, 90).Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE (Age >= ? OR Age <= ?) AND Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealDelete(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").Where("Username", "Shirone").Delete()
	assert.NoError(err)
	assert.Equal("DELETE FROM Users WHERE Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealOrderBy(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").OrderBy("Age", "DESC").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users ORDER BY Age DESC", rb.Query())
	assert.Equal(3, rb.Count())

	_, err = rb.Table("Users").OrderBy("Age", "DESC").Limit(0, 12).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users ORDER BY Age DESC LIMIT 0, 12", rb.Query())
	assert.Equal(3, rb.Count())
}

func TestRealOrderByField(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").OrderBy("Username", "DESC", "YamiOdymel", "Karisu", "Dave").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users ORDER BY FIELD (Username, ?, ?, ?) DESC", rb.Query())
	assert.Equal(3, rb.Count())
}

func TestRealGroupBy(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("Users").GroupBy("Username").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users GROUP BY Username", rb.Query())
	assert.Equal(3, rb.Count())

	_, err = rb.Table("Users").GroupBy("Username", "Age").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users GROUP BY Username, Age", rb.Query())
	assert.Equal(3, rb.Count())

	_, err = rb.Table("Users").GroupBy("Username").Limit(0, 12).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users GROUP BY Username LIMIT 0, 12", rb.Query())
	assert.Equal(3, rb.Count())
}

func TestRealJoin(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username) WHERE Users.Username = ?", rb.Query())
	assert.Equal(1, rb.Count())

	_, err = rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		RightJoin("Products", "Products.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username) RIGHT JOIN Products ON (Products.Username = Users.Username) WHERE Users.Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealJoinWhere(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		JoinWhere("Posts", "Posts.ID", 0).
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Users LEFT JOIN Posts ON (Posts.Username = Users.Username AND Posts.ID = ?) WHERE Users.Username = ?", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	subQuery := rb.SubQuery()
	subQuery.Table("Products").Get("Username")
	_, err := rb.Table("Users").Where("Username", "IN", subQuery).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IN (SELECT Username FROM Products)", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	subQuery := rb.SubQuery()
	subQuery.Table("Users").Where("Username", "YamiOdymel").Get("Username")
	_, err := rb.Table("Posts").Insert(map[string]interface{}{
		"ID":       1,
		"Title":    "測試商品",
		"Username": subQuery,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Posts (ID, Title, Username) VALUES (?, ?, (SELECT Username FROM Users WHERE Username = ?))", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealSubQueryJoin(t *testing.T) {
	assert := assert.New(t)
	subQuery := rb.SubQuery("Users")
	subQuery.Table("Users").Where("Username", "YamiOdymel").Get()
	_, err := rb.
		Table("Products").
		LeftJoin(subQuery, "Products.Username = Users.Username").
		Get("Users.Username", "Products.PostID")
	assert.NoError(err)
	assert.Equal("SELECT Users.Username, Products.PostID FROM Products LEFT JOIN (SELECT * FROM Users WHERE Username = ?) AS Users ON (Products.Username = Users.Username)", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealSubQueryExist(t *testing.T) {
	assert := assert.New(t)
	subQuery := rb.SubQuery("Users")
	subQuery.Table("Users").Where("Username", "YamiOdymel").Get("Username")
	_, err := rb.Table("Products").Where(subQuery, "EXISTS").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Products WHERE EXISTS (SELECT Username FROM Users WHERE Username = ?)", rb.Query())
	assert.Equal(1, rb.Count())
}

func TestRealHas(t *testing.T) {
	assert := assert.New(t)
	has, err := rb.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ? AND Password = ? LIMIT 1", rb.Query())
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

	_, err = tx.Table("Users").Insert(map[string]interface{}{
		"Username": "Petrarca",
		"Password": "yamiodymel",
		"Age":      123456,
	})
	assert.NoError(err)

	_, err = tx2.Table("Users").Insert(map[string]interface{}{
		"Username": "Kadeon",
		"Password": "MoonMoon",
		"Age":      123456,
	})
	assert.NoError(err)

	err = tx2.Commit()
	assert.NoError(err)

	_, err = rb.Table("Users").Insert(map[string]interface{}{
		"Username": "NotInTransaction",
		"Password": "HelloWorld",
		"Age":      123456,
	})
	assert.NoError(err)

	err = tx.Rollback()
	assert.NoError(err)

	_, err = rb.Table("Users").Where("Username", "Petrarca").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(0, rb.Count())

	_, err = rb.Table("Users").Where("Username", "NotInTransaction").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(1, rb.Count())

	_, err = rb.Table("Users").Where("Username", "Kadeon").Limit(1).Get()
	assert.NoError(err)
	assert.Equal(1, rb.Count())
}

func TestRealNull(t *testing.T) {
	assert := assert.New(t)
	_, err := rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID": 123123,
	})
	assert.NoError(err)

	var n struct {
		ID       int
		Username string
	}
	_, err = rb.Table("NullAllowed").Bind(&n).Get()
	assert.Error(err)
	assert.Equal("SELECT * FROM NullAllowed", rb.Query())

	var np struct {
		ID       int
		Username *string
	}
	_, err = rb.Table("NullAllowed").Bind(&np).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM NullAllowed", rb.Query())
	//assert.Nil(n.Username)
	assert.Equal("", n.Username)

	_, err = rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID":       456456,
		"Username": "Wow",
	})
	assert.NoError(err)

	_, err = rb.Table("NullAllowed").Bind(&np).Where("Username", "Wow").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM NullAllowed WHERE Username = ?", rb.Query())
	assert.Equal("Wow", *np.Username)

	var nn struct {
		ID       int
		Username sql.NullString
	}
	nn.ID = 789789

	_, err = rb.Table("NullAllowed").Insert(map[string]interface{}{
		"ID":       nn.ID,
		"Username": nn.Username,
	})
	assert.NoError(err)
}
