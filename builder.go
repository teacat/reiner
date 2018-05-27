package reiner

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
	// MySQL 驅動程式。
	_ "github.com/go-sql-driver/mysql"
)

var (
	// ErrInvalidPointer 是會在資料的映射目的地為 nil 指標時所發生的錯誤。
	ErrInvalidPointer = errors.New("reiner: the destination of the result is an invalid pointer")
	// ErrIncorrectDataType 是個會在插入、更新資料時傳入非 `map[string]interface` 資料型態參數時所發生的錯誤。
	ErrIncorrectDataType = errors.New("reiner: the data type must be a `map[string]interface`")
	// ErrUnbegunTransaction 會在執行尚未透過 `Begin` 初始化的交易時所發生的錯誤。
	ErrUnbegunTransaction = errors.New("reiner: calling the transaction function without `Begin()`")
	// ErrNoTable 是個會在未指定資料表格時所發生的錯誤。
	ErrNoTable = errors.New("reiner: no table was specified")
)

// Function 重現了一個像 `SHA(?)` 或 `NOW()` 的資料庫函式。
type Function struct {
	query  string
	values []interface{}
}

// condition 是一個 `WHERE` 或 `HAVING` 的條件式。
type condition struct {
	args      []interface{}
	connector string
}

// order 是個基於 `ORDER` 的排序資訊。
type order struct {
	column string
	args   []interface{}
}

// join 帶有資料表格的加入資訊。
type join struct {
	typ        string
	table      interface{}
	condition  string
	conditions []condition
}

// Trace 是個已執行的 SQL 指令蹤跡、堆疊資料。
type Trace struct {
	Query    string
	Duration time.Duration
	Stacks   []map[string]interface{}
	Error    error
}

// Builder 是個資料庫的 SQL 指令建置系統，同時也帶有資料庫的連線資料。
type Builder struct {
	db *DB
	// executable 表示是否該執行建置後的指令，當沒有連線的時候這會是 `false`。
	// 這表示僅用於建置 SQL 指令，而不是執行它。
	executable bool
	// alias 是作為子指令時所帶有的別名，這會用在子指令資料表格的加入上。
	alias string
	// destination 呈現了資料的映射目的地指針。
	destination        interface{}
	tableName          []string
	conditions         []condition
	havingConditions   []condition
	queryOptions       []string
	joins              map[string]*join
	onDuplicateColumns []string
	lastInsertIDColumn string
	limit              []int
	orders             []order
	groupBy            []string
	lockMethod         string
	tracing            bool
	query              string
	params             []interface{}
	count              int

	// Timestamp 是時間戳的相關輔助函式。
	Timestamp *Timestamp
	// Traces 是最後執行的 SQL 指令蹤跡資訊，適合用於效能除錯上。
	Traces []Trace
	// TotalCount 是結果的總計筆數。
	TotalCount int
	// PageLimit 限制了一頁僅能有幾筆資料。
	PageLimit int
	// TotalPage 是結果的總計頁數。
	TotalPage int
	// LasyQuery 是最後所執行的 SQL 指令。
	LastQuery string
	// LastInsertID 是最後所插入的資料 ID 編號。
	LastInsertID int
	// LastParams 是最後執行時的參數資料。
	LastParams []interface{}
	// LastResult 是最後執行時的 `sql.Result` 資料。
	LastResult sql.Result
}

// newBuilder 會基於傳入的資料庫連線來建立一個新的 SQL 指令建置系統。
func newBuilder(db *DB) *Builder {
	return &Builder{executable: true, db: db, Timestamp: &Timestamp{}, joins: make(map[string]*join)}
}

// clone 會複製資料庫建置函式，並決定是否一同複製現有的條件與設定。
func (b *Builder) clone() (cloned *Builder) {
	a := *b
	cloned = &a
	return
}

// cleanAfter 會在 SQL 指令建置之後清除資料來避免下次使用到舊的資料。
func (b *Builder) cleanAfter() {
	b.queryOptions = []string{}
	b.tableName = []string{}
	b.params = []interface{}{}
	b.onDuplicateColumns = []string{}
	b.groupBy = []string{}
	b.joins = map[string]*join{}
	b.orders = []order{}
	b.conditions = []condition{}
	b.havingConditions = []condition{}
	b.limit = []int{}
	b.destination = nil
}

// cleanBefore 會在 SQL 指令建置之前清除以往的資料，
// 留到建置前才清除的原因是避免需要歷史資料的時候卻在執行後直接被清除了。
func (b *Builder) cleanBefore() {
	b.TotalCount = 0
	b.TotalPage = 0
	b.LastInsertID = 0
	b.LastResult = nil
	b.LastParams = []interface{}{}
	b.count = 0
}

//=======================================================
// 保存函式
//=======================================================

