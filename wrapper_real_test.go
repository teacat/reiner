package reiner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var rw *Wrapper

func TestRealRealMain(t *testing.T) {
	var err error
	assert := assert.New(t)

	rw, err = New("root:root@/test?charset=utf8")
	assert.NoError(err)

	migration = rw.Migration()

	err = migration.Drop("Users")
	assert.NoError(err)

	err = migration.Table("Users").
		Column("Username").Varchar(32).Primary().
		Column("Password").Varchar(32).
		Column("Age").Int(2).
		Create()
	assert.NoError(err)

	err = migration.Table("Posts").
		Column("ID").Int(32).Primary().
		Column("Username").Varchar(32).
		Column("Title").Varchar(32).
		Create()
	assert.NoError(err)

	err = migration.Table("Products").
		Column("ID").Int(32).Primary().
		Column("Username").Varchar(32).
		Column("PostID").Int(32).
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
	err := rw.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      64,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?)", rw.Query)
	assert.Equal(1, rw.Count)
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
	err := rw.Table("Users").InsertMulti(data)
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?), (?, ?, ?)", rw.Query)
	assert.Equal(2, rw.Count)
}

func TestRealReplace(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Replace(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "davai",
	})
	assert.NoError(err)
	assert.Equal("REPLACE INTO Users (Username, Password) VALUES (?, ?)", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealInsertFunc(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Insert(map[string]interface{}{
		"Username": rw.Now("+1Y"),
		"Password": rw.Func("SHA1(?)", "secretpassword+salt"),
		"Age":      rw.Func("1 + 1"),
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (NOW() + INTERVAL 1 YEAR, SHA(?), 1 + 1)", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealOnDuplicateInsert(t *testing.T) {
	assert := assert.New(t)
	lastInsertID := "ID"
	err := rw.Table("Users").OnDuplicate([]string{"Age", "Password"}, lastInsertID).Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
		"Age":      95,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Users (Username, Password, Age) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE Age = VALUE(Age), Password = VALUE(Password)", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealUpdate(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Username": "Davai",
		"Password": "123456",
	})
	assert.NoError(err)
	assert.Equal("UPDATE Users SET Username = ?, Password = ? WHERE Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealLimitUpdate(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Limit(1).Update(map[string]interface{}{
		"Username": "PiuPiu",
		"Password": "123456",
	})
	assert.NoError(err)
	assert.Equal("UPDATE Users SET Username = ?, Password = ? LIMIT 1", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	err := rw.Table("Users").Bind(&u).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users", rw.Query)
	assert.Equal(4, rw.Count)
	assert.Len(u, 4)
}

func TestRealLimitGet(t *testing.T) {
	assert := assert.New(t)
	var u []user
	err := rw.Table("Users").Limit(2).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 2", rw.Query)
	assert.Equal(2, rw.Count)
	assert.Len(u, 2)
}

func TestRealGetColumns(t *testing.T) {
	assert := assert.New(t)
	var u []user
	err := rw.Table("Users").Get("Username", "Age")
	assert.NoError(err)
	assert.Equal("SELECT Username, Age FROM Users", rw.Query)
	assert.Equal(4, rw.Count)
	assert.Len(u, 4)

	var c int
	err = rw.Table("Users").Bind(&c).Get("COUNT(*) AS Count")
	assert.NoError(err)
	assert.Equal("SELECT COUNT(*) AS Count FROM Users", rw.Query)
	assert.Equal(1, rw.Count)
	assert.Equal(4, c)
}

func TestRealGetOne(t *testing.T) {
	assert := assert.New(t)

	var u user
	err := rw.Table("Users").Bind(&u).Where("Username", "YamiOdymel").GetOne()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rw.Query)
	assert.Equal(u.Age, 32)
	assert.Equal(u.Password, 32)
	assert.Equal(u.Username, "YamiOdymel")

	var m map[string]interface{}
	err = rw.Table("Users").Bind(&m).GetOne("SUM(Username) AS Sum", "COUNT(*) AS Count")
	// rw.Table("Users").Bind(&sum, &cnt).GetOne("SUM(Username) AS Sum", "COUNT(*) AS Count")
	assert.NoError(err)
	assert.Equal("SELECT SUM(Username) AS Sum, COUNT(*) AS Count FROM Users", rw.Query)
	assert.Equal(2018, m["Sum"])
	assert.Equal(4, m["Count"])
}

