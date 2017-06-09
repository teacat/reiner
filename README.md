# Reiner

A Golang database ORM with the 1990's style. Flexible, and no struct tags needed. More actually, it's just [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) but in Golang.

# What is it?

A MySQL ORM written in Golang which lets you controll everything, just like writing a query but simpler, join tables are now easier than before.

# Why?

[Gorm](https://github.com/jinzhu/gorm) is great as fuck, but it's not really fits with a complex SQL query usage, and Reiner solved the problem. Reiner also decoupling the function usage with the struct (Loose coupling).

# Thread Safe?

# Installtion

```bash
$ go get github.com/TeaMeow/Reiner
```

# Helper Types

```go
reiner.Fields{
    "name": "value",
}

reiner.Values{"value1", "value2"}

reiner.Options{

}
```

# Usage

## Conenction

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("yamiodymel:yamiodymel@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

&nbsp;

## Insert

### Traditional/Replace

```go
err := db.Insert("users", reiner.Fields{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

### Functions

```go
err := db.Insert("users", reiner.Fields{
	"username":  "YamiOdymel",
	"password":  db.Func("SHA1(?)", reiner.Values{"secretpassword+salt"}),
	"expires":   db.Now("+1Y"),
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

### On Duplicate

```go
lastInsertID := "id"

err := db.Columns("updatedAt").OnDuplicate(lastInsertID).Insert("users", reiner.Fields{
	"username":  "YamiOdymel",
	"password":  "test",
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

### Multiple

```go
data := reiner.FieldGroup{
	reiner.Fields{
		"username": "YamiOdymel",
		"password": "test",
	},
	reiner.Fields{
		"username": "Karisu",
		"password": "12345",
	},
}

err := db.InsertMulti("users", data)
// ids := db.LastInsertIDs
```

&nbsp;

## Update

```go
err := db.Where("username", "YamiOdymel").Update("users", reiner.Fields{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

### Limit

```go
err := db.Limit(10).Update("users", data)
```

&nbsp;

## Select

```go
err := db.Bind(&users).Get("users")
```

### Limit

```go
err := db.Bind(&users).Limit(10).Get("users")
```

### Specified Columns

```go
err := db.Bind(&users).Columns("username", "nickname").Get("users")
// count := db.Count
```

### Single Row

```go
err := db.Bind(&user).Where("id", 1).GetOne("users")
// or with the custom query.
err := db.Bind(&stats).GetOne("users", reiner.Option{
	Query: "sum(id), count(*) as cnt",
})
```

### Get Value

```go
err := db.Bind(&usernames).GetValue("users", "username")
// or with the limit.
err := db.Bind(&usernames).Limit(5).GetValue("users", "username")
// or with the function.
err := db.Bind(&total).GetValue("users", "count(*)")
```

### Paginate

```go
page := 1
db.PageLimit = 2

err := db.Bind(&users).Paginate("users", page)
// fmt.Println("Showing %d out of %d", page, db.TotalPages)
```

&nbsp;

## Raw Queries

### Common

```go
err := db.Bind(&users).RawQuery("SELECT * from users WHERE id >= ?", reiner.Values{10})
```

### Single Row

```go
err := db.Bind(&user).RawQueryOne("SELECT * FROM users WHERE id = ?", reiner.Values{10})
```

### Single Value

```go
err := db.Bind(&password).RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", reiner.Values{10})
```

### Single Value From Multiple Rows

```go
err := db.Bind(&usernames).RawQueryValue("SELECT username FROM users LIMIT 10")
```

### Advanced

```go
params := reiner.Values{1, "admin"}
err := db.Bind(&users).RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", params)

// will handle any SQL query.
params = reiner.Values{10, 1, 10, 11, 2, 10}
query := "(
    SELECT a FROM t1
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
) UNION (
    SELECT a FROM t2
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
)"
err := db.Bind(&results).RawQuery(query, params)
```

&nbsp;

## Conditions

### Equals

```go
db.Where("id", 1)
db.Where("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 AND username='admin';
```

#### Having

```go
db.Where("id", 1)
db.Having("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 HAVING username='admin';
```

#### Columns Comparison

```go
// WRONG
db.Where("lastLogin", "createdAt")
// CORRECT
db.Where("lastLogin = createdAt")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE lastLogin = createdAt;
```

### Custom

```go
db.Bind(&users).Where("id", 50, ">=").Get("users")
// Equals: SELECT * FROM users WHERE id >= 50;
```

### Between / Not Between

```go
db.Bind(&users).Where("id", reiner.Values{0, 20}, "BETWEEN").Get("users")
// Equals: SELECT * FROM users WHERE id BETWEEN 4 AND 20
```

### In / Not In

```go
db.Bind(&users).Where("id", reiner.Values{1, 5, 27, -1, "d"}, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (1, 5, 27, -1, 'd');
```

### Or / And Or

```go
db.Where("firstName", "John")
db.OrWhere("firstName", "Peter")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE firstName='John' OR firstName='peter'
```

### Null

```go
db.Where("lastName", reiner.NULL, "IS NOT")
db.Bind(&users).Get("users")
// Equals: SELECT * FROM users where lastName IS NOT NULL
```

### Raw

```go
db.Where("id != companyId")
db.Where("DATE(createdAt) = DATE(lastLogin)")
db.Bind(&users).Get("users")
```

### Raw With Params

```go
db.Where("(id = ? or id = ?)", reiner.Fields{6, 2})
db.Where("login", "mike")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE (id = 6 or id = 2) and login='mike';
```

&nbsp;

## Delete

### Common

```go
err := db.Where("id", 1).Delete("users")
if err != nil {
    panic(err)
}
if db.Count != 0 {
    fmt.Println("Deleted successfully!")
}
```

&nbsp;

## Order

```go
db.OrderBy("id", "ASC")
db.OrderBy("login", "DESC")
db.OrderBy("RAND ()")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users ORDER BY id ASC,login DESC, RAND ();
```

### By Values

```go
db.OrderBy("userGroup", "ASC", []string{"superuser", "admin", "users"})
db.Bind(&users).Get("users")
// Equals: SELECT * FROM users ORDER BY FIELD (userGroup, 'superuser', 'admin', 'users') ASC;
```

&nbsp;

## Group

```go
db.GroupBy("name").Bind(&users).Get("users")
// Equals: SELECT * FROM users GROUP BY name;
```

&nbsp;

## Join

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.Where("u.id", 6)

db.Bind(&products).Get("products p", "u.name, p.productName")
```

### Conditions

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinWhere("users u", "u.tenantID", 5)

db.Bind(&products).Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID AND u.tenantID = 5)
```

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinOrWhere("users u", "u.tenantID", 5)

db.Bind(&products).Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID OR u.tenantID = 5)
```

&nbsp;

## Subqueries

```go
sq := db.SubQuery()
sq.Get("users")
```

```go
sq := db.SubQuery("sq")
sq.Get("users")
```

```go
ids := db.SubQuery()

```

$ids = $db->subQuery ();
$ids->where ("qty", 2, ">");
$ids->get ("products", null, "userId");

$db->where ("id", $ids, 'in');
$res = $db->get ("users");
// Gives SELECT * FROM users WHERE id IN (SELECT userId FROM products WHERE qty > 2)

&nbsp;

## Has

&nbsp;

## Helpers

&nbsp;

## Transactions

&nbsp;

## Lock

&nbsp;

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
// GIVES: INSERT LOW_PRIORITY IGNORE INTO users ...
```