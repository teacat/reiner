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

if err := db.Where("username", User.username).GetOne("users").Scan(&u); err != nil {
    // xxxxx
}
if u == nil {

}
```