# Reiner

一個由 [Golang](https://golang.org/) 撰寫且比起部分 [ORM](https://zh.wikipedia.org/wiki/%E5%AF%B9%E8%B1%A1%E5%85%B3%E7%B3%BB%E6%98%A0%E5%B0%84) 還要讚的 [MySQL](https://www.mysql.com/) 指令包覆函式庫。彈性高、不需要建構體標籤。實際上，這就只是 [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) 不過是用在 [Golang](https://golang.org/) 而已（但還是多了些功能）。

#  這是什麼？

萊納是一個由 [Golang](https://golang.org/) 撰寫的 [MySQL](https://www.mysql.com/) 的指令包覆函式庫（不是 [ORM](https://zh.wikipedia.org/wiki/%E5%AF%B9%E8%B1%A1%E5%85%B3%E7%B3%BB%E6%98%A0%E5%B0%84)，永遠也不會是），幾乎所有東西都能操控於你手中。類似自己撰寫資料庫指令但是更簡單，JOIN 表格也變得比以前更方便了。

* 幾乎全功能的函式庫。
* 支援 MySQL 複寫橫向擴展機制（區分讀／寫連線）。
* 容易理解與記住、且使用方式十分簡單。
* SQL 指令建構函式。
* 資料庫表格建構協助函式。
* 支援子指令（Sub Query）。
* 可手動操作的交易機制（Transaction）和回溯（Rollback）功能。
* 透過預置聲明（[Prepared Statement](https://en.wikipedia.org/wiki/Prepared_statement)），99.9% 避免 SQL 注入攻擊。

# 為什麼？

[Gorm](https://github.com/jinzhu/gorm) 已經是 [Golang](https://golang.org/) 裡的 [ORM](https://zh.wikipedia.org/wiki/%E5%AF%B9%E8%B1%A1%E5%85%B3%E7%B3%BB%E6%98%A0%E5%B0%84) 典範，但實際上要操作複雜與關聯性高的 SQL 指令時並不是很合適，而 Reiner 解決了這個問題。Reiner 也試圖不要和建構體扯上關係，不希望使用者需要手動指定任何標籤在建構體中。

# 執行緒與併發安全性？

我們都知道 [Golang](https://golang.org/) 的目標就是併發程式，當共用同個資料庫的時候請透過 `Copy()` 函式複製一份新的包覆函式庫，這能避免函式遭受干擾或覆寫。此方式並不會使資料庫連線遞增而造成效能問題，因此你可以有好幾個併發程式且有好幾個包覆函式庫的複製體都不會出現效能問題。

# 索引

* [安裝方式](#安裝方式)
* [命名建議](#命名建議)
* [使用方式](#使用方式)
    * [資料庫連線](#資料庫連線)
    	* [水平擴展（讀／寫分離）](#水平擴展讀寫分離)
		* [SQL 建構模式](#sql-建構模式)
	* [資料綁定與處理](#資料綁定與處理)
		* [逐行掃描](#逐行掃描)
	* [插入](#插入)
		* [覆蓋](#覆蓋)
		* [函式](#函式)
		* [當重複時](#當重複時)
		* [多筆資料](#多筆資料)
			* [省略重複鍵名](#省略重複鍵名)
	* [更新](#更新)
		* [筆數限制](#筆數限制)
	* [選擇與取得](#選擇與取得)
		* [筆數限制](#筆數限制-1)
		* [指定欄位](#指定欄位)
		* [單行資料](#單行資料)
		* [取得單值](#取得單值)
		* [分頁功能](#分頁功能)
	* [執行生指令](#執行生指令)
		* [單行資料](#單行資料-1)
		* [取得單值](#取得單值-1)
		* [單值多行](#單值多行)
		* [進階方式](#進階方式)
	* [條件宣告](#條件宣告)
		* [擁有](#擁有)
		* [欄位比較](#欄位比較)
		* [自訂運算子](#自訂運算子)
		* [介於／不介於](#介於不介於)
		* [於清單／不於清單內](#於清單不於清單內)
		* [或／還有或](#或還有或)
		* [空值](#空值)
		* [時間戳](#時間戳)
			* [相對](#相對)
			* [日期](#日期)
			* [時間](#時間)
		* [生條件](#生條件)
			* [條件變數](#條件變數)
	* [刪除](#刪除)
	* [排序](#排序)
		* [從值排序](#從值排序)
	* [群組](#群組)
	* [加入](#加入)
		* [條件限制](#條件限制)
	* [子指令](#子指令)
		* [選擇／取得](#選擇取得)
		* [插入](#插入-1)
		* [加入](#加入-1)
		* [存在／不存在](#存在不存在)
	* [是否擁有該筆資料](#是否擁有該筆資料)
	* [輔助函式](#輔助函式)
		* [資料庫連線](#資料庫連線)
		* [最後執行的 SQL 指令](#最後執行的-sql-指令)
		* [結果／影響的行數](#結果影響的行數)
		* [最後插入的編號](#最後插入的編號)
	* [交易函式](#交易函式)
	* [鎖定表格](#鎖定表格)
	* [指令關鍵字](#指令關鍵字)
		* [多個選項](#多個選項)
* [表格建構函式](#表格建構函式)

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/TeaMeow/Reiner
```

# 命名建議

在 Reiner 中為了配合 [Golang](https://golang.org/) 程式命名規範，我們建議你將所有事情以[駝峰式大小寫](https://zh.wikipedia.org/zh-tw/%E9%A7%9D%E5%B3%B0%E5%BC%8F%E5%A4%A7%E5%B0%8F%E5%AF%AB)命名，因為這能夠確保兩邊的風格相同。事實上，甚至連資料庫內的表格名稱、欄位名稱都該這麼做。當遇上 `ip`、`id`、`url` 時，請遵循 Golang 的命名方式皆以大寫使用，如 `AddrIP`、`UserID`、`PhotoURL`，而不是 `AddrIp`、`UserId`、`PhotoUrl`。

# 使用方式

Reiner 的使用方式十分直覺與簡易，類似基本的 SQL 指令集但是更加地簡化了。

## 資料庫連線

首先你需要透過函式來將 Reiner 連上資料庫，如此一來才能夠初始化包覆函式庫與相關的資料庫表格建構函式。一個最基本的單資料庫連線，讀寫都將透過此連線，連線字串共用於其它套件是基於 DSN（[Data Source Name](https://en.wikipedia.org/wiki/Data_source_name)）。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### 水平擴展（讀／寫分離）

這種方式可以有好幾個主要資料庫、副從資料庫，這意味著寫入時都會流向到主要資料庫，而讀取時都會向副從資料庫請求。這很適合用在大型結構還有水平擴展上。當你有多個資料庫來源時，Reiner 會逐一遞詢每個資料庫來源，英文稱其為 [Round Robin](https://zh.wikipedia.org/zh-tw/%E5%BE%AA%E7%92%B0%E5%88%B6)，也就是每個資料庫都會輪流呼叫而避免單個資料庫負荷過重，也不會有隨機呼叫的事情發生。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/master?charset=utf8", []string{
	"root:root@/slave?charset=utf8",
	"root:root@/slave2?charset=utf8",
	"root:root@/slave3?charset=utf8",
})
if err != nil {
    panic(err)
}
```

### SQL 建構模式

如果你已經有喜好的 SQL 資料庫處理套件，那麼你就可以在建立 Reiner 時不要傳入任何資料，這會使 Reiner 避免與資料庫互動，透過這個設計你可以將 Reiner 作為你的 SQL 指令建構函式。

```go
db, _ := reiner.New()
db.Table("Users").Where("Username", "YamiOdymel").Get()

// 然後像這樣透過原生的 `database/sql` 執行指令。
sql.Prepare(db.LastQuery)
sql.Exec("YamiOdymel")
```

## 資料綁定與處理

Reiner 允許你將結果與結構體切片或結構體綁定在一起。

```go
var user []*User
err := db.Bind(&user).Get("Users")
```

### 逐行掃描

如果你偏好傳統的 `rows.Next` 來對每筆資料進行逐行掃描，Reiner 亦提供了 `Scan` 方式允許你傳入自訂的資料處理函式。你能夠在網路上找到ㄧ些輔助 `*sql.Rows` 的函式。

```go
err := db.Scan(func(rows *sql.Rows) {
	var username, password string
	rows.Scan(&username, &password)
}).Table("Users").Get()
```

或者你不想要透過 Reiner 的 `Scan` 方式，你可以透過 `LastRows` 直接取得最後一次的 `*sql.Rows`。

```go
err := db.Table("Users").Get()
rows := db.LastRows
for rows.Next() {
	var username, password string
	rows.Scan(&username, &password)
}
```

## 插入

透過 Reiner 你可以很輕鬆地透過建構體或是 map 來插入一筆資料。這是最傳統的插入方式，若該表格有自動遞增的編號欄位，插入後你就能透過 `LastInsertID` 獲得最後一次插入的編號。

```go
err := db.Table("Users").Insert(map[string]string{
	"Username": "YamiOdymel",
	"Password": "test",
})
// 等效於：INSERT INTO Users (Username, Password) VALUES (?, ?)
```

### 覆蓋

覆蓋的用法與插入相同，當有同筆資料時會先進行刪除，然後再插入一筆新的，這對有外鍵的表格來說十分危險。

```go
err := db.Table("Users").Replace(map[string]string{
	"Username": "YamiOdymel",
	"Password": "test",
})
// 等效於：REPLACE INTO Users (Username, Password) VALUES (?, ?)
```

### 函式

插入時你可以透過 Reiner 提供的函式來執行像是 `SHA1()` 或者取得目前時間的 `NOW()`，甚至將目前時間加上一年⋯等。

```go
err := db.Table("Users").Insert(map[string]interface{}{
	"Username":  "YamiOdymel",
	"Password":  db.Func("SHA1(?)", "secretpassword+salt"),
	"Expires":   db.Now("+1Y"),
	"CreatedAt": db.Now(),
})
// 等效於：INSERT INTO Users (Username, Password, Expires, CreatedAt) VALUES (?, SHA1(?), NOW() + INTERVAL 1 YEAR, NOW())
```

### 當重複時

Reiner 支援了插入資料若重複時可以更新該筆資料的指定欄位，這類似「覆蓋」，但這並不會先刪除原先的資料，這種方式僅會在插入時檢查是否重複，若重複則更新該筆資料。

```go
lastInsertID := "ID"
err := db.Table("Users").OnDuplicate([]string{"UpdatedAt"}, lastInsertID).Insert(map[string]interface{}{
	"Username":  "YamiOdymel",
	"Password":  "test",
	"UpdatedAt": db.Now(),
})
// 等效於：INSERT INTO Users (Username, Password, UpdatedAt) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE UpdatedAt = VALUE(UpdatedAt)
```

### 多筆資料

Reiner 允許你透過 `InsertMulti` 同時間插入多筆資料（單指令插入多筆資料），這省去了透過迴圈不斷執行單筆插入的困擾，這種方式亦大幅度提升了效能。

```go
data := []map[string]string{
	{
		"Username": "YamiOdymel",
		"Password": "test",
	}, {
		"Username": "Karisu",
		"Password": "12345",
	},
}
db.Table("Users").InsertMulti(data)
// 等效於：INSERT INTO Users (Username, Password) VALUES (?, ?), (?, ?)
```

#### 省略重複鍵名

當所有資料都有一樣的插入欄位名稱時，鍵名可以只需要輸入一次。

```go
values := [][]interface{}{
	{"YamiOdymel", "test"},
	{"Karisu", "12345"},
}
columns := []string{"username", "password"}
db.Table("Users").InsertMulti(values, columns)
// 等效於：INSERT INTO Users (Username, Password) VALUES (?, ?), (?, ?)
```

## 更新

更新一筆資料在 Reiner 中極為簡單，你只需要指定表格名稱還有資料即可。

```go
db.Table("Users").Where("Username", "YamiOdymel").Update(map[string]string{
	"Username": "Karisu",
	"Password": "123456",
})
// 等效於：UPDATE Users SET Username = ?, Password = ? WHERE Username = ?
```

### 筆數限制

`Limit` 能夠限制更新的筆數，如果是 10，那就表示只更新最前面 10 筆資料而非全部。

```go
db.Table("Users").Limit(10).Update(data)
// 等效於：UPDATE Users SET ... LIMIT 10
```

## 選擇與取得

最基本的選擇在 Reiner 中稱之為 `Get` 而不是 `Select`。如果你想要取得 `rows.Next` 來掃描每一行的結果，Reiner 提供了 `LastRows` 即為最後一次的 `*sql.rows` 資料。

```go
err := db.Table("Users").Get()
// 等效於：SELECT * FROM Users
```

### 筆數限制

`Limit` 能夠限制取得的筆數，如果是 10，那就表示只取得最前面 10 筆資料而非全部。

```go
db.Table("Users").Limit(10).Get()
// 等效於：SELECT * FROM Users LIMIT 10
```

### 指定欄位

在 `Get` 中傳遞欄位名稱作為參數，多個欄位由逗點區分，亦能是函式。

```go
db.Table("Users").Get("Username", "Nickname")
// 等效於：SELECT Username, Nickname FROM Users

db.Table("Users").Get("COUNT(*) AS Count")
// 等效於：SELECT COUNT(*) AS Count FROM Users
```

### 單行資料

預設來說 `Get` 會回傳一個切片或是陣列，這令你需要透過迴圈逐一取得資料，但某些情況下你很確信你僅要取得一筆資料的話，可以嘗試 `GetOne`。這能將資料直接映射到單個建構體上而避免你需要透過迴圈處理資料的麻煩。

```go
db.Table("Users").Where("ID", 1).GetOne()
// 等效於：SELECT * FROM Users WHERE ID = ?

db.Table("Users").GetOne("SUM(ID)", "COUNT(*) AS Count")
// 等效於：SELECT SUM(ID), COUNT(*) AS Count FROM Users
```

### 取得單值

這就像 `GetOne`，但 `GetValue` 取得的是單個欄位的內容，例如說你想要單個使用者的暱稱，甚至是多個使用者的暱稱陣列就很適用。

```go
db.Table("Users").GetValue("Username")
// 也能搭配 Limit。
db.Table("Users").Limit(5).GetValue("Username")
// 或者是函式。
db.Table("Users").GetValue("COUNT(*)")
```

### 分頁功能

分頁就像是取得資料ㄧ樣，但更擅長用於多筆資料、不會一次顯示完畢的內容。Reiner 能夠幫你自動處理換頁功能，讓你不需要自行計算換頁時的筆數應該從何開始。為此，你需要定義兩個變數，一個是目前的頁數，另一個是單頁能有幾筆資料。

```go
page := 2
db.PageLimit = 20
db.Table("Users").Paginate(page)
// 等效於：SELECT * FROM Users LIMIT 20, 20

fmt.Println("目前頁數為 %d，共有 %d 頁", page, db.TotalPages)
```

## 執行生指令

Reiner 已經提供了近乎日常中 80% 會用到的方式，但如果好死不死你想使用的功能在那 20% 之中，我們還提供了原生的方法能讓你直接輸入 SQL 指令執行自己想要的鳥東西。一個最基本的生指令（Raw Query）就像這樣。

其中亦能帶有預置聲明（Prepared Statement），也就是指令中的問號符號替代了原本的值。這能避免你的 SQL 指令遭受注入攻擊。

```go
db.RawQuery("SELECT * FROM Users WHERE ID >= ?", 10)
```

### 單行資料

僅選擇單筆資料的生指令函式，這意味著你能夠將取得的資料直接映射到一個建構體上。

```go
db.RawQueryOne("SELECT * FROM Users WHERE ID = ?", 10)
```

### 取得單值

透過 `RawQueryValue` 可以直接取得單個欄位得值，而不是一個陣列或切片。

```go
db.RawQueryValue("SELECT Password FROM Users WHERE ID = ? LIMIT 1", 10)
```

### 單值多行

透過 `RawQueryValue` 能夠取得單一欄位的值，當有多筆結果的時候會取得一個值陣列。

```go
db.RawQueryValue("SELECT Username FROM Users LIMIT 10")
```

### 進階方式

如果你對 SQL 指令夠熟悉，你也可以使用更進階且複雜的用法。

```go
db.RawQuery("SELECT ID, FirstName, LastName FROM Users WHERE ID = ? AND Username = ?", 1, "admin")

params := []int{10, 1, 10, 11, 2, 10}
query := (`
    SELECT A FROM TestTable
        WHERE A = ? AND B = ?
        ORDER BY A LIMIT ?
) UNION (
    SELECT A FROM TestTable2
        WHERE A = ? AND B = ?
        ORDER BY A LIMIT ?
`)
db.RawQuery(query, params...)
```

## 條件宣告

透過 Reiner 宣告 `WHERE` 條件也能夠很輕鬆。一個最基本的 `WHERE AND` 像這樣使用。

```go
db.Table("Users").Where("ID", 1).Where("Username", "admin").Get()
// 等效於：SELECT * FROM Users WHERE ID = ? AND Username = ?
```

### 擁有

`HAVING` 能夠與 `WHERE` 一同使用。

```go
db.Table("Users").Where("ID", 1).Having("Username", "admin").Get()
// 等效於：SELECT * FROM Users WHERE ID = ? HAVING Username = ?
```

### 欄位比較

如果你想要在條件中宣告某個欄位是否等於某個欄位⋯你能夠像這樣。

```go
// 別這樣。
db.Table("Users").Where("LastLogin", "CreatedAt").Get()
// 這樣才對。
db.Table("Users").Where("LastLogin = CreatedAt").Get()
// 等效於：SELECT * FROM Users WHERE LastLogin = CreatedAt
```

### 自訂運算子

在 `Where` 或 `Having` 中，你可以自訂條件的運算子，如 >=、<=、<>⋯等。

```go
db.Table("Users").Where("ID", ">=", 50).Get()
// 等效於：SELECT * FROM Users WHERE ID >= ?
```

### 介於／不介於

條件也可以用來限制數值內容是否在某數之間（相反之，也能夠限制是否不在某範圍內）。

```go
db.Table("Users").Where("ID", "BETWEEN", 0, 20).Get()
// 等效於：SELECT * FROM Users WHERE ID BETWEEN ? AND ?
```

### 於清單／不於清單內

條件能夠限制並確保取得的內容不在（或者在）指定清單內。

```go
db.Table("Users").WhereIn("ID", "IN", 1, 5, 27, -1, "d").Get()
// 等效於：SELECT * FROM Users WHERE ID IN (?, ?, ?, ?, ?)

list := []interface{}{1, 5, 27, -1, "d"}
db.Table("Users").WhereIn("ID", "IN", list...).Get()
// 等效於：SELECT * FROM Users WHERE ID IN (?, ?, ?, ?, ?)
```

### 或／還有或

通常來說多個 `Where` 會產生 `AND` 條件，這意味著所有條件都必須符合，有些時候你只希望符合部分條件即可，就能夠用上 `OrWhere`。

```go
db.Table("Users").Where("FirstNamte", "John").OrWhere("FirstNamte", "Peter").Get()
// 等效於：SELECT * FROM Users WHERE FirstName = ? OR FirstName = ?
```

如果你的要求比較多，希望達到「A = B 或者 (A = C 或 A = D)」的話，你可以嘗試這樣。

```go
db.Table("Users").Where("A = B").OrWhere("(A = C OR A = D)").Get()
// 等效於：SELECT * FROM Users WHERE A = B OR (A = C OR A = D)
```

### 空值

確定某個欄位是否為空值。

```go
// 別這樣。
db.Table("Users").Where("LastName", "NULL").Get()
// 這樣才對。
db.Table("Users").Where("LastName", "IS", nil).Get()
// 等效於：SELECT * FROM Users WHERE LastName IS NULL
```

### 時間戳

[Unix Timestamp](https://en.wikipedia.org/wiki/Unix_time) 是一項將日期與時間秒數換算成數字的格式（範例：`1498001308`），這令你能夠輕易地換算其秒數，但當你要判斷時間是否為某一年、月、日，甚至範圍的時候就會有些許困難，而 Reiner 也替你想到了這一點。

需要注意的是 Reiner 中的 `Timestamp` 工具無法串聯使用，這意味著當你想要確認時間戳是否為某年某月時，你需要有兩個 `Where` 條件，而不行使用 `IsYear().IsMonth()`。更多的用法可以在原生文件中找到，這裡僅列出不完全的範例供大略參考。

#### 日期

判斷是否為特定年、月、日、星期或完整日期。

```go
t := db.Timestamp

db.Table("Users").Where("CreatedAt", t.IsDate("2017-07-13")).Get()
// 等效於：SELECT * FROM Users WHERE DATE(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsYear(2017)).Get()
// 等效於：SELECT * FROM Users WHERE YEAR(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsMonth(1)).Get()
db.Table("Users").Where("CreatedAt", t.IsMonth("January")).Get()
// 等效於：SELECT * FROM Users WHERE MONTH(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsDay(16)).Get()
// 等效於：SELECT * FROM Users WHERE DAY(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsWeekday(5)).Get()
db.Table("Users").Where("CreatedAt", t.IsWeekday("Friday")).Get()
// 等效於：SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?
```

#### 時間

確定是否為特定時間。

```go
t := db.Timestamp()

db.Table("Users").Where("CreatedAt", t.IsHour(18)).Get()
// 等效於：SELECT * FROM Users WHERE HOUR(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsMinute(25)).Get()
// 等效於：SELECT * FROM Users WHERE MINUTE(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsSecond(16)).Get()
// 等效於：SELECT * FROM Users WHERE SECOND(FROM_UNIXTIME(CreatedAt)) = ?

db.Table("Users").Where("CreatedAt", t.IsWeekday(5)).Get()
// 等效於：SELECT * FROM Users WHERE WEEKDAY(FROM_UNIXTIME(CreatedAt)) = ?
```

### 生條件

你也能夠直接在條件中輸入指令。

```go
db.Table("Users").Where("ID != CompanyID").Where("DATE(CreatedAt) = DATE(LastLogin)").Get()
// 等效於：SELECT * FROM Users WHERE ID != CompanyID AND DATE(CreatedAt) = DATE(LastLogin)
```

#### 條件變數

生條件中可以透過 `?` 符號，並且在後面傳入自訂變數。

```go
db.Table("Users").Where("(ID = ? OR ID = ?)", 6, 2).Where("Login", "Mike").Get()
// 等效於：SELECT * FROM Users WHERE (ID = ? OR ID = ?) AND Login = ?
```

## 刪除

刪除一筆資料再簡單不過了，透過 `Count` 計數能夠清楚知道你的 SQL 指令影響了幾行資料，如果是零的話即是無刪除任何資料。

```go
err := db.Table("Users").Where("ID", 1).Delete()
if err == nil && db.Count != 0 {
    fmt.Printf("成功地刪除了 %d 筆資料！", db.Count)
}
// 等效於：DELETE FROM Users WHERE ID = ?
```

## 排序

Reiner 亦支援排序功能，如遞增或遞減，亦能擺放函式。

```go
db.Table("Users").OrderBy("ID", "ASC").OrderBy("Login", "DESC").OrderBy("RAND()").Get()
// 等效於：SELECT * FROM Users ORDER BY ID ASC, Login DESC, RAND()
```

### 從值排序

也能夠從值進行排序，只需要傳入一個切片即可。

```go
db.Table("Users").OrderBy("UserGroup", "ASC", "SuperUser", "Admin", "Users").Get()
// 等效於：SELECT * FROM Users ORDER BY FIELD (UserGroup, ?, ?, ?) ASC
```

## 群組

簡單的透過 `GroupBy` 就能夠將資料由指定欄位群組排序。

```go
db.Table("Users").GroupBy("Name").Get()
// 等效於：SELECT * FROM Users GROUP BY Name
```

## 加入

Reiner 支援多種表格加入方式，如：`InnerJoin`、`LeftJoin`、`RightJoin`、`NaturalJoin`、`CrossJoin`。

```go
db.Table("Products")
db.LeftJoin("Users", "Products.TenantID = Users.TenantID")
db.Where("Users.ID", 6)
db.Get("Users.Name", "Products.ProductName")
// 等效於：SELECT Users.Name, Products.ProductName FROM Products AS Products LEFT JOIN Users AS Users ON (Products.TenantID = Users.TenantID) WHERE Users.ID = ?
```

### 條件限制

你亦能透過 `JoinWhere` 或 `JoinOrWhere` 擴展表格加入的限制條件。

```go
db.Table("Products")
db.LeftJoin("Users", "Products.TenantID = Users.TenantID")
db.JoinOrWhere("Users", "Users.TenantID", 5)
db.Get("Users.Name", "Products.ProductName")
// 等效於：SELECT Users.Name, Products.ProductName FROM Products AS Products LEFT JOIN Users AS Users ON (Products.TenantID = Users.TenantID OR Users.TenantID = ?)
```

## 子指令

Reiner 支援複雜的子指令，欲要建立一個子指令請透過 `SubQuery` 函式，這將會建立一個不能被執行的資料庫包覆函式庫，令你可以透過 `Get`、`Update` 等建立相關 SQL 指令，但不會被資料庫執行。將其帶入到一個正常的資料庫函式中即可成為子指令。

```go
subQuery := db.SubQuery()
subQuery.Table("Users").Get()
// 等效於不會被執行的：SELECT * FROM Users
```

### 選擇／取得

你能夠輕易地將子指令放置在選擇／取得指令中。

```go
subQuery := db.SubQuery()
subQuery.Table("Products").Where("Quantity", ">", 2).Get("UserID")

db.Table("Users").Where("ID", "IN", subQuery).Get()
// 等效於：SELECT * FROM Users WHERE ID IN (SELECT UserID FROM Products WHERE Quantity > ?)
```

### 插入

插入新資料時也可以使用子指令。

```go
subQuery := db.SubQuery()
subQuery.Table("Users").Where("ID", 6).GetOne("Name")

db.Table("Products").Insert(map[string]interface{}{
	"ProductName": "測試商品",
	"UserID":      subQuery,
	"LastUpdated": db.Now(),
})
// 等效於：INSERT INTO Products (ProductName, UserID, LastUpdated) VALUES (?, (SELECT Name FROM Users WHERE ID = 6), NOW())
```

### 加入

就算是加入表格的時候也可以用上子指令，但你需要為子指令建立別名。

```go
subQuery := db.SubQuery("Users")
subQuery.Table("Users").Where("Active", 1).Get()

db.Table("Products")
db.LeftJoin(subQuery, "Products.UserID = U.ID")
db.Get("U.Username", "Products.ProductName")
// 等效於：SELECT Users.Username, Products.ProductName FROM Products AS Products LEFT JOIN (SELECT * FROM Users WHERE Active = ?) AS Users ON Products.UserID = Users.ID
```

### 存在／不存在

你同時也能夠透過子指令來確定某筆資料是否存在。

```go
subQuery := db.SubQuery()
subQuery.Table("Users").Where("Company", "測試公司").Get("UserID")

db.Table("Products").Where(subQuery, "EXISTS").Get()
// 等效於：SELECT * FROM Products WHERE EXISTS (SELECT UserID FROM Users WHERE Company = ?)
```

## 是否擁有該筆資料

有些時候我們只想知道資料庫是否有符合的資料，但並不是要取得其資料，舉例來說就像是登入是僅是要確認帳號密碼是否吻合，此時就可以透過 `Has` 用來確定資料庫是否有這筆資料。

```go
has, err := db.Table("Users").Where("Username", "yamiodymel").Where("Password", "123456").Has()
if has {
	fmt.Println("登入成功！")
} else {
	fmt.Println("帳號或密碼錯誤。")
}
```

## 輔助函式

Reiner 有提供一些輔助用的函式協助你除錯、紀錄，或者更加地得心應手。

### 資料庫連線

透過 Disconnect 結束一段連線。

```go
db.Disconnect()
```

你也能在資料庫發生錯誤、連線遺失時透過 `Connect` 來重新手動連線。

```go
if err := db.Ping(); err != nil {
	db.Connect()
}
```

### 最後執行的 SQL 指令

取得最後一次所執行的 SQL 指令，這能夠用來記錄你所執行的所有動作。

```go
db.Table("Users").Get()
fmt.Println("最後一次執行的 SQL 指令是：%s", db.LastQuery)
```

### 結果／影響的行數

行數很常用於檢查是否有資料、作出變更。資料庫不會因為沒有變更任何資料而回傳一個錯誤（資料庫僅會在真正發生錯誤時回傳錯誤資料），所以這是很好的檢查方法。

```go
db.Table("Users").Get()
fmt.Println("總共獲取 %s 筆資料", db.Count)
db.Table("Users").Delete()
fmt.Println("總共刪除 %s 筆資料", db.Count)
db.Table("Users").Update(data)
fmt.Println("總共更新 %s 筆資料", db.Count)
```

### 最後插入的編號

當插入一筆新的資料，而該表格帶有自動遞增的欄位時，就能透過 `LastInsertID` 取得最新一筆資料的編號。

```go
db.Table("Users").Insert(data)
id := db.LastInsertID
```

如果你是同時間插入多筆資料，你仍可以透過 `LastInsertIDs` 取得剛才插入的所有資料編號。

```go
db.Table("Users").InsertMulti(data)
ids := db.LastInsertIDs
```

## 交易函式

交易函式僅限於 [InnoDB](https://zh.wikipedia.org/zh-tw/InnoDB) 型態的資料表格，這能令你的資料寫入更加安全。你可以透過 `Begin` 開始記錄並繼續你的資料庫寫入行為，如果途中發生錯誤，你能透過 `Rollback` 回到紀錄之前的狀態，即為回溯（或滾回、退回），如果這筆交易已經沒有問題了，透過 `Commit` 將這次的變更永久地儲存到資料庫中。

```go
err := db.Table("Wallets").Begin().Insert(data)
if err != nil {
	db.Rollback()
} else {
	db.Commit()
}
```

## 鎖定表格

你能夠手動鎖定資料表格，避免同時間寫入相同資料而發生錯誤。

```go
db.Table("Users").SetLockMethod("WRITE").Lock()

// 呼叫其他的 Lock() 函式也會自動將前一個上鎖解鎖，當然你也可以手動呼叫 Unlock() 解鎖。
db.Unlock()

// 同時間要鎖上兩個表格也很簡單。
db.Table("Users", "Logs").SetLockMethod("READ").Lock()
```

## 指令關鍵字

Reiner 也支援設置指令關鍵字。

```go
db.Table("Users").SetQueryOption("LOW_PRIORITY").Insert(data)
// 等效於：INSERT LOW_PRIORITY INTO Users ...

db.Table("Users").SetQueryOption("FOR UPDATE").Get()
// 等效於：SELECT * FROM Users FOR UPDATE

db.Table("Users").SetQueryOption("SQL_NO_CACHE").Get()
// 等效於：SELECT SQL_NO_CACHE * FROM Users
```

### 多個選項

你亦能同時設置多個關鍵字給同個指令。

```go
db.Table("Users").SetQueryOption("LOW_PRIORITY", "IGNORE").Insert(data)
// Gives: INSERT LOW_PRIORITY IGNORE INTO Users ...
```

# 表格建構函式

Reiner 除了基本的資料庫函式可供使用外，還能夠建立一個表格並且規劃其索引、外鍵、型態。

```go
migration := db.Migration()

migration.Table("Users").Column("Username").Varchar(32).Primary().Create()
// 等效於：CREATE TABLE Users (Username VARCHAR(32) NOT NULL PRIMARY KEY) ENGINE=INNODB
```


| 數值       | 字串       | 二進制     | 檔案資料     | 時間      | 浮點數     | 固組   |
|-----------|------------|-----------|------------|-----------|-----------|-------|
| TinyInt   | Char       | Binary    | Blob       | Date      | Double    | Enum  |
| SmallInt  | Varchar    | VarBinary | MediumBlob | DateTime  | Decimal   | Set   |
| MediumInt | TinyText   | Bit       | LongBlob   | Time      | Float     |       |
| Int       | Text       |           |            | Timestamp |           |       |
| BigInt    | MediumText |           |            | Year      |           |       |
|           | LongText   |           |            |           |           |       |

# 相關連結

這裡是 Reiner 受啟發，或是和資料庫有所關聯的連結。

* [kisielk/sqlstruct](http://godoc.org/github.com/kisielk/sqlstruct)
* [jmoiron/sqlx](https://github.com/jmoiron/sqlx)
* [russross/meddler](https://github.com/russross/meddler)
* [jinzhu/gorm](https://github.com/jinzhu/gorm)