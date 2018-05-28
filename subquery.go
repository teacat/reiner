package reiner

// SubQuery 是單個子指令，任何的變更都會回傳一份複製子指令來避免多個 Goroutine 編輯同個子指令指標建構體。
type SubQuery struct {
	builder *Builder
	// PageLimit 限制了一頁僅能有幾筆資料。
	PageLimit int
}

// clone 會複製子指令來避免多個 Goroutine 編輯同個子指令指標建構體。
func (s *SubQuery) clone() (cloned *SubQuery) {
	a := *s
	a.builder = a.builder.clone()
	cloned = &a
	return
}

//=======================================================
// 輸出函式
//=======================================================

// Table 能夠指定資料表格的名稱。
func (s *SubQuery) Table(tableName ...string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.Table(tableName...)
	return
}

//=======================================================
// 選擇函式
//=======================================================

// Get 會取得多列的資料結果，傳入的參數為欲取得的欄位名稱，不傳入參數表示取得所有欄位。
func (s *SubQuery) Get(columns ...string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder, _ = subQuery.builder.Get(columns...)
	return
}

// Paginate 基本上和 `Get` 取得函式無異，但此函式能夠自動依照分頁數來推算該從哪裡繼續取得資料。
// 使用時須先確定是否有指定 `PageLimit`（預設為：20），這樣才能限制一頁有多少筆資料。
func (s *SubQuery) Paginate(pageCount int, columns ...string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder.PageLimit = subQuery.PageLimit
	subQuery.builder, _ = subQuery.builder.Paginate(pageCount, columns...)
	return
}

//=======================================================
// 限制函式
//=======================================================

// Limit 能夠在 SQL 查詢指令中建立限制筆數的條件。
func (s *SubQuery) Limit(from int, count ...int) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.Limit(from, count...)
	return
}

// OrderBy 會依照指定的欄位來替結果做出排序（例如：`DESC`、`ASC`）。
func (s *SubQuery) OrderBy(column string, args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.OrderBy(column, args...)
	return
}

// GroupBy 會在執行 SQL 指令時依照特定的欄位來做執行區分。
func (s *SubQuery) GroupBy(columns ...string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.GroupBy(columns...)
	return
}

//=======================================================
// 指令函式
//=======================================================

// RawQuery 會接收傳入的變數來執行傳入的 SQL 執行語句，變數可以在語句中以 `?`（Prepared Statements）使用來避免 SQL 注入攻擊。
// 這會將多筆資料映射到本地的建構體切片、陣列上。
func (s *SubQuery) RawQuery(query string, values ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder, _ = subQuery.builder.RawQuery(query, values...)
	return
}

//=======================================================
// 條件函式
//=======================================================

// Where 會增加一個 `WHERE AND` 條件式。
func (s *SubQuery) Where(args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.Where(args...)
	return
}

// OrWhere 會增加一個 `WHERE OR` 條件式。
func (s *SubQuery) OrWhere(args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.OrWhere(args...)
	return
}

// Having 會增加一個 `HAVING AND` 條件式。
func (s *SubQuery) Having(args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.Having(args...)
	return
}

// OrHaving 會增加一個 `HAVING OR` 條件式。
func (s *SubQuery) OrHaving(args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.OrHaving(args...)
	return
}

//=======================================================
// 加入函式
//=======================================================

// LeftJoin 會向左插入一個資料表格。
func (s *SubQuery) LeftJoin(table interface{}, condition string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.LeftJoin(table, condition)
	return
}

// RightJoin 會向右插入一個資料表格。
func (s *SubQuery) RightJoin(table interface{}, condition string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.RightJoin(table, condition)
	return
}

// InnerJoin 會內部插入一個資料表格。
func (s *SubQuery) InnerJoin(table interface{}, condition string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.InnerJoin(table, condition)
	return
}

// NaturalJoin 會自然插入一個資料表格。
func (s *SubQuery) NaturalJoin(table interface{}, condition string) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.NaturalJoin(table, condition)
	return
}

// JoinWhere 能夠建立一個基於 `WHERE AND` 的條件式給某個指定的插入資料表格。
func (s *SubQuery) JoinWhere(table interface{}, args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.JoinWhere(table, args...)
	return
}

// JoinOrWhere 能夠建立一個基於 `WHERE OR` 的條件式給某個指定的插入資料表格。
func (s *SubQuery) JoinOrWhere(table interface{}, args ...interface{}) (subQuery *SubQuery) {
	subQuery = s.clone()
	subQuery.builder = subQuery.builder.JoinOrWhere(table, args...)
	return
}