// saveTrace 會取得、紀錄呼叫函式的名稱，並且計算執行時間然後將其保存於蹤跡資訊中。
func (b *Builder) saveTrace(err error, query string, startedAt time.Time) {
	if !b.tracing {
		return
	}
	var stacks []map[string]interface{}
	for skip := 0; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		stacks = append(stacks, map[string]interface{}{
			"Skip": skip,
			"PC":   pc,
			"File": file,
			"Line": line,
		})
	}
	b.Traces = append(b.Traces, Trace{
		Query:    query,
		Duration: time.Since(startedAt),
		Stacks:   stacks,
		Error:    err,
	})
}

// saveJoin 會保存資料表格的加入資訊。
func (b *Builder) saveJoin(table interface{}, typ string, condition string) {
	switch v := table.(type) {
	// 子指令加入。
	case *Builder:
		b.joins[v.query] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	// 普通的表格加入。
	case string:
		b.joins[v] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	}
}

// saveJoinCondition 會將資料表格的加入條件式資訊保存到指定的資料表格加入資訊中。
func (b *Builder) saveJoinCondition(connector string, table interface{}, args ...interface{}) {
	switch v := table.(type) {
	// 子指令條件式。
	case *Builder:
		b.joins[v.query].conditions = append(b.joins[v.query].conditions, condition{
			args:      args,
			connector: connector,
		})
	// 普通條件式。
	case string:
		b.joins[v].conditions = append(b.joins[v].conditions, condition{
			args:      args,
			connector: connector,
		})
	}
}

// saveCondition 會保存欄位的查詢條件。
func (b *Builder) saveCondition(typ, connector string, args ...interface{}) {
	var c condition
	c.connector = connector
	c.args = args
	if typ == "HAVING" {
		b.havingConditions = append(b.havingConditions, c)
	} else {
		b.conditions = append(b.conditions, c)
	}
}

//=======================================================
// 參數函式
//=======================================================

// bindParams 會將接收到的多個變數綁定到本次的建置工作中，並且產生、回傳相對應的 SQL 指令片段。
func (b *Builder) bindParams(data interface{}) (query string) {
	switch d := data.(type) {
	case []interface{}:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", b.bindParam(v))
		}
	case []int:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", b.bindParam(v))
		}
	case []string:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", b.bindParam(v))
		}
	}
	query = trim(query)
	return
}

// bindParam 會將單個傳入的變數綁定到本次的建置工作中，並且依照變數型態來產生並回傳相對應的 SQL 指令片段與決定是否要以括號包覆。
func (b *Builder) bindParam(data interface{}, parentheses ...bool) (param string) {
	switch v := data.(type) {
	case *Builder:
		if len(v.Params()) > 0 {
			b.params = append(b.params, v.Params()...)
		}
	case Function:
		if len(v.values) > 0 {
			b.params = append(b.params, v.values...)
		}
	case nil:
	case Timestamp:
		b.params = append(b.params, v.value)
	default:
		b.params = append(b.params, data)
	}
	param = b.paramToQuery(data, parentheses...)
	return
}

// paramToQuery 會將參數的變數資料型態轉換成 SQL 指令片段，並決定是否要加上括號。
func (b *Builder) paramToQuery(data interface{}, parentheses ...bool) (param string) {
	switch v := data.(type) {
	case *Builder:
		if len(parentheses) > 0 {
			if parentheses[0] == false {
				param = fmt.Sprintf("%s", v.query)
			}
		} else {
			param = fmt.Sprintf("(%s)", v.query)
		}
	case Function:
		param = v.query
	case nil:
		param = "NULL"
	default:
		param = "?"
	}
	return
}

//=======================================================
// 建置函式
//=======================================================

// buildWhere 會基於目前所擁有的條件式來建置一串 `WHERE` 和 `HAVING` 的 SQL 指令。
func (b *Builder) buildWhere(typ string) (query string) {
	var conditions []condition
	if typ == "HAVING" {
		conditions = b.havingConditions
		query = "HAVING "
	} else {
		conditions = b.conditions
		query = "WHERE "
	}
	if len(conditions) == 0 {
		query = ""
		return
	}
	query += b.buildConditions(conditions)
	return
}

// buildUpdate 會建置 `UPDATE` 的 SQL 指令。
func (b *Builder) buildUpdate(data interface{}) (query string, err error) {
	if len(b.tableName) == 0 {
		err = ErrNoTable
		return
	}
	var set string
	beforeOptions, _ := b.buildQueryOptions()
	query = fmt.Sprintf("UPDATE %s%s SET ", beforeOptions, b.tableName[0])

	switch realData := data.(type) {
	case map[string]interface{}:

		var columnNames []string
		for name := range realData {
			columnNames = append(columnNames, name)
		}

		sort.Slice(columnNames, func(i, j int) bool {
			return columnNames[i] < columnNames[j]
		})

		for _, name := range columnNames {
			set += fmt.Sprintf("%s = %s, ", name, b.bindParam(realData[name]))
		}

	}
	query += fmt.Sprintf("%s ", trim(set))
	return
}

