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
		Table("Users").
		Column("Username").Varchar(32).Primary().
		Column("Password").Varchar(64).
		Column("Age").Int(2).
		Create()
	assert.NoError(err)
}

type user struct {
	Username string `db:"Username"`
	Password string `db:"Password"`
	Age      int    `db:"Age"`
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Insert(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
	})
	assert.Equal("INSERT INTO Users (Username, Password) VALUES (?, ?)", wrapper.LastQuery)
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
	assert.Equal("INSERT INTO Users (Username, Password) VALUES (?, ?), (?, ?)", wrapper.LastQuery)
}

func TestReplace(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Replace(map[string]interface{}{
		"Username": "YamiOdymel",
		"Password": "test",
	})
	assert.Equal("REPLACE INTO Users (Username, Password) VALUES (?, ?)", wrapper.LastQuery)
}

func TestInsertFunc(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Insert(map[string]interface{}{
		"Username":  "YamiOdymel",
		"Password":  wrapper.Func("SHA1(?)", "secretpassword+salt"),
		"Expires":   wrapper.Now("+1Y"),
		"CreatedAt": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Users (Username, Password, Expires, CreatedAt) VALUES (?, SHA1(?), NOW() + INTERVAL 1 YEAR, NOW())", wrapper.LastQuery)
}

func TestOnDuplicateInsert(t *testing.T) {
	assert := assert.New(t)
	lastInsertID := "ID"
	wrapper.Table("Users").OnDuplicate([]string{"UpdatedAt"}, lastInsertID).Insert(map[string]interface{}{
		"Username":  "YamiOdymel",
		"Password":  "test",
		"UpdatedAt": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Users (Username, Password, UpdatedAt) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE ID=LAST_INSERT_ID(ID), UpdatedAt = VALUE(UpdatedAt)", wrapper.LastQuery)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("Username", "YamiOdymel").Update(map[string]interface{}{
		"Username": "Karisu",
		"Password": "123456",
	})
	assert.Equal("UPDATE Users SET Username = ?, Password = ? WHERE Username = ?", wrapper.LastQuery)
}

func TestLimitUpdate(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Limit(10).Update(map[string]interface{}{
		"Username": "Karisu",
		"Password": "123456",
	})
	assert.Equal("UPDATE Users SET Username = ?, Password = ? LIMIT 10", wrapper.LastQuery)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Get()
	assert.Equal("SELECT * FROM Users", wrapper.LastQuery)
}

func TestLimitGet(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Limit(10).Get()
	assert.Equal("SELECT * FROM Users LIMIT 10", wrapper.LastQuery)
}

func TestGetColumns(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Get("Username", "Nickname")
	assert.Equal("SELECT Username, Nickname FROM Users", wrapper.LastQuery)

	wrapper.Table("Users").Get("COUNT(*) AS Count")
	assert.Equal("SELECT COUNT(*) AS Count FROM Users", wrapper.LastQuery)
}

func TestGetOne(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).GetOne()
	assert.Equal("SELECT * FROM Users WHERE ID = ?", wrapper.LastQuery)

	wrapper.Table("Users").GetOne("SUM(ID)", "COUNT(*) AS Count")
	assert.Equal("SELECT SUM(ID), COUNT(*) AS Count FROM Users", wrapper.LastQuery)
}

func TestGetValue(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").GetValue("Username")
	assert.Equal("SELECT Username AS Value FROM Users", wrapper.LastQuery)

	wrapper.Table("Users").Limit(5).GetValue("Username")
	assert.Equal("SELECT Username AS Value FROM Users LIMIT 5", wrapper.LastQuery)

	wrapper.Table("Users").GetValue("COUNT(*)")
	assert.Equal("SELECT COUNT(*) AS Value FROM Users", wrapper.LastQuery)
}

func TestPaginate(t *testing.T) {
	assert := assert.New(t)
	wrapper.PageLimit = 20
	wrapper.Table("Users").Paginate(2)
	assert.Equal("SELECT * FROM Users LIMIT 20, 40", wrapper.LastQuery)
}

func TestRawQuery(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQuery("SELECT * FROM Users WHERE ID >= ?", 10)
	assert.Equal("SELECT * FROM Users WHERE ID >= ?", wrapper.LastQuery)
}

func TestRawQueryOne(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQueryOne("SELECT * FROM Users WHERE ID = ?", 10)
	assert.Equal("SELECT * FROM Users WHERE ID = ?", wrapper.LastQuery)
}

func TestRawQueryValue(t *testing.T) {
	assert := assert.New(t)
	wrapper.RawQueryValue("SELECT Password FROM Users WHERE ID = ? LIMIT 1", 10)
	assert.Equal("SELECT Password FROM Users WHERE ID = ? LIMIT 1", wrapper.LastQuery)
}

func TestWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Where("Username", "admin").Get()
	assert.Equal("SELECT * FROM Users WHERE ID = ? AND Username = ?", wrapper.LastQuery)
}

