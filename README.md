# Reiner

A Golang MySQL wrapper which is better than some ORMs. Flexible, and no struct tags needed. More actually, it's just [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) but in Golang (also with more functions).

# What is it?

A MySQL wrapper written in Golang which lets you controll everything, just like writing a query but simpler, join tables are now easier than before.

* Almost full-featured wrapper
* MySQL replication supported (read/write split)
* Easy to remember, understand
* SQL builder
* Table migrations
* Sub queries
* Transactions

# Why?

[Gorm](https://github.com/jinzhu/gorm) is great, but it's not really fits with a complex SQL query usage, and Reiner solved the problem. Reiner also decoupling the function usage with the struct (Loose coupling).

# Thread Safe?

# Field Naming

[kisielk/sqlstruct](http://godoc.org/github.com/kisielk/sqlstruct)
[jmoiron/sqlx](https://github.com/jmoiron/sqlx)
[russross/meddler](https://github.com/russross/meddler)
[jinzhu/gorm](https://github.com/jinzhu/gorm)

# Installtion

```bash
$ go get github.com/TeaMeow/Reiner
```

# Usage

## Conenction

### Common

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### Replication (Read / Write Split)

Round Robin

`slave` -> `slave2` -> `slave3` -> `slave` ...

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/master?charset=utf8", []string{
	"root:root@/slaveReadOnly?charset=utf8",
	"root:root@/slaveReadOnly2?charset=utf8",
	"root:root@/slaveReadOnly3?charset=utf8",
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
db.Where("lastName", reiner.NULL, "IS NOT")

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
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinOrWhere("users u", "u.tenantID", 5)

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