// buildLimit 會建置 `LIMIT` 的 SQL 指令。
func (b *Builder) buildLimit() (query string) {
	switch len(b.limit) {
	case 0:
		return
	case 1:
		query = fmt.Sprintf("LIMIT %d ", b.limit[0])
	case 2:
		query = fmt.Sprintf("LIMIT %d, %d ", b.limit[0], b.limit[1])
	}
	return
}

// buildSelect 會建置 `SELECT` 的 SQL 指令。
func (b *Builder) buildSelect(columns ...string) (query string, err error) {
	if len(b.tableName) == 0 {
		err = ErrNoTable
		return
	}
	beforeOptions, _ := b.buildQueryOptions()

	if len(columns) == 0 {
		query = fmt.Sprintf("SELECT %s* FROM %s ", beforeOptions, b.tableName[0])
	} else {
		query = fmt.Sprintf("SELECT %s%s FROM %s ", beforeOptions, strings.Join(columns, ", "), b.tableName[0])
	}
	return
}

// buildConditions 會將傳入的條件式轉換成指定的 `WHERE` 或 `HAVING` SQL 指令。
func (b *Builder) buildConditions(conditions []condition) (query string) {
	for i, v := range conditions {
		// 如果不是第一個條件式的話，那麼就增加連結語句。
		if i != 0 {
			query += fmt.Sprintf("%s ", v.connector)
		}

		// 取得欄位名稱的種類，有可能是個 SQL 指令或普通的欄位名稱、甚至是子指令。
		var typ string
		switch q := v.args[0].(type) {
		case string:
			if strings.Contains(q, "?") || strings.Contains(q, "(") || len(v.args) == 1 {
				typ = "Query"
			} else {
				typ = "Column"
			}
		case *Builder:
			typ = "SubQuery"
		}

		// 基於種類來建置相對應的條件式。
		switch len(v.args) {
		// .Where("Column = Column")
		case 1:
			query += fmt.Sprintf("%s ", v.args[0].(string))
		// .Where("Column = ?", "Value")
		// .Where("Column", "Value")
		// .Where(subQuery, "EXISTS")
		case 2:
			switch typ {
			case "Query":
				query += fmt.Sprintf("%s ", v.args[0].(string))
				b.bindParam(v.args[1])
			case "Column":
				switch d := v.args[1].(type) {
				case Timestamp:
					query += fmt.Sprintf(d.query, v.args[0].(string), b.bindParam(d))
				default:
					query += fmt.Sprintf("%s = %s ", v.args[0].(string), b.bindParam(d))
				}
			case "SubQuery":
				query += fmt.Sprintf("%s %s ", v.args[1].(string), b.bindParam(v.args[0]))
			}
		// .Where("Column", ">", "Value")
		// .Where("Column", "IN", subQuery)
		// .Where("Column", "IS", nil)
		case 3:
			if typ == "Query" {
				query += fmt.Sprintf("%s ", v.args[0].(string))
				b.bindParams(v.args[1:])
			} else {
				if v.args[1].(string) == "IN" || v.args[1].(string) == "NOT IN" {
					query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), b.bindParam(v.args[2], false))
				} else {
					query += fmt.Sprintf("%s %s %s ", v.args[0].(string), v.args[1].(string), b.bindParam(v.args[2]))
				}
			}
		// .Where("(Column = ? OR Column = SHA(?))", "Value", "Value")
		// .Where("Column", "BETWEEN", 1, 20)
		default:
			if typ == "Query" {
				query += fmt.Sprintf("%s ", v.args[0].(string))
				b.bindParams(v.args[1:])
			} else {
				switch v.args[1].(string) {
				case "BETWEEN", "NOT BETWEEN":
					query += fmt.Sprintf("%s %s %s AND %s ", v.args[0].(string), v.args[1].(string), b.bindParam(v.args[2]), b.bindParam(v.args[3]))
				case "IN", "NOT IN":
					query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), b.bindParams(v.args[2:]))
				}
			}
		}
	}
	return
}

// buildDelete 會建置 `DELETE` 的 SQL 指令。
func (b *Builder) buildDelete(tableNames ...string) (query string, err error) {
	if len(b.tableName) == 0 {
		err = ErrNoTable
		return
	}
	beforeOptions, _ := b.buildQueryOptions()
	query += fmt.Sprintf("DELETE %sFROM %s ", beforeOptions, strings.Join(tableNames, ", "))
	return
}

// buildQueryOptions 依照以保存的語句選項來建置執行選項的 SQL 指令片段。
// 這會回傳兩個 SQL 指令片段，分別是放在整體 SQL 指令的前面與後面。
func (b *Builder) buildQueryOptions() (before string, after string) {
	for _, v := range b.queryOptions {
		switch v {
		case "ALL", "DISTINCT", "SQL_CACHE", "SQL_NO_CACHE", "DISTINCTROW", "HIGH_PRIORITY", "STRAIGHT_JOIN", "SQL_SMALL_RESULT", "SQL_BIG_RESULT", "SQL_BUFFER_RESULT", "SQL_CALC_FOUND_ROWS", "LOW_PRIORITY", "QUICK", "IGNORE", "DELAYED":
			before += fmt.Sprintf("%s, ", v)
		case "FOR UPDATE", "LOCK IN SHARE MODE":
			after += fmt.Sprintf("%s, ", v)
		}
	}
	if before != "" {
		before = fmt.Sprintf("%s ", trim(before))
	}
	if after != "" {
		after = fmt.Sprintf("%s ", trim(after))
	}
	return
}