func TestWhereHaving(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Having("Username", "admin").Get()
	assert.Equal("SELECT * FROM Users WHERE ID = ? HAVING Username = ?", wrapper.LastQuery)
}

func TestWhereColumns(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("LastLogin = CreatedAt").Get()
	assert.Equal("SELECT * FROM Users WHERE LastLogin = CreatedAt", wrapper.LastQuery)
}

func TestWhereOperator(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", ">=", 50).Get()
	assert.Equal("SELECT * FROM Users WHERE ID >= ?", wrapper.LastQuery)
}

func TestWhereBetween(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").WhereBetween("ID", []int{0, 20}).Get()
	assert.Equal("SELECT * FROM Users WHERE ID BETWEEN ? AND ?", wrapper.LastQuery)

	wrapper.Table("Users").WhereNotBetween("ID", []int{0, 20}).Get()
	assert.Equal("SELECT * FROM Users WHERE ID NOT BETWEEN ? AND ?", wrapper.LastQuery)
}

func TestWhereIn(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").WhereIn("ID", []interface{}{1, 5, 27, -1, "d"}).Get()
	assert.Equal("SELECT * FROM Users WHERE ID IN (?, ?, ?, ?, ?)", wrapper.LastQuery)

	wrapper.Table("Users").WhereNotIn("ID", []interface{}{1, 5, 27, -1, "d"}).Get()
	assert.Equal("SELECT * FROM Users WHERE ID IN (?, ?, ?, ?, ?)", wrapper.LastQuery)
}

func TestOrWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("FirstNamte", "John").OrWhere("FirstNamte", "Peter").Get()
	assert.Equal("SELECT * FROM Users WHERE FirstName = ? OR FirstName = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("A = B").OrWhere("(A = C OR A = D)").Get()
	assert.Equal("SELECT * FROM Users WHERE A = B OR (A = C OR A = D)", wrapper.LastQuery)
}

func TestWhereNull(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").WhereNull("LastName").Get()
	assert.Equal("SELECT * FROM Users WHERE LastName IS NULL", wrapper.LastQuery)

	wrapper.Table("Users").WhereNotNull("LastName").Get()
	assert.Equal("SELECT * FROM Users WHERE LastName IS NOT NULL", wrapper.LastQuery)
}

