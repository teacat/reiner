package reiner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var wrapper *Wrapper

func TestMain(t *testing.T) {
	assert := assert.New(t)

	var err error
	wrapper, err = New()
	assert.NoError(err)
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
	})
	assert.Equal("INSERT INTO Users (Username, Password) VALUES (?, ?)", wrapper.Query())
}

func TestInsertMulti(t *testing.T) {
	assert := assert.New(t)
	data := []map[string]interface{}{
		{
			"Username": "YamiOdymel",
			"Password": "test",
		}, {
			"Username": "Karisu",
			"Password": "12345",
		},
	}
	wrapper.Table("Users").InsertMulti(data)
	assert.Equal("INSERT INTO Users (Username, Password) VALUES (?, ?), (?, ?)", wrapper.Query())
}

func TestReplace(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Replace(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
	})
	assert.Equal("REPLACE INTO Users (Username, Password) VALUES (?, ?)", wrapper.Query())
}

func TestInsertFunc(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Insert(map[string]interface{}{
		"Username":  "YamiOdymel",
		"Password":  wrapper.Func("SHA1(?)", "secretpassword+salt"),
		"Expires":   wrapper.Now("+1Y"),
		"CreatedAt": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Users (Username, Password, Expires, CreatedAt) VALUES (?, SHA1(?), NOW() + INTERVAL 1 YEAR, NOW())", wrapper.Query())
}

func TestOnDuplicateInsert(t *testing.T) {
	assert := assert.New(t)
	lastInsertID := "ID"
	wrapper.Table("Users").OnDuplicate([]string{"UpdatedAt"}, lastInsertID).Insert(map[string]interface{}{
		"Username":  "YamiOdymel",
		"Password":  "test",
		"UpdatedAt": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Users (Username, Password, UpdatedAt) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID), UpdatedAt = VALUES(UpdatedAt)", wrapper.Query())
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Username": "Karisu",
		"Password": "123456",
	})
	assert.Equal("UPDATE Users SET Username = ?, Password = ? WHERE Username = ?", wrapper.Query())
}

func TestLimitUpdate(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Limit(10).Update(map[string]interface{}{
		"Username": "Karisu",
		"Password": "123456",
	})
	assert.Equal("UPDATE Users SET Username = ?, Password = ? LIMIT 10", wrapper.Query())
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Get()
	assert.Equal("SELECT * FROM Users", wrapper.Query())
}

func TestLimitGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Limit(10).Get()
	assert.Equal("SELECT * FROM Users LIMIT 10", wrapper.Query())
}

func TestGetColumns(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Get("Username", "Nickname")
	assert.Equal("SELECT Username, Nickname FROM Users", wrapper.Query())

	wrapper.Table("Users").Get("COUNT(*) AS Count")
	assert.Equal("SELECT COUNT(*) AS Count FROM Users", wrapper.Query())
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Get()
	assert.Equal("SELECT * FROM Users WHERE ID = ?", wrapper.Query())

	wrapper.Table("Users").GetOne()
	assert.Equal("SELECT * FROM Users LIMIT 1", wrapper.Query())

	wrapper.Table("Users").Get("SUM(ID)", "COUNT(*) AS Count")
	assert.Equal("SELECT SUM(ID), COUNT(*) AS Count FROM Users", wrapper.Query())
}

func TestGetValue(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Get("Username")
	assert.Equal("SELECT Username FROM Users", wrapper.Query())

	wrapper.Table("Users").Limit(5).Get("Username")
	assert.Equal("SELECT Username FROM Users LIMIT 5", wrapper.Query())

	wrapper.Table("Users").Get("COUNT(*)")
	assert.Equal("SELECT COUNT(*) FROM Users", wrapper.Query())
}

func TestPaginate(t *testing.T) {
	assert := assert.New(t)
	wrapper.PageLimit = 20
	wrapper.Table("Users").Paginate(2)
	assert.Equal("SELECT SQL_CALC_FOUND_ROWS * FROM Users LIMIT 20, 20", wrapper.Query())
}

func TestRawQuery(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQuery("SELECT * FROM Users WHERE ID >= ?", 10)
	assert.Equal("SELECT * FROM Users WHERE ID >= ?", wrapper.Query())
}

func TestRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQuery("SELECT * FROM Users WHERE ID = ?", 10)
	assert.Equal("SELECT * FROM Users WHERE ID = ?", wrapper.Query())
}

func TestRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQuery("SELECT Password FROM Users WHERE ID = ? LIMIT 1", 10)
	assert.Equal("SELECT Password FROM Users WHERE ID = ? LIMIT 1", wrapper.Query())
}

func TestWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Where("Username", "admin").Get()
	assert.Equal("SELECT * FROM Users WHERE ID = ? AND Username = ?", wrapper.Query())
}

func TestWhereHaving(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Having("Username", "admin").Get()
	assert.Equal("SELECT * FROM Users WHERE ID = ? HAVING Username = ?", wrapper.Query())
}

func TestWhereColumns(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("LastLogin = CreatedAt").Get()
	assert.Equal("SELECT * FROM Users WHERE LastLogin = CreatedAt", wrapper.Query())
}

func TestWhereOperator(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", ">=", 50).Get()
	assert.Equal("SELECT * FROM Users WHERE ID >= ?", wrapper.Query())
}

func TestWhereBetween(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", "BETWEEN", 0, 20).Get()
	assert.Equal("SELECT * FROM Users WHERE ID BETWEEN ? AND ?", wrapper.Query())

	wrapper.Table("Users").Where("ID", "NOT BETWEEN", 0, 20).Get()
	assert.Equal("SELECT * FROM Users WHERE ID NOT BETWEEN ? AND ?", wrapper.Query())
}

func TestWhereIn(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", "IN", 1, 5, 27, -1, "d").Get()
	assert.Equal("SELECT * FROM Users WHERE ID IN (?, ?, ?, ?, ?)", wrapper.Query())

	wrapper.Table("Users").Where("ID", "NOT IN", 1, 5, 27, -1, "d").Get()
	assert.Equal("SELECT * FROM Users WHERE ID NOT IN (?, ?, ?, ?, ?)", wrapper.Query())
}

func TestOrWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("FirstName", "John").OrWhere("FirstName", "Peter").Get()
	assert.Equal("SELECT * FROM Users WHERE FirstName = ? OR FirstName = ?", wrapper.Query())

	wrapper.Table("Users").Where("A = B").OrWhere("(A = C OR A = D)").Get()
	assert.Equal("SELECT * FROM Users WHERE A = B OR (A = C OR A = D)", wrapper.Query())
}

func TestWhereNull(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("LastName", "IS", nil).Get()
	assert.Equal("SELECT * FROM Users WHERE LastName IS NULL", wrapper.Query())

	wrapper.Table("Users").Where("LastName", "IS NOT", nil).Get()
	assert.Equal("SELECT * FROM Users WHERE LastName IS NOT NULL", wrapper.Query())
}

func TestTimestampDate(t *testing.T) {
	assert := assert.New(t)
	ts := wrapper.Timestamp
	wrapper.Table("Users").Where("CreatedAt", ts.IsDate("2017-07-13")).Get()
	assert.Equal("SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsYear(2017)).Get()
	assert.Equal("SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsMonth(1)).Get()
	assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())
	wrapper.Table("Users").Where("CreatedAt", ts.IsMonth("January")).Get()
	assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsDay(16)).Get()
	assert.Equal("SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday("Friday")).Get()
	assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())
}

func TestTimestampTime(t *testing.T) {
	assert := assert.New(t)
	ts := wrapper.Timestamp
	wrapper.Table("Users").Where("CreatedAt", ts.IsHour(18)).Get()
	assert.Equal("SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsMinute(25)).Get()
	assert.Equal("SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsSecond(16)).Get()
	assert.Equal("SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())

	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.Query())
}

func TestRawWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID != CompanyID").Where("DATE(CreatedAt) = DATE(LastLogin)").Get()
	assert.Equal("SELECT * FROM Users WHERE ID != CompanyID AND DATE(CreatedAt) = DATE(LastLogin)", wrapper.Query())
}

func TestRawWhereParams(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("(ID = ? OR ID = ?)", 6, 2).Where("Login", "Mike").Get()
	assert.Equal("SELECT * FROM Users WHERE (ID = ? OR ID = ?) AND Login = ?", wrapper.Query())
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Delete()
	assert.Equal("DELETE FROM Users WHERE ID = ?", wrapper.Query())
}

func TestOrderBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").OrderBy("ID", "ASC").OrderBy("Login", "DESC").OrderBy("RAND()").Get()
	assert.Equal("SELECT * FROM Users ORDER BY ID ASC, Login DESC, RAND()", wrapper.Query())
}

