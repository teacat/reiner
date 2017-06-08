# Reiner

```go
type User struct {
    Username string
    Password string
    Birthday string
    Nickname string
}

db, err := reiner.New('yamiodymel:yamiodymel@/test?charset=utf8')

id := db.insert('users', reiner.H{
    "username": User.username,
    "password": User.password,
    "birthday": User.birthday,
    "nickname": User.nickname,
})

u = User{}

if err := db.where("username", User.username).getOne("users").scan(&u); err != nil {
    // xxxxx
}
if u == nil {

}
```