func TestTimestampRelative(t *testing.T) {
	assert := assert.New(t)
	ts := wrapper.Timestamp
	wrapper.Table("Users").Where("CreatedAt", ts.Now("-1Y")).Get()
	assert.Equal("SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.Now("-1D")).Get()
	assert.Equal("SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)
}

func TestTimestampDate(t *testing.T) {
	assert := assert.New(t)
	ts := wrapper.Timestamp
	wrapper.Table("Users").Where("CreatedAt", ts.IsDate("2017-07-13")).Get()
	assert.Equal("SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsYear(2017)).Get()
	assert.Equal("SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsMonth(1)).Get()
	assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)
	wrapper.Table("Users").Where("CreatedAt", ts.IsMonth("January")).Get()
	assert.Equal("SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsDay(16)).Get()
	assert.Equal("SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday("Friday")).Get()
	assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)
}

func TestTimestampTime(t *testing.T) {
	assert := assert.New(t)
	ts := wrapper.Timestamp
	wrapper.Table("Users").Where("CreatedAt", ts.IsHour(18)).Get()
	assert.Equal("SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsMinute(25)).Get()
	assert.Equal("SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsSecond(16)).Get()
	assert.Equal("SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)

	wrapper.Table("Users").Where("CreatedAt", ts.IsWeekday(5)).Get()
	assert.Equal("SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?", wrapper.LastQuery)
}

func TestRawWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID != CompanyID").Where("DATE(CreatedAt) = DATE(LastLogin)").Get()
	assert.Equal("SELECT * FROM Users WHERE ID != CompanyID AND DATE(CreatedAt) = DATE(LastLogin)", wrapper.LastQuery)
}

func TestRawWhereParams(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("(ID = ? OR ID = ?)", []int{6, 2}).Where("Login", "Mike").Get()
	assert.Equal("SELECT * FROM Users WHERE (ID = ? OR ID = ?) AND Login = ?", wrapper.LastQuery)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("ID", 1).Delete()
	assert.Equal("DELETE FROM Users WHERE ID = ?", wrapper.LastQuery)
}

func TestOrderBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").OrderBy("ID", "ASC").OrderBy("Login", "DESC").OrderBy("RAND()").Get()
	assert.Equal("SELECT * FROM Users ORDER BY ID ASC, Login DESC, RAND();", wrapper.LastQuery)
}

func TestOrderByField(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").OrderBy("UserGroup", "ASC", []string{"SuperUser", "Admin", "Users"}).Get()
	assert.Equal("SELECT * FROM Users ORDER BY FIELD (UserGroup, ?, ?, ?) ASC", wrapper.LastQuery)
}

func TestGroupBy(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").GroupBy("Name").Get()
	assert.Equal("SELECT * FROM Users GROUP BY Name", wrapper.LastQuery)
}

func TestJoin(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		Where("Users.ID", 6).
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products AS Products LEFT JOIN Users AS Users ON (Products.TenantID = Users.TenantID) WHERE Users.ID = ?", wrapper.LastQuery)
}

func TestJoinWhere(t *testing.T) {
	assert := assert.New(t)
	wrapper.
		Table("Products").
		LeftJoin("Users", "Products.TenantID = Users.TenantID").
		JoinWhere("Users", "Users.Username", "Wow").
		JoinOrWhere("Users", "Users.TenantID", 5).
		Get("Users.Name", "Products.ProductName")
	assert.Equal("SELECT Users.Name, Products.ProductName FROM Products AS Products LEFT JOIN Users AS Users ON (Products.TenantID = Users.TenantID OR Users.TenantID = ?)", wrapper.LastQuery)
}

func TestSubQueryGet(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery()
	subQuery.Table("Products").Where("Quantity", 2, ">").Get("UserID")
	wrapper.Table("Users").Where("ID", subQuery, "IN").Get()
	assert.Equal("SELECT * FROM Users WHERE ID IN (SELECT UserID FROM Products WHERE Quantity > ?)", wrapper.LastQuery)
}

func TestSubQueryInsert(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery()
	subQuery.Table("Users").Where("ID", 6).GetOne("Name")
	wrapper.Table("Products").Insert(map[string]interface{}{
		"ProductName": "測試商品",
		"UserID":      subQuery,
		"LastUpdated": wrapper.Now(),
	})
	assert.Equal("INSERT INTO Products (ProductName, UserID, LastUpdated) VALUES (?, (SELECT Name FROM Users WHERE ID = 6), NOW())", wrapper.LastQuery)
}

func TestSubQueryJoin(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery("Users")
	subQuery.Table("Users").Where("Active", 1).Get()
	wrapper.
		Table("Products").
		LeftJoin(subQuery, "Products.UserID = U.ID").
		Get("U.Username", "Products.ProductName")
	assert.Equal("SELECT Users.Username, Products.ProductName FROM Products AS Products LEFT JOIN (SELECT * FROM Users WHERE Active = ?) AS Users ON Products.UserID = Users.ID", wrapper.LastQuery)
}

func TestSubQueryExist(t *testing.T) {
	assert := assert.New(t)
	subQuery := wrapper.SubQuery("Users")
	subQuery.Table("Users").Where("Company", "測試公司").Get("UserID")
	wrapper.Table("Products").Where(subQuery, "EXISTS").Get()
	assert.Equal("SELECT * FROM Products WHERE EXISTS (SELECT UserID FROM Users WHERE Company = ?)", wrapper.LastQuery)
}

func TestHas(t *testing.T) {
	assert := assert.New(t)
	wrapper.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
	assert.Equal("SELECT * FROM Users WHERE Username = ? AND Password = ?", wrapper.LastQuery)
}
