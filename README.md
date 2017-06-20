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
* 透過預置聲明（Prepared Statement），99.9% 避免 SQL 插入攻擊。
* 自動脫逸表格名稱，避免觸動保留字。

# 為什麼？

[Gorm](https://github.com/jinzhu/gorm) 已經是 Golang 裡的 ORM 典範，但實際上要操作複雜與關聯性高的 SQL 指令時並不是很合適，而 Reiner 解決了這個問題。Reiner 也試圖不要和建構體扯上關係，不希望使用者需要手動指定任何標籤在建構體中。

# 執行緒與併發安全性？

我們都知道 Golang 的目標就是併發程式，當共用同個資料庫的時候請透過 `Copy()` 函式複製一份新的包覆函式庫，這能避免函式遭受干擾或覆寫。此方式並不會使資料庫連線遞增而造成效能問題，因此你可以有好幾個併發程式且有好幾個包覆函式庫的複製體都不會出現效能問題。

# 相關連結

這裡是 Reiner 受啟發，或是和資料庫有所關聯的連結。

* [kisielk/sqlstruct](http://godoc.org/github.com/kisielk/sqlstruct)
* [jmoiron/sqlx](https://github.com/jmoiron/sqlx)
* [russross/meddler](https://github.com/russross/meddler)
* [jinzhu/gorm](https://github.com/jinzhu/gorm)

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/TeaMeow/Reiner
```

# 使用方式

Reiner 的使用方式十分直覺與簡易，類似基本的 SQL 指令集但是更加地簡化了。

## 資料庫連線

首先你需要透過函式來將 Reiner 連上資料庫，如此一來才能夠初始化包覆函式庫與相關的資料庫表格建構函式。一個最基本的單資料庫連線，讀寫都將透過此連線，連線字串共用於其它套件是基於 DSN（Data Source Name）。

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

## 插入

透過 Reiner 你可以很輕鬆地透過建構體或是 `map` 來插入一筆資料。這是最傳統的插入方式，若該表格有自動遞增的編號欄位，插入後你就能透過 `LastInsertID` 獲得最後一次插入的編號。

```go
err := db.Insert("users", map[string]string{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

### 覆蓋

```go
```

### 函式

插入時你可以透過 Reiner 提供的函式來執行像是 `SHA1()` 或者取得目前時間的 `NOW()`，甚至將目前時間加上一年⋯等。

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

### 多筆資料

Reiner 允許你透過 `InsertMulti` 同時間插入多筆資料，這省去了透過迴圈不斷執行單筆插入的困擾，這種方式亦大幅度提升了效能。

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

## 更新

更新一筆資料在 Reiner 中極為簡單，你只需要指定表格名稱還有資料即可。

```go
err := db.Where("username", "YamiOdymel").Update("users", map[string]string{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

### 筆數限制

`Limit` 能夠限制更新的筆數，如果是 `10`，那就表示只更新最前面 10 筆資料而非全部。

```go
err := db.Limit(10).Update("users", data)
```

## 選擇與取得

最基本的選擇在 Reiner 中稱之為 `Get` 而不是 `Select`。如果你想要取得 `rows.Next` 來掃描每一行的結果，Reiner 提供了 `LastRows` 即為最後一次的 `rows` 資料。

```go
// Equals: SELECT * FROM users
err := db.Get("users")
// rows := db.LastRows
// for rows.Next() {
//     rows.Scan(...)
// }
```

### 筆數限制

`Limit` 能夠限制取得的筆數，如果是 `10`，那就表示只取得最前面 10 筆資料而非全部。

```go
// Equals: SELECT * FROM users LIMIT 10
err := db.Limit(10).Get("users")
```

### 指定欄位

你可以透過 `Columns` 指定要取得的欄位名稱，亦能是個函式。

```go
// Equals: SELECT username, nickname FROM users
err := db.Columns("username", "nickname").Get("users")
// Equals: SELECT COUNT(*) AS count FROM users
err := db.Columns("COUNT(*) AS count").Get("users")
```

### 單行資料

預設來說 `Get` 會回傳一個切片或是陣列，這令你需要透過迴圈逐一取得資料，但某些情況下你很確信你僅要取得一筆資料的話，可以嘗試 `GetOne`。這能將資料直接映射到單個建構體上而避免你需要透過迴圈處理資料的麻煩。

```go
err := db.Where("id", 1).GetOne("users")
// 或者像這樣使用函式。
err := db.Columns("SUM(id)", "COUNT(*) AS cnt").GetOne("users")
```

### 取得單值

這就像 `GetOne`，但 `GetValue` 取得的是單個欄位的內容，例如說你想要單個使用者的暱稱，甚至是多個使用者的暱稱陣列就很適用。

```go
err := db.Columns("username").GetValue("users")
// 也能搭配 Limit。
err := db.Limit(5).Columns("username").GetValue("users")
// 或者是函式。
err := db.Columns("COUNT(*)").GetValue("users")
```

### 分頁功能

分頁就像是取得資料ㄧ樣，但更擅長用於多筆資料、不會一次顯示完畢的內容。Reiner 能夠幫你自動處理換頁功能，讓你不需要自行計算換頁時的筆數應該從何開始。為此，你需要定義兩個變數，一個是目前的頁數，另一個是單頁能有幾筆資料。

```go
page := 1
db.PageLimit = 2

err := db.Paginate("users", page)
// fmt.Println("目前頁數為 %d，共有 %d 頁", page, db.TotalPages)
```



## Raw Queries

### Common

```go
err := db.RawQuery("SELECT * from users WHERE id >= ?", 10)
```

### Single Row

```goerr := db.RawQueryOne("SELECT * FROM users WHERE id = ?", 10)
```

### Single Value

```go
err := db.RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", 10)
```

### Single Value From Multiple Rows

```go
err := db.RawQueryValue("SELECT username FROM users LIMIT 10")
```

### Advanced

```go
err := db.RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", 1, "admin")

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
err := db.RawQuery(query, params...)
```



## Conditions

### Equals

```go
db.Where("id", 1)
db.Where("username", "admin")

err := db.Get("users")
// Equals: SELECT * FROM users WHERE id=1 AND username='admin';
```

#### Having

```go
db.Where("id", 1)
db.Having("username", "admin")

err := db.Get("users")
// Equals: SELECT * FROM users WHERE id=1 HAVING username='admin';
```

#### Columns Comparison

```go
// WRONG
db.Where("lastLogin", "createdAt")
// CORRECT
db.Where("lastLogin = createdAt")

err := db.Get("users")
// Equals: SELECT * FROM users WHERE lastLogin = createdAt;
```

### Custom

```go
err := db.Where("id", 50, ">=").Get("users")
// Equals: SELECT * FROM users WHERE id >= 50;
```

### Between / Not Between

```go
err := db.Where("id", []int{0, 20}, "BETWEEN").Get("users")
// Equals: SELECT * FROM users WHERE id BETWEEN 4 AND 20
```

### In / Not In

```go
err := db.Where("id", []interface{}{1, 5, 27, -1, "d"}, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (1, 5, 27, -1, 'd');
```

### Or / And Or

```go
db.Where("firstName", "John")
db.OrWhere("firstName", "Peter")

err := db.Get("users")
// Equals: SELECT * FROM users WHERE firstName='John' OR firstName='peter'
```

### Null

```go
db.Where("lastName", nil, "IS NOT")

err := db.Get("users")
// Equals: SELECT * FROM users where lastName IS NOT NULL
```

### Raw

```go
db.Where("id != companyId")
db.Where("DATE(createdAt) = DATE(lastLogin)")

err := db.Get("users")
// Equals: SELECT * FROM users WHERE id != companyId AND DATE(createdAt) = DATE(lastLogin)
```

### Raw With Params

```go
db.Where("(id = ? or id = ?)", []int{6, 2})
db.Where("login", "mike")

err := db.Get("users")
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

err := db.Get("users")
// Equals: SELECT * FROM users ORDER BY id ASC,login DESC, RAND ();
```

### By Values

```go
db.OrderBy("userGroup", "ASC", []string{"superuser", "admin", "users"})
err := db.Get("users")
// Equals: SELECT * FROM users ORDER BY FIELD (userGroup, 'superuser', 'admin', 'users') ASC;
```



## Group

```go
err := db.GroupBy("name").Get("users")
// Equals: SELECT * FROM users GROUP BY name;
```



## Join

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.Where("u.id", 6)

err := db.Get("products p", "u.name, p.productName")
```

### Conditions

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinWhere("users u", "u.tenantID", 5)

err := db.Get("products p", "u.name, p.productName")
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

err := db.Get("products p", "u.name, p.productName")
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

err := db.Where("id", idSubQuery, "IN").Get("users")
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

err := db.Join(userSubQuery, "p.userId = u.id", "LEFT").Get("products p", "u.login, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN (SELECT * FROM t_users WHERE active = 1) u on p.userId=u.id;
```

### Exist / Not Exist

```go
subQuery := db.SubQuery()
subQuery.Where("company", "testCompany")
subQuery.Get("users", "userId")

err := db.Where("", subQuery, "EXISTS").Get("products")
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
err := db.Get("users")
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