// buildQuery 會將所有建置工作串連起來並且依序執行來建置整個可用的 SQL 指令。
func (b *Builder) buildQuery() {
	b.query += b.buildDuplicate()
	b.query += b.buildJoin()
	b.query += b.buildWhere("WHERE")
	b.query += b.buildWhere("HAVING")
	b.query += b.buildOrderBy()
	b.query += b.buildGroupBy()
	b.query += b.buildLimit()

	_, afterOptions := b.buildQueryOptions()
	b.query += afterOptions
	b.query = strings.TrimSpace(b.query)
}

// buildOrderBy 會基於現有的排序資料來建置 `ORDERY BY` 的 SQL 指令。
func (b *Builder) buildOrderBy() (query string) {
	if len(b.orders) == 0 {
		return
	}
	query += "ORDER BY "
	for _, v := range b.orders {
		switch len(v.args) {
		// .OrderBy("RAND()")
		case 0:
			query += fmt.Sprintf("%s, ", v.column)
		// .OrderBy("ID", "ASC")
		case 1:
			query += fmt.Sprintf("%s %s, ", v.column, v.args[0])
		// .OrderBy("UserGroup", "ASC", "SuperUser", "Admin")
		default:
			query += fmt.Sprintf("FIELD (%s, %s) %s, ", v.column, b.bindParams(v.args[1:]), v.args[0])
		}
	}
	query = trim(query) + " "
	return
}

// buildGroupBy 會建置 `GROUP BY` 的 SQL 指令。
func (b *Builder) buildGroupBy() (query string) {
	if len(b.groupBy) == 0 {
		return
	}
	query += "GROUP BY "
	for _, v := range b.groupBy {
		query += fmt.Sprintf("%s, ", v)
	}
	query = trim(query) + " "
	return
}

// buildDuplicate 會建置 `ON DUPLICATE KEY UPDATE` 的 SQL 指令。
func (b *Builder) buildDuplicate() (query string) {
	if len(b.onDuplicateColumns) == 0 {
		return
	}
	query += "ON DUPLICATE KEY UPDATE "
	if b.lastInsertIDColumn != "" {
		query += fmt.Sprintf("%s=LAST_INSERT_ID(%s), ", b.lastInsertIDColumn, b.lastInsertIDColumn)
	}
	for _, v := range b.onDuplicateColumns {
		query += fmt.Sprintf("%s = VALUES(%s), ", v, v)
	}
	query = trim(query)
	return
}

// buildInsert 會建置 `INSERT INTO` 的 SQL 指令。
func (b *Builder) buildInsert(operator string, data interface{}) (query string, err error) {
	if len(b.tableName) == 0 {
		err = ErrNoTable
		return
	}
	var columns, values string
	beforeOptions, _ := b.buildQueryOptions()

	// 會基於資料型態建置不同的指令。
	switch realData := data.(type) {
	case map[string]interface{}:
		var columnNames []string
		for name := range realData {
			columnNames = append(columnNames, name)
		}

		sort.Slice(columnNames, func(i, j int) bool {
			return columnNames[i] < columnNames[j]
		})

		for _, name := range columnNames {
			columns += fmt.Sprintf("%s, ", name)
			values += fmt.Sprintf("%s, ", b.bindParam(realData[name]))
		}
		values = fmt.Sprintf("(%s)", trim(values))

	case []map[string]interface{}:
		var columnNames []string
		// 先取得欄位的名稱，這樣才能照順序遍歷整個 `map`。
		for name := range realData[0] {
			columnNames = append(columnNames, name)

		}

		sort.Slice(columnNames, func(i, j int) bool {
			return columnNames[i] < columnNames[j]
		})

		for _, name := range columnNames {
			// 先建置欄位名稱的 SQL 指令片段。
			columns += fmt.Sprintf("%s, ", name)
			//
			//
			//

		}
		for _, single := range realData {
			var currentValues string
			for _, name := range columnNames {
				currentValues += fmt.Sprintf("%s, ", b.bindParam(single[name]))
			}
			values += fmt.Sprintf("(%s), ", trim(currentValues))
		}
		values = trim(values)
	}
	columns = trim(columns)
	query = fmt.Sprintf("%s %sINTO %s (%s) VALUES %s ", operator, beforeOptions, b.tableName[0], columns, values)
	return
}

