package main

import (
	"fmt"

	"github.com/teacat/reiner"
)

func main() {
	db, _ := reiner.New()
	// 當交易開始時請使用回傳的 `tx` 而不是原先的 `db`，這樣才能確保交易繼續。
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	// 如果插入資料時發生錯誤，則呼叫 `Rollback()` 回到交易剛開始的時候。
	if _, err = tx.Table("Wallets").Insert(data); err != nil {
		tx.Rollback()
		panic(err)
	}
	if _, err = tx.Table("Users").Insert(data); err != nil {
		tx.Rollback()
		panic(err)
	}

	// 透過 `Commit()` 確保上列變更都已經永久地儲存到資料庫。
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	fmt.Println(builder.LastQuery)
}
