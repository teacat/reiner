# Reiner

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

if err := db.Where("username", User.username).GetOne("users", &u); err != nil {
    // xxxxx
}
if u == nil {

}
```

```php
$db->join("users u", "p.tenantID=u.tenantID", "LEFT");
$db->joinWhere("users u", "u.tenantID", 5);
$products = $db->get ("products p", null, "u.name, p.productName");
``

```go
err := db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
  .JoinWhere("users u", "u.tenantID", 5)
  .Get("products p", renier.O{
    Scan: &u,
    Amount: 30,
    Fields: "u.name, p.productName",
})
if err != nil {
    panic(err)
}
```