func TestOrderByField(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").OrderBy("UserGroup", "ASC", "SuperUser", "Admin", "Users").Get()
	assert.Equal("SELECT * FROM Users ORDER BY FIELD (UserGroup, ?, ?, ?) ASC", wrapper.Query())
}

func TestGroupBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").GroupBy("Name").Get()
	assert.Equal("SELECT * FROM Users GROUP BY Name", wrapper.Query())
	wrapper.Table("Users").GroupBy("Name", "ID").Get()
	assert.Equal("SELECT * FROM Users GROUP BY Name, ID", wrapper.Query())
}

func TestJoin(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		Where("Users.ID", 6).
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products LEFT JOIN Users ON (Products.TenantID = Users.TenantID) WHERE Users.ID = ?", wrapper.Query())
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		RightJoin("Posts", "Products.TenantID = Posts.TenantID").
		Where("Users.ID", 6).
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products LEFT JOIN Users ON (Products.TenantID = Users.TenantID) RIGHT JOIN Posts ON (Products.TenantID = Posts.TenantID) WHERE Users.ID = ?", wrapper.Query())
}

func TestJoinWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		JoinOrWhere("Users", "Users.TenantID", 5).
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products LEFT JOIN Users ON (Products.TenantID = Users.TenantID OR Users.TenantID = ?)", wrapper.Query())
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		JoinWhere("Users", "Users.Username", "Wow").
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products LEFT JOIN Users ON (Products.TenantID = Users.TenantID AND Users.Username = ?)", wrapper.Query())
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		RightJoin("Posts", "Products.TenantID = Posts.TenantID").
		JoinWhere("Posts", "Posts.Username", "Wow").
		JoinWhere("Users", "Users.Username", "Wow").
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products LEFT JOIN Users ON (Products.TenantID = Users.TenantID AND Users.Username = ?) RIGHT JOIN Posts ON (Products.TenantID = Posts.TenantID AND Posts.Username = ?)", wrapper.Query())
}

func TestSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery()
	subQuery.Table("Products").Where("Quantity", ">", 2).Get("UserID")
	wrapper.Table("Users").Where("ID", "IN", subQuery).Get()
	assert.Equal("SELECT * FROM Users WHERE ID IN (SELECT UserID FROM Products WHERE Quantity > ?)", wrapper.Query())
}

func TestSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery()
	subQuery.Table("Users").Where("ID", 6).Get("Name")
	wrapper.Table("Products").Insert(map[string]interface{}{
		"ProductName": "測試商品",
		"UserID":      subQuery,
		"LastUpdated": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Products (ProductName, UserID, LastUpdated) VALUES (?, (SELECT Name FROM Users WHERE ID = ?), NOW())", wrapper.Query())
}

func TestSubQueryJoin(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery("Users")
	subQuery.Table("Users").Where("Active", 1).Get()
	wrapper.
		Table("Products").
		LeftJoin(subQuery, "Products.UserID = Users.ID").
		Get("Users.Username", "Products.ProductName")
	assert.Equal("SELECT Users.Username, Products.ProductName FROM Products LEFT JOIN (SELECT * FROM Users WHERE Active = ?) AS Users ON (Products.UserID = Users.ID)", wrapper.Query())
}

func TestSubQueryExist(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery("Users")
	subQuery.Table("Users").Where("Company", "測試公司").Get("UserID")
	wrapper.Table("Products").Where(subQuery, "EXISTS").Get()
	assert.Equal("SELECT * FROM Products WHERE EXISTS (SELECT UserID FROM Users WHERE Company = ?)", wrapper.Query())
}

func TestHas(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
	assert.Equal("SELECT * FROM Users WHERE Username = ? AND Password = ? LIMIT 1", wrapper.Query())
}
