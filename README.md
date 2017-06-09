# Reiner

The Golang database ORM with 1990s style.

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
err := db.Get("users", reiner.O{
	Scan: &u,
})
// or with the limit.
err := db.Limit(10).Get("users", reiner.O{
	Scan: &u,
})
```

#### Specified Columns

```go
err := db.Get("users", reiner.O{
	Scan:    &users,
	Columns: {"username", "nickname"},
})
// count := db.Count
```

#### Single Row

```go
err := db.Where("id", 1).GetOne("users", reiner.O{
	Scan: &user,
})
// or with the custom query.
err := db.GetOne("users", reiner.O{
	Query: "sum(id), count(*) as cnt",
	Scan:  &stats,
})
```

#### Get Value

```go
err := db.GetValue("users", "username", reiner.O{
	Scan: &usernames,
})
// or with the limit.
err := db.Limit(5).GetValue("users", "username", reiner.O{
	Scan: &usernames,
})
// or with the function.
err := db.GetValue("users", "count(*)", reiner.O{
	Scan: &total,
})
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