// buildJoin 會建置資料表的插入 SQL 指令。
func (b *Builder) buildJoin() (query string) {
	if len(b.joins) == 0 {
		return
	}

	for _, v := range b.joins {
		// 插入的種類（例如：`LEFT JOIN`、`RIGHT JOIN`、`INNER JOIN`）。
		query += fmt.Sprintf("%s ", v.typ)
		switch d := v.table.(type) {
		// 子指令。
		case *Builder:
			query += fmt.Sprintf("%s AS %s ON ", b.bindParam(d), d.alias)
		// 資料表格名稱。
		case string:
			query += fmt.Sprintf("%s ON ", d)
		}

		if len(v.conditions) == 0 {
			query += fmt.Sprintf("(%s) ", v.condition)
		} else {
			conditionsQuery := strings.TrimSpace(b.buildConditions(v.conditions))
			query += fmt.Sprintf("(%s %s %s) ", v.condition, v.conditions[0].connector, conditionsQuery)
		}
	}
	return
}

//=======================================================
// 執行函式
//=======================================================

// runQuery 會以 `Query` 的方式執行建置出來的 SQL 指令。
func (b *Builder) runQuery() (rows *sql.Rows, err error) {
	b.cleanBefore()
	b.buildQuery()
	b.LastQuery = b.query
	b.LastParams = b.params

	// 如果有啟用追蹤模式的話，開始計算執行時間。
	var start time.Time
	if b.tracing {
		start = time.Now()
	}

	// 如果這個建置建構體是可執行的話，就執行 SQL 指令。
	if b.executable {
		var stmt *sql.Stmt
		var count int
		var tx *sql.Tx

		// 如果指令選項中有 `SQL_CALC_FOUND_ROWS` 的話就開始一段交易，
		// 因為這個指令僅能用於同個連線中。
		for _, v := range b.queryOptions {
			if v == "SQL_CALC_FOUND_ROWS" {
				// 開始一個交易。
				tx, err = b.db.Begin()
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				// 準備執行指令。
				stmt, err = tx.Prepare(b.query)
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				// 傳入參數並且執行指令。
				rows, err = stmt.Query(b.params...)
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				// 將取得到的結果映射置目的地指標。
				// 這同時會關閉 `rows` 所以就不會觸發 `busy buffer` 錯誤。
				count, err = load(rows, b.destination)
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				// 保存結果行數。
				b.count = count

				// 選擇 `FOUND_ROWS` 來取得總計的行數。
				rows, err = tx.Query("SELECT FOUND_ROWS()")
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				// 掃描資料來取得總計的行數。
				for rows.Next() {
					var totalCount int
					rows.Scan(&totalCount)
					if rows.Err() != nil {
						err = rows.Err()
						b.saveTrace(err, b.query, start)
						b.cleanAfter()
						return
					}
					b.TotalCount = totalCount
				}
				// 關閉、結束整個指令環境。
				err = stmt.Close()
				if err != nil {
					b.saveTrace(err, b.query, start)
					b.cleanAfter()
					return
				}
				b.saveTrace(err, b.query, start)
				b.cleanAfter()
				return
			}
		}

		// 如果沒有設置 `SQL_CALC_FOUND_ROWS` 的話就使用正常的連線池。
		stmt, err = b.db.Prepare(b.query)
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		rows, err = stmt.Query(b.params...)
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		err = stmt.Close()
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		count, err = load(rows, b.destination)
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		b.count = count
	}
	b.saveTrace(err, b.query, start)
	b.cleanAfter()
	return
}

// executeQuery 會透過 `Exec` 的方式執行建置出來的 SQL 指令。
func (b *Builder) executeQuery() (res sql.Result, err error) {
	b.cleanBefore()
	b.buildQuery()
	b.LastQuery = b.query
	b.LastParams = b.params

	// 如果有啟用追蹤模式的話，開始計算執行時間。
	var start time.Time
	if b.tracing {
		start = time.Now()
	}

	// 如果這個建置建構體是可執行的話，就執行 SQL 指令。
	if b.executable {
		var stmt *sql.Stmt
		var count int64
		stmt, err = b.db.Prepare(b.query)
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		res, err = stmt.Exec(b.params...)
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		b.LastResult = res
		count, err = res.RowsAffected()
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
		b.count = int(count)
		err = stmt.Close()
		if err != nil {
			b.saveTrace(err, b.query, start)
			b.cleanAfter()
			return
		}
	}
	b.saveTrace(err, b.query, start)
	b.cleanAfter()
	return
}

//=======================================================
// 輸出函式
//=======================================================

// Table 能夠指定資料表格的名稱。
func (b *Builder) Table(tableName ...string) (builder *Builder) {
	builder = b.clone()
	builder.tableName = tableName
	return
}

//=======================================================
// 選擇函式
//=======================================================

// Get 會取得多列的資料結果，傳入的參數為欲取得的欄位名稱，不傳入參數表示取得所有欄位。
func (b *Builder) Get(columns ...string) (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildSelect(columns...)
	if err != nil {
		return
	}
	_, err = builder.runQuery()
	if err != nil {
		return
	}
	return
}

// GetValue 會取得單個欄位的資料（例如：字串、正整數）。
func (b *Builder) GetValue(column string) (builder *Builder, err error) {
	builder, err = b.GetOne(column)
	return
}