func TestRealGetValue(t *testing.T) {
	assert := assert.New(t)
	var u []string
	err := rw.Table("Users").Bind(&u).GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username AS Value FROM Users", rw.Query)
	assert.Len(u, 4)

	err = rw.Table("Users").Bind(&u).Limit(2).GetValue("Username")
	assert.NoError(err)
	assert.Equal("SELECT Username AS Value FROM Users LIMIT 2", rw.Query)
	assert.Len(u, 2)
	assert.Equal(2, rw.Count)

	var c int
	err = rw.Table("Users").Bind(&c).GetValue("COUNT(*)")
	assert.NoError(err)
	assert.Equal("SELECT COUNT(*) AS Value FROM Users", rw.Query)
	assert.Equal(1, rw.Count)
	assert.Equal(4, c)
}

func TestRealPaginate(t *testing.T) {
	assert := assert.New(t)
	rw.PageLimit = 2
	err := rw.Table("Users").Paginate(1)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 0, 2", rw.Query)
	assert.Equal(2, rw.Count)

	err = rw.Table("Users").Paginate(2)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users LIMIT 2, 2", rw.Query)
	assert.Equal(2, rw.Count)
}

func TestRealRawQuery(t *testing.T) {
	assert := assert.New(t)
	var u []user
	err := rw.Bind(&u).RawQuery("SELECT * FROM Users WHERE Age >= ?", 80)
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username >= ?", rw.Query)
	assert.Equal(3, rw.Count)
	assert.Len(u, 3)
}

func TestRealRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	var u user
	err := rw.RawQueryOne("SELECT * FROM Users WHERE Username = ?", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
	assert.Equal("YamiOdymel", u.Username)
}

func TestRealRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	var p string
	err := rw.Bind(&p).RawQueryValue("SELECT Password FROM Users WHERE Username = ? LIMIT 1", "YamiOdymel")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users WHERE Username = ? LIMIT 1", rw.Query)
	assert.Equal(1, rw.Count)
	assert.Equal("davai", p)

	var ps []string
	err = rw.Bind(&p).RawQueryValue("SELECT Password FROM Users")
	assert.NoError(err)
	assert.Equal("SELECT Password FROM Users LIMIT 1", rw.Query)
	assert.Equal(4, rw.Count)
	assert.Len(ps, 4)
}

func TestRealWhere(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealWhereHaving(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Having("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users HAVING Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealWhereColumns(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username = Password").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = Password", rw.Query)
	assert.Equal(0, rw.Count)
}

func TestRealWhereOperator(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Age", "<=", 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age <= ?", rw.Query)
	assert.Equal(3, rw.Count)
}

func TestRealWhereBetween(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Age", "BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age BETWEEN ? AND ?", rw.Query)
	assert.Equal(3, rw.Count)

	err = rw.Table("Users").Where("Age", "NOT BETWEEN", 0, 80).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Age NOT BETWEEN ? AND ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealWhereIn(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "IN", "YamiOdymel", "Karisu", 27, -1, "d").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IN (?, ?, ?, ?, ?)", rw.Query)
	assert.Equal(2, rw.Count)

	err = rw.Table("Users").Where("Username", "NOT IN", 1, 5, 27, -1, "d").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username NOT IN (?, ?, ?, ?, ?)", rw.Query)
	assert.Equal(5, rw.Count)
}

func TestRealOrWhere(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "Dave").OrWhere("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ? OR Username = ?", rw.Query)
	assert.Equal(1, rw.Count)

	err = rw.Table("Users").Where("Username = ?", 123456).OrWhere("(Username = ? OR Username = ?)", "xxx", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username = ? OR (Username = ? OR Username = ?)", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealWhereNull(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "IS", nil).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IS NULL", rw.Query)
	assert.Equal(0, rw.Count)

	err = rw.Table("Users").Where("Username", "IS NOT", nil).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IS NOT NULL", rw.Query)
	assert.Equal(5, rw.Count)
}

func TestRealTimestampDate(t *testing.T) {
	//assert := assert.New(t)
	//ts := rw.Timestamp
	//err := rw.Table("Users").Where("Age", ts.IsDate("2017-07-13")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsYear(2017)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsMonth(1)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//err = rw.Table("Users").Where("CreatedAt", ts.IsMonth("January")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsDay(16)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//err = rw.Table("Users").Where("CreatedAt", ts.IsWeekday("Friday")).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
}

func TestRealTimestampTime(t *testing.T) {
	//assert := assert.New(t)
	//ts := rw.Timestamp
	//err := rw.Table("Users").Where("CreatedAt", ts.IsHour(18)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsMinute(25)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsSecond(16)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
	//
	//err = rw.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", rw.Query)
}

