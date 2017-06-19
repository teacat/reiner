# Reiner

一個由 Golang 撰寫且比起部分 ORM 還要讚的 MySQL 指令包覆函式庫。彈性高、不需要建構體標籤。實際上，這就只是 [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) 不過是用在 Golang 而已（但還是多了些功能）。

#  這是什麼？

一個由 Golang 撰寫的 MySQL 的指令包覆函式庫，幾乎所有東西都能操控於你手中。類似自己撰寫資料庫指令但是更簡單，JOIN 表格也變得比以前更方便了。

* 幾乎全功能的函式庫。
* 支援 MySQL 複寫橫向擴展機制（區分讀／寫連線）。
* 容易理解與記住、且使用方式十分簡單。
* SQL 指令建構函式。
* 資料庫表格建構協助函式。
* 支援子指令（Sub Query）。
* 可手動操作的交易機制（Transaction）和回溯（Rollback）功能。

# 為什麼？

[Gorm](https://github.com/jinzhu/gorm) 已經是 Golang 裡的 ORM 典範，但實際上要操作複雜與關聯性高的 SQL 指令時並不是很合適，而 Reiner 解決了這個問題。Reiner 也試圖不要和建構體扯上關係，不希望使用者需要手動指定任何標籤在建構體中。

# 執行緒與併發安全性？

我們都知道 Golang 的目標就是併發程式，當共用同個資料庫的時候請透過 `Copy()` 函式複製一份新的包覆函式庫，這能避免函式遭受干擾或覆寫。此方式並不會使資料庫連線遞增而造成效能問題，因此你可以有好幾個併發程式且有好幾個包覆函式庫的複製體都不會出現效能問題。

# 相關連結

這裡是 Reiner 受啟發，或是和資料庫有所關聯的連結。

[kisielk/sqlstruct](http://godoc.org/github.com/kisielk/sqlstruct)
[jmoiron/sqlx](https://github.com/jmoiron/sqlx)
[russross/meddler](https://github.com/russross/meddler)
[jinzhu/gorm](https://github.com/jinzhu/gorm)

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/TeaMeow/Reiner
```

# 使用方式

## 資料庫連線

### 基本

一個最基本的單資料庫連線，讀寫都將透過此連線，連線字串共用於其它套件是基於 DSN（Data Source Name）。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### 水平擴展（讀／寫分離）

這種方式可以有好幾個主要資料庫、副從資料庫，這意味著寫入時都會流向到主要資料庫，而讀取時都會向副從資料庫請求。這很適合用在大型結構還有水平擴展上。當你有多個資料庫來源時，Reiner 會逐一遞詢每個資料庫來源，英文稱其為 Round Robin，也就是每個資料庫都會輪流呼叫而避免單個資料庫負荷過重，也不會有隨機呼叫的事情發生。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/master?charset=utf8", []string{
	"root:root@/slave?charset=utf8",
	"root:root@/slave2?charset=utf8",
	"root:root@/slave3?charset=utf8",
})
if err != nil {
    panic(err)
}
```

## Insert

### Traditional/Replace

```go
err := db.Insert("users", map[string]string{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

### Functions

```go
id, err := db.Insert("users", map[string]interface{}{
	"username":  "YamiOdymel",
	"password":  db.Func("SHA1(?)", "secretpassword+salt"),
	"expires":   db.Now("+1Y"),
	"createdAt": db.Now(),
})
```

### On Duplicate

```go
lastInsertID := "id"

id, err := db.OnDuplicate([]string{"updatedAt"}, lastInsertID).Insert("users", map[string]interface{}{
	"username":  "YamiOdymel",
	"password":  "test",
	"createdAt": db.Now(),
})
```

### Multiple

```go
data := []map[string]string{
	map[string]string{
		"username": "YamiOdymel",
		"password": "test",
	},
	map[string]string{
		"username": "Karisu",
		"password": "12345",
	},
}

err := db.InsertMulti("users", data)
// ids := db.LastInsertIDs
```

## Update

```go
err := db.Where("username", "YamiOdymel").Update("users", map[string]string{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

### Limit

```go
err := db.Limit(10).Update("users", data)
```



## Select

```go
// Equals: SELECT * FROM users
rows, err := db.Get("users")
for rows.Next() {
	// rows.Scan(...)
}
```

### Limit

```go
// Equals: SELECT * FROM users LIMIT 10
rows, err := db.Limit(10).Get("users")
// for rows.Next() { ...
```

### Specified Columns

```go
// Equals: SELECT username, nickname FROM users
rows, err := db.Get("users", "username, nickname")
// for rows.Next() { ...
```

### Single Row

```go
rows, err := db.Where("id", 1).GetOne("users")
// or with the custom query.
rows, err := db.GetOne("users", "sum(id), count(*) as cnt")
```

### Get Value

```go
rows, err := db.GetValue("users", "username")
// or with the limit.
rows, err := db.Limit(5).GetValue("users", "username")
// or with the function.
rows, err := db.GetValue("users", "count(*)")
```

### Paginate

```go
page := 1
db.PageLimit = 2

rows, err := db.Paginate("users", page)
// fmt.Println("Showing %d out of %d", page, db.TotalPages)
```



## Raw Queries

### Common

```go
rows, err := db.RawQuery("SELECT * from users WHERE id >= ?", 10)
```

### Single Row

```go
row, err := db.RawQueryOne("SELECT * FROM users WHERE id = ?", 10)
```

### Single Value

```go
rows, err := db.RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", 10)
```

### Single Value From Multiple Rows

```go
rows, err := db.RawQueryValue("SELECT username FROM users LIMIT 10")
```

### Advanced

```go
rows, err := db.RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", 1, "admin")

// will handle any SQL query.
params := []int{10, 1, 10, 11, 2, 10}
query := `(
    SELECT a FROM t1
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
) UNION (
    SELECT a FROM t2
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
)`
rows, err := db.RawQuery(query, params...)
```



## Conditions

### Equals

```go
db.Where("id", 1)
db.Where("username", "admin")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE id=1 AND username='admin';
```

#### Having

```go
db.Where("id", 1)
db.Having("username", "admin")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE id=1 HAVING username='admin';
```

#### Columns Comparison

```go
// WRONG
db.Where("lastLogin", "createdAt")
// CORRECT
db.Where("lastLogin = createdAt")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE lastLogin = createdAt;
```

### Custom

```go
rows, err := db.Where("id", 50, ">=").Get("users")
// Equals: SELECT * FROM users WHERE id >= 50;
```

### Between / Not Between

```go
rows, err := db.Where("id", []int{0, 20}, "BETWEEN").Get("users")
// Equals: SELECT * FROM users WHERE id BETWEEN 4 AND 20
```

### In / Not In

```go
rows, err := db.Where("id", []interface{}{1, 5, 27, -1, "d"}, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (1, 5, 27, -1, 'd');
```

### Or / And Or

```go
db.Where("firstName", "John")
db.OrWhere("firstName", "Peter")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE firstName='John' OR firstName='peter'
```

### Null

```go
db.Where("lastName", nil, "IS NOT")

rows, err := db.Get("users")
// Equals: SELECT * FROM users where lastName IS NOT NULL
```

### Raw

```go
db.Where("id != companyId")
db.Where("DATE(createdAt) = DATE(lastLogin)")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE id != companyId AND DATE(createdAt) = DATE(lastLogin)
```

### Raw With Params

```go
db.Where("(id = ? or id = ?)", []int{6, 2})
db.Where("login", "mike")

rows, err := db.Get("users")
// Equals: SELECT * FROM users WHERE (id = 6 or id = 2) and login='mike';
```



## Delete

### Common

```go
err := db.Where("id", 1).Delete("users")
if err == nil && db.Count != 0 {
    fmt.Println("Deleted successfully!")
}
```



## Order

```go
db.OrderBy("id", "ASC")
db.OrderBy("login", "DESC")
db.OrderBy("RAND ()")

rows, err := db.Get("users")
// Equals: SELECT * FROM users ORDER BY id ASC,login DESC, RAND ();
```

### By Values

```go
db.OrderBy("userGroup", "ASC", []string{"superuser", "admin", "users"})
rows, err := db.Get("users")
// Equals: SELECT * FROM users ORDER BY FIELD (userGroup, 'superuser', 'admin', 'users') ASC;
```



## Group

```go
rows, err := db.GroupBy("name").Get("users")
// Equals: SELECT * FROM users GROUP BY name;
```



## Join

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.Where("u.id", 6)

rows, err := db.Get("products p", "u.name, p.productName")
```

### Conditions

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinWhere("users u", "u.tenantID", 5)

rows, err := db.Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID AND u.tenantID = 5)
```

```go
// db.InnerJoin()
// db.LeftJoin()
// db.RightJoin()
// db.NaturalJoin()
// db.CrossJoin()
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinOrWhere("users u", "u.tenantID", "=", 5)

rows, err := db.Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID OR u.tenantID = 5)
```



## Subqueries

```go
subQuery := db.SubQuery()
subQuery.Get("users")
```

```go
subQuery := db.SubQuery("sq")
subQuery.Get("users")
```

### Select

```go
idSubQuery := db.SubQuery()
idSubQuery.Where("qty", 2, ">").Get("products", "userId")

rows, err := db.Where("id", idSubQuery, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (SELECT userId FROM products WHERE qty > 2)
```

### Insert

```go
idSubQuery := db.SubQuery()
idSubQuery.Where("id", 6).GetOne("users", "name")

err := db.Insert("products", map[string]interface{}{
	"productName": "test product",
	"userID":      idSubQuery,
	"lastUpdated": db.Now(),
})
// Equals: INSERT INTO PRODUCTS (productName, userId, lastUpdated) values ("test product", (SELECT name FROM users WHERE id = 6), NOW());
```

### Join

```go
userSubQuery := db.SubQuery("u")
userSubQuery.Where("active", 1).Get("users")

rows, err := db.Join(userSubQuery, "p.userId = u.id", "LEFT").Get("products p", "u.login, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN (SELECT * FROM t_users WHERE active = 1) u on p.userId=u.id;
```

### Exist / Not Exist

```go
subQuery := db.SubQuery()
subQuery.Where("company", "testCompany")
subQuery.Get("users", "userId")

rows, err := db.Where("", subQuery, "EXISTS").Get("products")
// Equals: SELECT * FROM products WHERE EXISTS (select userId from users where company='testCompany')
```

## Has

```go
db.Where("username", "yamiodymel")
db.Where("password", "123456")

if db.Has("users") {
	fmt.Println("Logged in successfully!")
} else {
	fmt.Println("Incorrect username or the password.")
}
```

## Helpers

### Connection

```go
db.Disconnect()
```

```go
if !db.Ping() {
	db.Connect()
}
```

### Last Query

```go
rows, err := db.Get("users")
// And ... Get the last executed query like this.
fmt.Println("Last executed query was %s", db.LastQuery)
```

## Transactions

```go
err := db.Begin().Insert("myTable", data)
if err != nil {
	db.Rollback()
} else {
	db.Commit()
}
```

## Lock

```go
db.SetLockMethod("WRITE").Lock("users")

// Calling another `Lock()` will unlock the first lock. You could also use `Unlock()`.
db.Unlock()

// Lock the multiple tables at the same time is easy.
db.SetLockMethod("READ").Lock("users", "log")
```



## Query Keywords

### Common

```go
db.SetQueryOption("LOW_PRIORITY").Insert("users", data)
// Equals: INSERT LOW_PRIORITY INTO table ...

db.SetQueryOption("FOR UPDATE").Get("users")
// Equals: SELECT * FROM users FOR UPDATE;

db.SetQueryOption("SQL_NO_CACHE").Get("users")
// Equals: GIVES: SELECT SQL_NO_CACHE * FROM users;
```

### Multiple

```go
db.SetQueryOption("LOW_PRIORITY", "IGNORE").Insert("users", data)
// Gives: INSERT LOW_PRIORITY IGNORE INTO users ...
```

# Table Migrations

```go
migration := db.Migration()

migration.Column("test").Varchar(32).Primary().CreateTable("test_table")
// Equals: CREATE TABLE `test_table` (`test` varchar(32) NOT NULL PRIMARY KEY) ENGINE=INNODB
```


| Numbers   | Strings    | Binaries  | Blobs      | Times     | Floatings | Enums |
|-----------|------------|-----------|------------|-----------|-----------|-------|
| TinyInt   | Char       | Binary    | Blob       | Date      | Double    | Enum  |
| SmallInt  | Varchar    | VarBinary | MediumBlob | DateTime  | Decimal   | Set   |
| MediumInt | TinyText   | Bit       | LongBlob   | Time      | Float     |       |
| Int       | Text       |           |            | Timestamp |           |       |
| BigInt    | MediumText |           |            | Year      |           |       |
|           | LongText   |           |            |           |           |       |