// GetValues 會取得將多筆單個欄位的資料映射到本地的字串、正整數切片、陣列。
func (b *Builder) GetValues(column string) (builder *Builder, err error) {
	builder, err = b.Get(column)
	return
}

// GetOne 會取得僅單列的資料作為結果，傳入的參數為欲取得的欄位名稱，不傳入參數表示取得所有欄位。
// 簡單說，這就是 `.Limit(1).Get()` 的縮寫用法。
func (b *Builder) GetOne(columns ...string) (builder *Builder, err error) {
	builder, err = b.clone().Limit(1).Get(columns...)
	return
}

// Paginate 基本上和 `Get` 取得函式無異，但此函式能夠自動依照分頁數來推算該從哪裡繼續取得資料。
// 使用時須先確定是否有指定 `PageLimit`（預設為：20），這樣才能限制一頁有多少筆資料。
func (b *Builder) Paginate(pageCount int, columns ...string) (builder *Builder, err error) {
	builder, err = b.WithTotalCount().Limit(b.PageLimit*(pageCount-1), b.PageLimit).Get(columns...)
	builder.TotalPage = builder.TotalCount / builder.PageLimit
	return
}

// WithTotalCount 會在 SQL 執行指令中安插 `SQL_CALC_FOUND_ROWS` 選項，
// 如此一來就能夠在執行完 SQL 指令後取得查詢的總計行數。在不同情況下，這可能會拖低執行效能。
func (b *Builder) WithTotalCount() (builder *Builder) {
	builder = b.clone().SetQueryOption("SQL_CALC_FOUND_ROWS")
	return
}

//=======================================================
// 插入函式
//=======================================================

// Insert 會插入一筆新的資料。
func (b *Builder) Insert(data interface{}) (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildInsert("INSERT", data)
	if err != nil {
		return
	}
	res, err := builder.executeQuery()
	if err != nil || !builder.executable {
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		return
	}
	builder.LastInsertID = int(id)
	return
}

// InsertMulti 會一次插入多筆資料。
func (b *Builder) InsertMulti(data interface{}) (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildInsert("INSERT", data)
	if err != nil {
		return
	}
	res, err := builder.executeQuery()
	if err != nil || !builder.executable {
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		return
	}
	builder.LastInsertID = int(id)
	return
}

// Delete 會移除相符的資料列，記得用上 `Where` 條件式來避免整個資料表格被清空。
// 這很重要好嗎，因為⋯你懂的⋯。喔，不。
func (b *Builder) Delete() (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildDelete(b.tableName...)
	if err != nil {
		return
	}
	_, err = builder.executeQuery()
	return
}

//=======================================================
// 更新函式
//=======================================================

// Replace 基本上和 `Insert` 無異，這會在有重複資料時移除該筆資料並重新插入。
// 若無該筆資料則插入新的資料。
func (b *Builder) Replace(data interface{}) (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildInsert("REPLACE", data)
	if err != nil {
		return
	}
	_, err = builder.executeQuery()
	return
}

// Update 會以指定的資料來更新相對應的資料列。
func (b *Builder) Update(data interface{}) (builder *Builder, err error) {
	builder = b.clone()
	builder.query, err = builder.buildUpdate(data)
	_, err = builder.executeQuery()
	return
}

// OnDuplicate 能夠指定欲更新的欄位名稱，這會在插入的資料重複時自動更新相對應的欄位。
func (b *Builder) OnDuplicate(columns []string, lastInsertID ...string) (builder *Builder) {
	builder = b.clone()
	builder.onDuplicateColumns = columns
	if len(lastInsertID) != 0 {
		builder.lastInsertIDColumn = lastInsertID[0]
	}
	return
}

//=======================================================
// 限制函式
//=======================================================

// Limit 能夠在 SQL 查詢指令中建立限制筆數的條件。
func (b *Builder) Limit(from int, count ...int) (builder *Builder) {
	builder = b.clone()
	if len(count) == 0 {
		builder.limit = []int{from}
	} else {
		builder.limit = []int{from, count[0]}
	}
	return
}

// OrderBy 會依照指定的欄位來替結果做出排序（例如：`DESC`、`ASC`）。
func (b *Builder) OrderBy(column string, args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.orders = append(builder.orders, order{
		column: column,
		args:   args,
	})
	return
}

// GroupBy 會在執行 SQL 指令時依照特定的欄位來做執行區分。
func (b *Builder) GroupBy(columns ...string) (builder *Builder) {
	builder = b.clone()
	builder.groupBy = columns
	return
}

//=======================================================
// 指令函式
//=======================================================

// RawQuery 會接收傳入的變數來執行傳入的 SQL 執行語句，變數可以在語句中以 `?`（Prepared Statements）使用來避免 SQL 注入攻擊。
// 這會將多筆資料映射到本地的建構體切片、陣列上。
func (b *Builder) RawQuery(query string, values ...interface{}) (builder *Builder, err error) {
	builder = b.clone()
	builder.query = query
	builder.params = values
	_, err = builder.runQuery()
	return
}

