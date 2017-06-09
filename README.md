# Reiner

The Golang database ORM with 1990s style.

## What is it?

## Why?

## Thread Safe?

## Usage

### Conenction

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("yamiodymel:yamiodymel@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### Insert

#### Traditional/Replace

```go
err := db.Insert("users", reiner.H{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

#### Functions

```go
err := db.Insert("users", reiner.H{
	"username":  "YamiOdymel",
	"password":  db.Func("SHA1(?)", []string{"secretpassword+salt"}),
	"expires":   db.Now("+1Y"),
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

#### On Duplicate

```go
updateColumns := []string{"updatedAt"}
lastInsertID := "id"

err := db.OnDuplicate(updateColumns, lastInsertID).Insert("users", reiner.H{
	"username":  "YamiOdymel",
	"password":  "test",
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

#### Multiple

```go
data := reiner.Hs{
	reiner.H{
		"username": "YamiOdymel",
		"password": "test",
	},
	reiner.H{
		"username": "Karisu",
		"password": "12345",
	},
}

err := db.InsertMulti("users", data)
// ids := db.LastInsertIDs
```

### Update

```go
err := db.Where("username", "YamiOdymel").Update("users", reiner.H{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

#### Limit

```go
err := db.Limit(10).Update("users", data)
```

### Select

```go
err := db.Bind(&users).Get("users")
```

#### Limit

```go
err := db.Bind(&users).Limit(10).Get("users")
```

#### Specified Columns

```go
err := db.Bind(&users).Columns("username", "nickname").Get("users")
// count := db.Count
```

#### Single Row

```go
err := db.Bind(&user).Where("id", 1).GetOne("users")
// or with the custom query.
err := db.Bind(&stats).GetOne("users", reiner.O{
	Query: "sum(id), count(*) as cnt",
})
```

#### Get Value

```go
err := db.Bind(&usernames).GetValue("users", "username")
// or with the limit.
err := db.Bind(&usernames).Limit(5).GetValue("users", "username")
// or with the function.
err := db.Bind(&total).GetValue("users", "count(*)")
```

#### Paginate

```go
page := 1
db.PageLimit = 2

err := db.Bind(&users).Paginate("users", page)
// fmt.Println("Showing %d out of %d", page, db.TotalPages)
```

### Raw Queries

#### Common

```go
err := db.Bind(&users).RawQuery("SELECT * from users WHERE id >= ?", reiner.V{10})
```

#### Single Row

```go
err := db.Bind(&user).RawQueryOne("SELECT * FROM users WHERE id = ?", reiner.V{10})
```

#### Single Value

```go
err := db.Bind(&password).RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", reiner.V{10})
```

#### Single Value From Multiple Rows

```go
err := db.Bind(&usernames).RawQueryValue("SELECT username FROM users LIMIT 10")
```

#### Advanced

```go
params := reiner.V{1, "admin"}
err := db.Bind(&users).RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", params)

// will handle any SQL query.
params = reiner.V{10, 1, 10, 11, 2, 10}
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

### Conditions

#### Equals

```go
db.Where("id", 1)
db.Where("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 AND username='admin';
```

##### Having

```go
db.Where("id", 1)
db.Having("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 HAVING username='admin';
```

##### Columns Comparison

```go
// WRONG
db.Where("lastLogin", "createdAt")
// CORRECT
db.Where("lastLogin = createdAt")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE lastLogin = createdAt;
```

#### Custom

```go
db.Where("id", 50, ">=").Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE id >= 50;
```


```go
type User struct {
    Username string
    Password string
    Birthday string
    Nickname string
}

u := User{}

db, err := reiner.New('yamiodymel:yamiodymel@/test?charset=utf8')

id := db.Insert('users', reiner.H{
    "username": u.Username,
    "password": u.Password,
    "birthday": u.Birthday,
    "nickname": u.Nickname,
})

u = User{}

if err := db.Where("username", User.username).GetOne("users").Scan(&u); err != nil {
    // xxxxx
}
if u == nil {

}
```

```php
$db->join("users u", "p.tenantID=u.tenantID", "LEFT");
$db->joinWhere("users u", "u.tenantID", 5);
$products = $db->get ("products p", null, "u.name, p.productName");
```

```go
func main() {
	err := db.
		Join("users u", "p.tenantID = u.tenantID", "LEFT").
		JoinWhere("users u", "u.tenantID", 5).
		Get("products p", 30, "u.name, p.productName").
		Scan(&u)
	if err != nil {
		panic(err)
	}
	if db.Count == 0 {
		panic("No rows")
	}
}
```
