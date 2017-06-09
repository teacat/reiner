# Reiner

### Conenction

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("yamiodymel:yamiodymel@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### Traditional Insertion

```go
id, err := db.Insert("users", reiner.H{
    "username": "YamiOdymel",
    "password": "test",
})
```

### Struct Insertion

```go
type User struct {
    Username, Password string
}

u := User{"YamiOdymel", "test"}
id, err := db.Insert("users", u)
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