// RawQueryValue 與 RawQuery 使用方法相同，但這會在 SQL 指令後追加 `LIMIT 1` 並將單個欄位的資料映射到本地的變數（字串、正整數），
// 這很適合用於像是你想要取得單個使用者暱稱的時候。
func (b *Builder) RawQueryValue(query string, values ...interface{}) (builder *Builder, err error) {
	builder, err = b.RawQuery(fmt.Sprintf("%s LIMIT 1", strings.TrimSpace(query)), values...)
	return
}

// RawQueryValues 與 RawQuery 使用方法相同，但這會將多筆單個欄位的資料映射到本地的字串、正整數切片、陣列。
// 這很適合用於像是你想要取得多個使用者暱稱陣列的時候。
func (b *Builder) RawQueryValues(query string, values ...interface{}) (builder *Builder, err error) {
	builder, err = b.RawQuery(query, values...)
	return
}

// RawQueryOne 與 RawQuery 使用方法相同，但這會在 SQL 指令後追加 `LIMIT 1` 且僅會將單筆資料映射到本地的建構體、`map`。
func (b *Builder) RawQueryOne(query string, values ...interface{}) (builder *Builder, err error) {
	builder, err = b.RawQuery(fmt.Sprintf("%s LIMIT 1", strings.TrimSpace(query)), values...)
	return
}

//=======================================================
// 條件函式
//=======================================================

// Where 會增加一個 `WHERE AND` 條件式。
func (b *Builder) Where(args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveCondition("WHERE", "AND", args...)
	return
}

// OrWhere 會增加一個 `WHERE OR` 條件式。
func (b *Builder) OrWhere(args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveCondition("WHERE", "OR", args...)
	return
}

// Having 會增加一個 `HAVING AND` 條件式。
func (b *Builder) Having(args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveCondition("HAVING", "AND", args...)
	return
}

// OrHaving 會增加一個 `HAVING OR` 條件式。
func (b *Builder) OrHaving(args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveCondition("HAVING", "OR", args...)
	return
}

//=======================================================
// 加入函式
//=======================================================

// LeftJoin 會向左插入一個資料表格。
func (b *Builder) LeftJoin(table interface{}, condition string) (builder *Builder) {
	builder = b.clone()
	builder.saveJoin(table, "LEFT JOIN", condition)
	return
}

// RightJoin 會向右插入一個資料表格。
func (b *Builder) RightJoin(table interface{}, condition string) (builder *Builder) {
	builder = b.clone()
	builder.saveJoin(table, "RIGHT JOIN", condition)
	return
}

// InnerJoin 會內部插入一個資料表格。
func (b *Builder) InnerJoin(table interface{}, condition string) (builder *Builder) {
	builder = b.clone()
	builder.saveJoin(table, "INNER JOIN", condition)
	return
}

// NaturalJoin 會自然插入一個資料表格。
func (b *Builder) NaturalJoin(table interface{}, condition string) (builder *Builder) {
	builder = b.clone()
	builder.saveJoin(table, "NATURAL JOIN", condition)
	return
}

// JoinWhere 能夠建立一個基於 `WHERE AND` 的條件式給某個指定的插入資料表格。
func (b *Builder) JoinWhere(table interface{}, args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveJoinCondition("AND", table, args...)
	return
}

// JoinOrWhere 能夠建立一個基於 `WHERE OR` 的條件式給某個指定的插入資料表格。
func (b *Builder) JoinOrWhere(table interface{}, args ...interface{}) (builder *Builder) {
	builder = b.clone()
	builder.saveJoinCondition("OR", table, args...)
	return
}

// SubQuery 能夠將目前的 SQL 指令轉換為子指令（Sub Query）來防止建置後直接被執行，這讓你可以將子指令傳入其他的條件式（例如：`WHERE`），
// 若欲將子指令傳入插入（Join）條件中，必須在參數指定此子指令的別名。
func (b *Builder) SubQuery(alias ...string) (newBuilder *Builder) {
	newBuilder = &Builder{
		executable: false,
	}
	if len(alias) > 0 {
		newBuilder.alias = alias[0]
	}
	return
}

// Has 會在有查詢結果時回傳 `true`，這很適合用於一些資料驗證的時機（例如：使用者名稱是否已存在⋯等）。
func (b *Builder) Has() (builder *Builder, has bool, err error) {
	builder, err = b.Limit(1).Get()
	if err != nil {
		has = false
		return
	}
	if builder.Count() > 0 {
		has = true
		return
	}
	return
}

//=======================================================
// 資料庫函式
//=======================================================

// Disconnect 會結束目前的資料庫連線。
func (b *Builder) Disconnect() (err error) {
	err = b.db.Disconnect()
	return
}

// Ping 會以 ping 來檢查資料庫連線。
func (b *Builder) Ping() (err error) {
	err = b.db.Ping()
	return
}