func TestRealRawWhere(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username != Password").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username != Password", rw.Query)
	assert.Equal(5, rw.Count)
}

func TestRealRawWhereParams(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("(Age >= ? OR Age <= ?)", 0, 90).Where("Username", "YamiOdymel").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE (Age >= ? OR Age <= ?) AND Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealDelete(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").Where("Username", "Shirone").Delete()
	assert.NoError(err)
	assert.Equal("DELETE FROM Users WHERE Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealOrderBy(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").OrderBy("Age", "DESC").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users ORDER BY Age DESC", rw.Query)
	assert.Equal(4, rw.Count)
}

func TestRealOrderByField(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").OrderBy("Username", "DESC", "YamiOdymel", "Karisu", "Dave").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users ORDER BY FIELD (Username, ?, ?, ?) DESC", rw.Query)
	assert.Equal(2, rw.Count)
}

func TestRealGroupBy(t *testing.T) {
	assert := assert.New(t)
	err := rw.Table("Users").GroupBy("Username").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users GROUP BY Username", rw.Query)
	assert.Equal(4, rw.Count)

	err = rw.Table("Users").GroupBy("Username", "Age").Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users GROUP BY Username, Age", rw.Query)
	assert.Equal(4, rw.Count)
}

func TestRealJoin(t *testing.T) {
	assert := assert.New(t)
	err := rw.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Posts LEFT JOIN Users ON (Posts.Username = Users.Username) WHERE Users.Username = ?", rw.Query)
	assert.Equal(1, rw.Count)

	err = rw.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		RightJoin("Products", "Products.Username = Users.Username").
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Posts LEFT JOIN Users ON (Posts.Username = Users.Username) RIGHT JOIN Products ON (Products.Username = Users.Username) WHERE Users.Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealJoinWhere(t *testing.T) {
	assert := assert.New(t)
	err := rw.
		Table("Users").
		LeftJoin("Posts", "Posts.Username = Users.Username").
		JoinWhere("Posts", "Posts.ID", 0).
		Where("Users.Username", "YamiOdymel").
		Get("Users.Age", "Posts.Title")
	assert.NoError(err)
	assert.Equal("SELECT Users.Age, Posts.Title FROM Posts LEFT JOIN Users ON (Posts.Username = Users.Username AND Posts.ID = ?) WHERE Users.Username = ?", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	subQuery := rw.SubQuery()
	subQuery.Table("Products").Get("Username")
	err := rw.Table("Users").Where("Username", "IN", subQuery).Get()
	assert.NoError(err)
	assert.Equal("SELECT * FROM Users WHERE Username IN (SELECT Username FROM Products)", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	subQuery := rw.SubQuery()
	subQuery.Table("Users").Where("Username", "YamiOdymel").GetOne("Username")
	err := rw.Table("Posts").Insert(map[string]interface{}{
		"Title":    "測試商品",
		"Username": subQuery,
	})
	assert.NoError(err)
	assert.Equal("INSERT INTO Posts (Title, Username) VALUES (?, (SELECT Username FROM Users WHERE Username = ?))", rw.Query)
	assert.Equal(1, rw.Count)
}

func TestRealSubQueryJoin(t *testing.T) {
	//assert := assert.New(t)
	//subQuery := rw.SubQuery("Users")
	//subQuery.Table("Users").Where("Active", 1).Get()
	//err := rw.
	//	Table("Products").
	//	LeftJoin(subQuery, "Products.UserID = Users.ID").
	//	Get("Users.Username", "Products.ProductName")
	//assert.NoError(err)
	//assert.Equal("SELECT Users.Username, Products.ProductName FROM Products LEFT JOIN (SELECT * FROM Users WHERE Active = ?) AS Users ON //(Products.UserID = Users.ID)", rw.Query)
}

func TestRealSubQueryExist(t *testing.T) {
	//assert := assert.New(t)
	//subQuery := rw.SubQuery("Users")
	//subQuery.Table("Users").Where("Company", "測試公司").Get("UserID")
	//err := rw.Table("Products").Where(subQuery, "EXISTS").Get()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Products WHERE EXISTS (SELECT UserID FROM Users WHERE Company = ?)", rw.Query)
}

func TestRealHas(t *testing.T) {
	//assert := assert.New(t)
	//has, err := rw.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
	//assert.NoError(err)
	//assert.Equal("SELECT * FROM Users WHERE Username = ? AND Password = ?", rw.Query)
	//assert.True(has)
}
