package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var db *DB

func TestMain(t *testing.T) {
	assert := assert.New(t)

	var err error
	db, err = New("root:root@/test?charset=utf8")

	assert.NoError(err)
}

func TestInsert(t *testing.T) {
	//assert := assert.New(t)
	//
	//err := db.Insert("users", map[string]string{
	//	"username": "YamiOdymel",
	//	"password": "test12345",
	//})
	//
	//assert.NoError(err)
	//assert.Equal("INSERT username, password INTO `users` VALUES (?, ?)", //db.LastQuery)
}

func TestOnDuplicate(t *testing.T) {

}

func TestInsertMulti(t *testing.T) {

}

func TestWhere(t *testing.T) {

}

func TestOrWhere(t *testing.T) {

}

func TestUpdate(t *testing.T) {

}

func TestLimit(t *testing.T) {

}

func TestBind(t *testing.T) {

}

func TestGet(t *testing.T) {

}

func TestGetOne(t *testing.T) {

}

func TestGetValue(t *testing.T) {

}

func TestPaginate(t *testing.T) {

}

func TestRawQuery(t *testing.T) {

}

func TestRawQueryOne(t *testing.T) {

}

func TestRawQueryValue(t *testing.T) {

}

func TestHaving(t *testing.T) {

}

func TestOrHaving(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

func TestOrderBy(t *testing.T) {

}

func TestGroupBy(t *testing.T) {

}

func TestJoin(t *testing.T) {

}

func TestJoinWhere(t *testing.T) {

}

func TestJoinOrWhere(t *testing.T) {

}

func TestSubQuery(t *testing.T) {

}

func TestHas(t *testing.T) {

}

func TestNow(t *testing.T) {

}

func TestFunc(t *testing.T) {

}

func TestDisconnect(t *testing.T) {

}

func TestPing(t *testing.T) {

}

func TestConnect(t *testing.T) {

}

func TestBegin(t *testing.T) {

}

func TestRollback(t *testing.T) {

}

func TestCommit(t *testing.T) {

}

func TestSetLockMethod(t *testing.T) {

}

func TestLock(t *testing.T) {

}

func TestUnlock(t *testing.T) {

}

func TestSetQueryOption(t *testing.T) {

}