// Connect 會試圖在斷線之後重新連線至資料庫。
func (b *Builder) Connect() (err error) {
	err = b.db.Connect()
	return
}

//=======================================================
// 交易函式
//=======================================================

// Begin 會開始一個新的交易。
func (b *Builder) Begin() (builder *Builder, err error) {
	builder = b.clone()
	var tx *sql.Tx
	tx, err = builder.db.Begin()
	if err != nil {
		return
	}
	master := *builder.db.master
	builder.db.master = &master
	builder.db.master.tx = tx
	return
}

// Rollback 能夠回溯到交易剛開始的時候，並且在不保存資料變動的情況下結束交易。
func (b *Builder) Rollback() error {
	return b.db.Rollback()
}

// Commit 會讓交易中所產生的資料異動成為永久紀錄並保存於資料庫中且結束交易。
func (b *Builder) Commit() error {
	return b.db.Commit()
}

//=======================================================
// 輔助函式
//=======================================================

// Count 會回傳查詢的結果行數。
func (b *Builder) Count() (count int) {
	count = b.count
	return
}

// Func 會基於參數來返回一個新的 SQL 資料庫函式，
// 這能夠當作函式放置於查詢指令中，而不會被當作普通的資料執行。
func (b *Builder) Func(query string, data ...interface{}) Function {
	return Function{
		query:  query,
		values: data,
	}
}

// Now 會回傳一個基於 `INTERVAL` 的 SQL 資料庫函式，
// 傳入的參數格式可以是 `+1Y`、`-2M`，同時也可以像 `Now("+1Y", "-2M")` 一樣地串連使用。
// 支援的格式為：`Y`(年)、`M`(月)、`D`(日)、`W`(星期)、`h`(小時)、`m`(分鐘)、`s`(秒數)。
func (b *Builder) Now(formats ...string) Function {
	query := "NOW() "
	unitMap := map[string]string{
		"Y": "YEAR",
		"M": "MONTH",
		"D": "DAY",
		"W": "WEEK",
		"h": "HOUR",
		"m": "MINUTE",
		"s": "SECOND",
	}
	for _, v := range formats {
		operator := string(v[0])
		interval := v[1 : len(v)-1]
		unit := string(v[len(v)-1])
		query += fmt.Sprintf("%s INTERVAL %s %s ", operator, interval, unitMap[unit])
	}
	return b.Func(strings.TrimSpace(query))
}

// SetLockMethod 會設置鎖定資料表格的方式（例如：`WRITE`、`READ`）。
func (b *Builder) SetLockMethod(method string) (builder *Builder) {
	builder = b.clone()
	builder.lockMethod = strings.ToUpper(method)
	return
}

// Lock 會以指定的上鎖方式來鎖定某個指定的資料表格，這能用以避免資料競爭問題。
func (b *Builder) Lock(tableNames ...string) (builder *Builder, err error) {
	var tables string
	for _, v := range tableNames {
		tables += fmt.Sprintf("%s %s, ", v, b.lockMethod)
	}
	tables = trim(tables)

	builder, err = b.RawQuery(fmt.Sprintf("LOCK TABLES %s", tables))
	return
}

// Unlock 能解鎖已鎖上的資料表格。
func (b *Builder) Unlock(tableNames ...string) (builder *Builder, err error) {
	builder, err = b.RawQuery("UNLOCK TABLES")
	return
}

// SetQueryOption 會設置 SQL 指令的額外選項（例如：`SQL_NO_CACHE`）。
func (b *Builder) SetQueryOption(options ...string) (builder *Builder) {
	builder = b.clone()
	builder.queryOptions = append(builder.queryOptions, options...)
	return
}

// SetTrace 會決定蹤跡模式的開關，當設置為 `true` 時會稍微地拖慢效能，
// 但你就能夠從 `Trace` 屬性中取得 SQL 執行後的堆疊與路徑結果。
func (b *Builder) SetTrace(status bool) (builder *Builder) {
	builder = b.clone()
	builder.tracing = status
	return
}

//=======================================================
// 物件函式
//=======================================================

// Bind 會設置資料的映射目的地，這樣就能在 SQL 指令執行後將資料映射到某個變數、記憶體指標。
func (b *Builder) Bind(destination interface{}) (builder *Builder) {
	builder = b.clone()
	builder.destination = destination
	return
}

// Query 會回傳最後一次所建置的 SQL 執行指令，這和 `LastQuery` 相同功能但這函式的名稱可能更符合某些場合。
func (b *Builder) Query() (query string) {
	query = b.LastQuery
	return
}

// Params returns the last used parameters, it's the same as `LastParams` but the name is more meaningful.
func (b *Builder) Params() (params []interface{}) {
	params = b.LastParams
	return
}

//=======================================================
// 其他函式
//=======================================================

// Migration 會返回一個新的資料表格遷移建構體。
// 主要是基於現有的資料庫連線來提供資料表格與欄位的的操作功能。
func (b *Builder) Migration() *Migration {
	return newMigration(b.db)
}
