package reiner

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
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

// Wrapper 是個資料庫的 SQL 指令建置系統，同時也帶有資料庫的連線資料。
type Wrapper struct {
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

// newWrapper 會基於傳入的資料庫連線來建立一個新的 SQL 指令建置系統。
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{executable: true, db: db, Timestamp: &Timestamp{}, joins: make(map[string]*join)}
}

// cloning 會複製現有的資料庫包覆函式。
func (w *Wrapper) cloning(deepCopy bool, database ...*DB) (clonedWrapper *Wrapper) {
	db := w.db
	if len(database) > 0 {
		db = database[0]
	}
	if !deepCopy {
		clonedWrapper = &Wrapper{
			db:         db,
			executable: true,
			Timestamp:  &Timestamp{},
			joins:      make(map[string]*join),
		}
	} else {
		clonedWrapper = &Wrapper{
			db:                 db,
			executable:         true,
			Timestamp:          &Timestamp{},
			alias:              w.alias,
			destination:        w.destination,
			tableName:          w.tableName,
			conditions:         w.conditions,
			havingConditions:   w.havingConditions,
			queryOptions:       w.queryOptions,
			joins:              w.joins,
			onDuplicateColumns: w.onDuplicateColumns,
			lastInsertIDColumn: w.lastInsertIDColumn,
			limit:              w.limit,
			orders:             w.orders,
			groupBy:            w.groupBy,
			lockMethod:         w.lockMethod,
			tracing:            w.tracing,
			query:              w.query,
			params:             w.params,
		}
	}
	return
}

// cleanAfter 會在 SQL 指令建置之後清除資料來避免下次使用到舊的資料。
func (w *Wrapper) cleanAfter() {
	w.queryOptions = []string{}
	w.tableName = []string{}
	w.params = []interface{}{}
	w.onDuplicateColumns = []string{}
	w.groupBy = []string{}
	w.joins = map[string]*join{}
	w.orders = []order{}
	w.conditions = []condition{}
	w.havingConditions = []condition{}
	w.limit = []int{}
	w.destination = nil
}

// cleanBefore 會在 SQL 指令建置之前清除以往的資料，
// 留到建置前才清除的原因是避免需要歷史資料的時候卻在執行後直接被清除了。
func (w *Wrapper) cleanBefore() {
	w.TotalCount = 0
	w.TotalPage = 0
	w.LastInsertID = 0
	w.LastResult = nil
	w.LastParams = []interface{}{}
	w.count = 0
}

//=======================================================
// 保存函式
//=======================================================

// saveTrace 會取得、紀錄呼叫函式的名稱，並且計算執行時間然後將其保存於蹤跡資訊中。
func (w *Wrapper) saveTrace(err error, query string, startedAt time.Time) {
	if !w.tracing {
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
	w.Traces = append(w.Traces, Trace{
		Query:    query,
		Duration: time.Since(startedAt),
		Stacks:   stacks,
		Error:    err,
	})
}

// saveJoin 會保存資料表格的加入資訊。
func (w *Wrapper) saveJoin(table interface{}, typ string, condition string) {
	switch v := table.(type) {
	// 子指令加入。
	case *Wrapper:
		w.joins[v.query] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	// 普通的表格加入。
	case string:
		w.joins[v] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	}
}

// saveJoinCondition 會將資料表格的加入條件式資訊保存到指定的資料表格加入資訊中。
func (w *Wrapper) saveJoinCondition(connector string, table interface{}, args ...interface{}) {
	switch v := table.(type) {
	// 子指令條件式。
	case *Wrapper:
		w.joins[v.query].conditions = append(w.joins[v.query].conditions, condition{
			args:      args,
			connector: connector,
		})
	// 普通條件式。
	case string:
		w.joins[v].conditions = append(w.joins[v].conditions, condition{
			args:      args,
			connector: connector,
		})
	}
}

// saveCondition 會保存欄位的查詢條件。
func (w *Wrapper) saveCondition(typ, connector string, args ...interface{}) {
	var c condition
	c.connector = connector
	c.args = args
	if typ == "HAVING" {
		w.havingConditions = append(w.havingConditions, c)
	} else {
		w.conditions = append(w.conditions, c)
	}
}

//=======================================================
// 參數函式
//=======================================================

// bindParams 會將接收到的多個變數綁定到本次的建置工作中，並且產生、回傳相對應的 SQL 指令片段。
func (w *Wrapper) bindParams(data interface{}) (query string) {
	switch d := data.(type) {
	case []interface{}:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	case []int:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	case []string:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	}
	query = trim(query)
	return
}

// bindParam 會將單個傳入的變數綁定到本次的建置工作中，並且依照變數型態來產生並回傳相對應的 SQL 指令片段與決定是否要以括號包覆。
func (w *Wrapper) bindParam(data interface{}, parentheses ...bool) (param string) {
	switch v := data.(type) {
	case *Wrapper:
		if len(v.Params()) > 0 {
			w.params = append(w.params, v.Params()...)
		}
	case Function:
		if len(v.values) > 0 {
			w.params = append(w.params, v.values...)
		}
	case nil:
	case Timestamp:
		w.params = append(w.params, v.value)
	default:
		w.params = append(w.params, data)
	}
	param = w.paramToQuery(data, parentheses...)
	return
}

// paramToQuery 會將參數的變數資料型態轉換成 SQL 指令片段，並決定是否要加上括號。
func (w *Wrapper) paramToQuery(data interface{}, parentheses ...bool) (param string) {
	switch v := data.(type) {
	case *Wrapper:
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
func (w *Wrapper) buildWhere(typ string) (query string) {
	var conditions []condition
	if typ == "HAVING" {
		conditions = w.havingConditions
		query = "HAVING "
	} else {
		conditions = w.conditions
		query = "WHERE "
	}
	if len(conditions) == 0 {
		query = ""
		return
	}
	query += w.buildConditions(conditions)
	return
}

// buildUpdate 會建置 `UPDATE` 的 SQL 指令。
func (w *Wrapper) buildUpdate(data interface{}) (query string) {
	var set string
	beforeOptions, _ := w.buildQueryOptions()
	query = fmt.Sprintf("UPDATE %s%s SET ", beforeOptions, w.tableName[0])

	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			set += fmt.Sprintf("%s = %s, ", column, w.bindParam(value))
		}
	}
	query += fmt.Sprintf("%s ", trim(set))
	return
}

// buildLimit 會建置 `LIMIT` 的 SQL 指令。
func (w *Wrapper) buildLimit() (query string) {
	switch len(w.limit) {
	case 0:
		return
	case 1:
		query = fmt.Sprintf("LIMIT %d ", w.limit[0])
	case 2:
		query = fmt.Sprintf("LIMIT %d, %d ", w.limit[0], w.limit[1])
	}
	return
}

// buildSelect 會建置 `SELECT` 的 SQL 指令。
func (w *Wrapper) buildSelect(columns ...string) (query string) {
	beforeOptions, _ := w.buildQueryOptions()

	if len(columns) == 0 {
		query = fmt.Sprintf("SELECT %s* FROM %s ", beforeOptions, w.tableName[0])
	} else {
		query = fmt.Sprintf("SELECT %s%s FROM %s ", beforeOptions, strings.Join(columns, ", "), w.tableName[0])
	}
	return
}

// buildConditions 會將傳入的條件式轉換成指定的 `WHERE` 或 `HAVING` SQL 指令。
func (w *Wrapper) buildConditions(conditions []condition) (query string) {
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
		case *Wrapper:
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
				w.bindParam(v.args[1])
			case "Column":
				switch d := v.args[1].(type) {
				case Timestamp:
					query += fmt.Sprintf(d.query, v.args[0].(string), w.bindParam(d))
				default:
					query += fmt.Sprintf("%s = %s ", v.args[0].(string), w.bindParam(d))
				}
			case "SubQuery":
				query += fmt.Sprintf("%s %s ", v.args[1].(string), w.bindParam(v.args[0]))
			}
		// .Where("Column", ">", "Value")
		// .Where("Column", "IN", subQuery)
		// .Where("Column", "IS", nil)
		case 3:
			if typ == "Query" {
				query += fmt.Sprintf("%s ", v.args[0].(string))
				w.bindParams(v.args[1:])
			} else {
				if v.args[1].(string) == "IN" || v.args[1].(string) == "NOT IN" {
					query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2], false))
				} else {
					query += fmt.Sprintf("%s %s %s ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2]))
				}
			}
		// .Where("(Column = ? OR Column = SHA(?))", "Value", "Value")
		// .Where("Column", "BETWEEN", 1, 20)
		default:
			if typ == "Query" {
				query += fmt.Sprintf("%s ", v.args[0].(string))
				w.bindParams(v.args[1:])
			} else {
				switch v.args[1].(string) {
				case "BETWEEN", "NOT BETWEEN":
					query += fmt.Sprintf("%s %s %s AND %s ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2]), w.bindParam(v.args[3]))
				case "IN", "NOT IN":
					query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), w.bindParams(v.args[2:]))
				}
			}
		}
	}
	return
}

// buildDelete 會建置 `DELETE` 的 SQL 指令。
func (w *Wrapper) buildDelete(tableNames ...string) (query string) {
	beforeOptions, _ := w.buildQueryOptions()
	query += fmt.Sprintf("DELETE %sFROM %s ", beforeOptions, strings.Join(tableNames, ", "))
	return
}

// buildQueryOptions 依照以保存的語句選項來建置執行選項的 SQL 指令片段。
// 這會回傳兩個 SQL 指令片段，分別是放在整體 SQL 指令的前面與後面。
func (w *Wrapper) buildQueryOptions() (before string, after string) {
	for _, v := range w.queryOptions {
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
func (w *Wrapper) buildQuery() {
	w.query += w.buildDuplicate()
	w.query += w.buildJoin()
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query += w.buildGroupBy()
	w.query += w.buildLimit()

	_, afterOptions := w.buildQueryOptions()
	w.query += afterOptions
	w.query = strings.TrimSpace(w.query)
}

// buildOrderBy 會基於現有的排序資料來建置 `ORDERY BY` 的 SQL 指令。
func (w *Wrapper) buildOrderBy() (query string) {
	if len(w.orders) == 0 {
		return
	}
	query += "ORDER BY "
	for _, v := range w.orders {
		switch len(v.args) {
		// .OrderBy("RAND()")
		case 0:
			query += fmt.Sprintf("%s, ", v.column)
		// .OrderBy("ID", "ASC")
		case 1:
			query += fmt.Sprintf("%s %s, ", v.column, v.args[0])
		// .OrderBy("UserGroup", "ASC", "SuperUser", "Admin")
		default:
			query += fmt.Sprintf("FIELD (%s, %s) %s, ", v.column, w.bindParams(v.args[1:]), v.args[0])
		}
	}
	query = trim(query) + " "
	return
}

// buildGroupBy 會建置 `GROUP BY` 的 SQL 指令。
func (w *Wrapper) buildGroupBy() (query string) {
	if len(w.groupBy) == 0 {
		return
	}
	query += "GROUP BY "
	for _, v := range w.groupBy {
		query += fmt.Sprintf("%s, ", v)
	}
	query = trim(query) + " "
	return
}

// buildDuplicate 會建置 `ON DUPLICATE KEY UPDATE` 的 SQL 指令。
func (w *Wrapper) buildDuplicate() (query string) {
	if len(w.onDuplicateColumns) == 0 {
		return
	}
	query += "ON DUPLICATE KEY UPDATE "
	if w.lastInsertIDColumn != "" {
		query += fmt.Sprintf("%s=LAST_INSERT_ID(%s), ", w.lastInsertIDColumn, w.lastInsertIDColumn)
	}
	for _, v := range w.onDuplicateColumns {
		query += fmt.Sprintf("%s = VALUES(%s), ", v, v)
	}
	query = trim(query)
	return
}

// buildInsert 會建置 `INSERT INTO` 的 SQL 指令。
func (w *Wrapper) buildInsert(operator string, data interface{}) (query string) {
	var columns, values string
	beforeOptions, _ := w.buildQueryOptions()

	// 會基於資料型態建置不同的指令。
	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			columns += fmt.Sprintf("%s, ", column)
			values += fmt.Sprintf("%s, ", w.bindParam(value))
		}
		values = fmt.Sprintf("(%s)", trim(values))

	case []map[string]interface{}:
		var columnNames []string
		// 先取得欄位的名稱，這樣才能照順序遍歷整個 `map`。
		for name := range realData[0] {
			columnNames = append(columnNames, name)
			// 先建置欄位名稱的 SQL 指令片段。
			columns += fmt.Sprintf("%s, ", name)
		}
		for _, single := range realData {
			var currentValues string
			for _, name := range columnNames {
				currentValues += fmt.Sprintf("%s, ", w.bindParam(single[name]))
			}
			values += fmt.Sprintf("(%s), ", trim(currentValues))
		}
		values = trim(values)
	}
	columns = trim(columns)
	query = fmt.Sprintf("%s %sINTO %s (%s) VALUES %s ", operator, beforeOptions, w.tableName[0], columns, values)
	return
}

// buildJoin 會建置資料表的插入 SQL 指令。
func (w *Wrapper) buildJoin() (query string) {
	if len(w.joins) == 0 {
		return
	}

	for _, v := range w.joins {
		// 插入的種類（例如：`LEFT JOIN`、`RIGHT JOIN`、`INNER JOIN`）。
		query += fmt.Sprintf("%s ", v.typ)
		switch d := v.table.(type) {
		// 子指令。
		case *Wrapper:
			query += fmt.Sprintf("%s AS %s ON ", w.bindParam(d), d.alias)
		// 資料表格名稱。
		case string:
			query += fmt.Sprintf("%s ON ", d)
		}

		if len(v.conditions) == 0 {
			query += fmt.Sprintf("(%s) ", v.condition)
		} else {
			conditionsQuery := strings.TrimSpace(w.buildConditions(v.conditions))
			query += fmt.Sprintf("(%s %s %s) ", v.condition, v.conditions[0].connector, conditionsQuery)
		}
	}
	return
}

//=======================================================
// 執行函式
//=======================================================

// runQuery 會以 `Query` 的方式執行建置出來的 SQL 指令。
func (w *Wrapper) runQuery() (rows *sql.Rows, err error) {
	w.cleanBefore()
	w.buildQuery()
	w.LastQuery = w.query
	w.LastParams = w.params

	// 如果有啟用追蹤模式的話，開始計算執行時間。
	var start time.Time
	if w.tracing {
		start = time.Now()
	}

	// 如果這個建置建構體是可執行的話，就執行 SQL 指令。
	if w.executable {
		var stmt *sql.Stmt
		var count int
		var tx *sql.Tx

		// 如果指令選項中有 `SQL_CALC_FOUND_ROWS` 的話就開始一段交易，
		// 因為這個指令僅能用於同個連線中。
		for _, v := range w.queryOptions {
			if v == "SQL_CALC_FOUND_ROWS" {
				// 開始一個交易。
				tx, err = w.db.Begin()
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// 準備執行指令。
				stmt, err = tx.Prepare(w.query)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// 傳入參數並且執行指令。
				rows, err = stmt.Query(w.params...)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// 將取得到的結果映射置目的地指標。
				// 這同時會關閉 `rows` 所以就不會觸發 `busy buffer` 錯誤。
				count, err = load(rows, w.destination)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// 保存結果行數。
				w.count = count

				// 選擇 `FOUND_ROWS` 來取得總計的行數。
				rows, err = tx.Query("SELECT FOUND_ROWS()")
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// 掃描資料來取得總計的行數。
				for rows.Next() {
					var totalCount int
					rows.Scan(&totalCount)
					if rows.Err() != nil {
						err = rows.Err()
						w.saveTrace(err, w.query, start)
						w.cleanAfter()
						return
					}
					w.TotalCount = totalCount
				}
				// 關閉、結束整個指令環境。
				err = stmt.Close()
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				w.saveTrace(err, w.query, start)
				w.cleanAfter()
				return
			}
		}

		// 如果沒有設置 `SQL_CALC_FOUND_ROWS` 的話就使用正常的連線池。
		stmt, err = w.db.Prepare(w.query)
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		rows, err = stmt.Query(w.params...)
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		err = stmt.Close()
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		count, err = load(rows, w.destination)
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		w.count = count
	}
	w.saveTrace(err, w.query, start)
	w.cleanAfter()
	return
}

// executeQuery 會透過 `Exec` 的方式執行建置出來的 SQL 指令。
func (w *Wrapper) executeQuery() (res sql.Result, err error) {
	w.cleanBefore()
	w.buildQuery()
	w.LastQuery = w.query
	w.LastParams = w.params

	// 如果有啟用追蹤模式的話，開始計算執行時間。
	var start time.Time
	if w.tracing {
		start = time.Now()
	}

	// 如果這個建置建構體是可執行的話，就執行 SQL 指令。
	if w.executable {
		var stmt *sql.Stmt
		var count int64
		stmt, err = w.db.Prepare(w.query)
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		res, err = stmt.Exec(w.params...)
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		w.LastResult = res
		count, err = res.RowsAffected()
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
		w.count = int(count)
		err = stmt.Close()
		if err != nil {
			w.saveTrace(err, w.query, start)
			w.cleanAfter()
			return
		}
	}
	w.saveTrace(err, w.query, start)
	w.cleanAfter()
	return
}

//=======================================================
// 輸出函式
//=======================================================

// Table 能夠指定資料表格的名稱。
func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

//=======================================================
// 選擇函式
//=======================================================

// Get 會取得多列的資料結果，傳入的參數為欲取得的欄位名稱，不傳入參數表示取得所有欄位。
func (w *Wrapper) Get(columns ...string) (err error) {
	w.query = w.buildSelect(columns...)
	_, err = w.runQuery()
	if err != nil {
		return
	}
	return
}

// GetOne 會取得僅單列的資料作為結果，傳入的參數為欲取得的欄位名稱，不傳入參數表示取得所有欄位。
// 簡單說，這就是 `.Limit(1).Get()` 的縮寫用法。
func (w *Wrapper) GetOne(columns ...string) (err error) {
	w.Limit(1)
	w.query = w.buildSelect(columns...)
	_, err = w.runQuery()
	if err != nil {
		return
	}
	return
}

// Paginate 基本上和 `Get` 取得函式無異，但此函式能夠自動依照分頁數來推算該從哪裡繼續取得資料。
// 使用時須先確定是否有指定 `PageLimit`（預設為：20），這樣才能限制一頁有多少筆資料。
func (w *Wrapper) Paginate(pageCount int, columns ...string) (err error) {
	err = w.WithTotalCount().Limit(w.PageLimit*(pageCount-1), w.PageLimit).Get(columns...)
	w.TotalPage = w.TotalCount / w.PageLimit
	return
}

// WithTotalCount 會在 SQL 執行指令中安插 `SQL_CALC_FOUND_ROWS` 選項，
// 如此一來就能夠在執行完 SQL 指令後取得查詢的總計行數。在不同情況下，這可能會拖低執行效能。
func (w *Wrapper) WithTotalCount() *Wrapper {
	w.SetQueryOption("SQL_CALC_FOUND_ROWS")
	return w
}

//=======================================================
// 插入函式
//=======================================================

// Insert 會插入一筆新的資料。
func (w *Wrapper) Insert(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	res, err := w.executeQuery()
	if err != nil || !w.executable {
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		return
	}
	w.LastInsertID = int(id)
	return
}

// InsertMulti 會一次插入多筆資料。
func (w *Wrapper) InsertMulti(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	res, err := w.executeQuery()
	if err != nil || !w.executable {
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		return
	}
	w.LastInsertID = int(id)
	return
}

// Delete 會移除相符的資料列，記得用上 `Where` 條件式來避免整個資料表格被清空。
// 這很重要好嗎，因為⋯你懂的⋯。喔，不。
func (w *Wrapper) Delete() (err error) {
	w.query = w.buildDelete(w.tableName...)
	_, err = w.executeQuery()
	return
}

//=======================================================
// 更新函式
//=======================================================

// Replace 基本上和 `Insert` 無異，這會在有重複資料時移除該筆資料並重新插入。
// 若無該筆資料則插入新的資料。
func (w *Wrapper) Replace(data interface{}) (err error) {
	w.query = w.buildInsert("REPLACE", data)
	_, err = w.executeQuery()
	return
}

// Update 會以指定的資料來更新相對應的資料列。
func (w *Wrapper) Update(data interface{}) (err error) {
	w.query = w.buildUpdate(data)
	_, err = w.executeQuery()
	return
}

// OnDuplicate 能夠指定欲更新的欄位名稱，這會在插入的資料重複時自動更新相對應的欄位。
func (w *Wrapper) OnDuplicate(columns []string, lastInsertID ...string) *Wrapper {
	w.onDuplicateColumns = columns
	if len(lastInsertID) != 0 {
		w.lastInsertIDColumn = lastInsertID[0]
	}
	return w
}

//=======================================================
// 限制函式
//=======================================================

// Limit 能夠在 SQL 查詢指令中建立限制筆數的條件。
func (w *Wrapper) Limit(from int, count ...int) *Wrapper {
	if len(count) == 0 {
		w.limit = []int{from}
	} else {
		w.limit = []int{from, count[0]}
	}
	return w
}

// OrderBy 會依照指定的欄位來替結果做出排序（例如：`DESC`、`ASC`）。
func (w *Wrapper) OrderBy(column string, args ...interface{}) *Wrapper {
	w.orders = append(w.orders, order{
		column: column,
		args:   args,
	})
	return w
}

// GroupBy 會在執行 SQL 指令時依照特定的欄位來做執行區分。
func (w *Wrapper) GroupBy(columns ...string) *Wrapper {
	w.groupBy = columns
	return w
}

//=======================================================
// 指令函式
//=======================================================

// RawQuery 會接收傳入的變數來執行傳入的 SQL 執行語句，
// 變數可以在語句中以 `?`（Prepared Statements）使用來避免 SQL 注入攻擊。
func (w *Wrapper) RawQuery(query string, values ...interface{}) (err error) {
	w.query = query
	w.params = values
	_, err = w.runQuery()
	return
}

//=======================================================
// 條件函式
//=======================================================

// Where 會增加一個 `WHERE AND` 條件式。
func (w *Wrapper) Where(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "AND", args...)
	return w
}

// OrWhere 會增加一個 `WHERE OR` 條件式。
func (w *Wrapper) OrWhere(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "OR", args...)
	return w
}

// Having 會增加一個 `HAVING AND` 條件式。
func (w *Wrapper) Having(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "AND", args...)
	return w
}

// OrHaving 會增加一個 `HAVING OR` 條件式。
func (w *Wrapper) OrHaving(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "OR", args...)
	return w
}

//=======================================================
// 加入函式
//=======================================================

// LeftJoin 會向左插入一個資料表格。
func (w *Wrapper) LeftJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "LEFT JOIN", condition)
	return w
}

// RightJoin 會向右插入一個資料表格。
func (w *Wrapper) RightJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "RIGHT JOIN", condition)
	return w
}

// InnerJoin 會內部插入一個資料表格。
func (w *Wrapper) InnerJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "INNER JOIN", condition)
	return w
}

// NaturalJoin 會自然插入一個資料表格。
func (w *Wrapper) NaturalJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "NATURAL JOIN", condition)
	return w
}

// JoinWhere 能夠建立一個基於 `WHERE AND` 的條件式給某個指定的插入資料表格。
func (w *Wrapper) JoinWhere(table interface{}, args ...interface{}) *Wrapper {
	w.saveJoinCondition("AND", table, args...)
	return w
}

// JoinOrWhere 能夠建立一個基於 `WHERE OR` 的條件式給某個指定的插入資料表格。
func (w *Wrapper) JoinOrWhere(table interface{}, args ...interface{}) *Wrapper {
	w.saveJoinCondition("OR", table, args...)
	return w
}

// SubQuery 能夠將目前的 SQL 指令轉換為子指令（Sub Query）來防止建置後直接被執行，這讓你可以將子指令傳入其他的條件式（例如：`WHERE`），
// 若欲將子指令傳入插入（Join）條件中，必須在參數指定此子指令的別名。
func (w *Wrapper) SubQuery(alias ...string) *Wrapper {
	newWrapper := &Wrapper{
		executable: false,
	}
	if len(alias) > 0 {
		newWrapper.alias = alias[0]
	}
	return newWrapper
}

// Has 會在有查詢結果時回傳 `true`，這很適合用於一些資料驗證的時機（例如：使用者名稱是否已存在⋯等）。
func (w *Wrapper) Has() (has bool, err error) {
	err = w.Limit(1).Get()
	if err != nil {
		has = false
		return
	}
	if w.Count() > 0 {
		has = true
		return
	}
	return
}

//=======================================================
// 資料庫函式
//=======================================================

// Disconnect 會結束目前的資料庫連線。
func (w *Wrapper) Disconnect() (err error) {
	err = w.db.Disconnect()
	return
}

// Ping 會以 ping 來檢查資料庫連線。
func (w *Wrapper) Ping() (err error) {
	err = w.db.Ping()
	return
}

// Connect 會試圖在斷線之後重新連線至資料庫。
func (w *Wrapper) Connect() (err error) {
	err = w.db.Connect()
	return
}

//=======================================================
// 交易函式
//=======================================================

// Begin 會開始一個新的交易。
func (w *Wrapper) Begin() (*Wrapper, error) {
	anotherDB := *w.db
	tx, err := anotherDB.Begin()
	if err != nil {
		return w, err
	}
	anotherMaster := *w.db.master
	anotherDB.master = &anotherMaster
	anotherDB.master.tx = tx
	return w.cloning(false, &anotherDB), nil
}

// Rollback 能夠回溯到交易剛開始的時候，並且在不保存資料變動的情況下結束交易。
func (w *Wrapper) Rollback() error {
	return w.db.Rollback()
}

// Commit 會讓交易中所產生的資料異動成為永久紀錄並保存於資料庫中且結束交易。
func (w *Wrapper) Commit() error {
	return w.db.Commit()
}

//=======================================================
// 輔助函式
//=======================================================

// Count 會回傳查詢的結果行數。
func (w *Wrapper) Count() (count int) {
	count = w.count
	return
}

// Func 會基於參數來返回一個新的 SQL 資料庫函式，
// 這能夠當作函式放置於查詢指令中，而不會被當作普通的資料執行。
func (w *Wrapper) Func(query string, data ...interface{}) Function {
	return Function{
		query:  query,
		values: data,
	}
}

// Now 會回傳一個基於 `INTERVAL` 的 SQL 資料庫函式，
// 傳入的參數格式可以是 `+1Y`、`-2M`，同時也可以像 `Now("+1Y", "-2M")` 一樣地串連使用。
// 支援的格式為：`Y`(年)、`M`(月)、`D`(日)、`W`(星期)、`h`(小時)、`m`(分鐘)、`s`(秒數)。
func (w *Wrapper) Now(formats ...string) Function {
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
	return w.Func(strings.TrimSpace(query))
}

// SetLockMethod 會設置鎖定資料表格的方式（例如：`WRITE`、`READ`）。
func (w *Wrapper) SetLockMethod(method string) *Wrapper {
	w.lockMethod = strings.ToUpper(method)
	return w
}

// Lock 會以指定的上鎖方式來鎖定某個指定的資料表格，這能用以避免資料競爭問題。
func (w *Wrapper) Lock(tableNames ...string) (err error) {
	var tables string
	for _, v := range tableNames {
		tables += fmt.Sprintf("%s %s, ", v, w.lockMethod)
	}
	tables = trim(tables)

	err = w.RawQuery(fmt.Sprintf("LOCK TABLES %s", tables))
	return
}

// Unlock 能解鎖已鎖上的資料表格。
func (w *Wrapper) Unlock(tableNames ...string) (err error) {
	err = w.RawQuery("UNLOCK TABLES")
	return
}

// SetQueryOption 會設置 SQL 指令的額外選項（例如：`SQL_NO_CACHE`）。
func (w *Wrapper) SetQueryOption(options ...string) *Wrapper {
	w.queryOptions = append(w.queryOptions, options...)
	return w
}

// SetTrace 會決定蹤跡模式的開關，當設置為 `true` 時會稍微地拖慢效能，
// 但你就能夠從 `Trace` 屬性中取得 SQL 執行後的堆疊與路徑結果。
func (w *Wrapper) SetTrace(status bool) *Wrapper {
	w.tracing = status
	return w
}

//=======================================================
// 物件函式
//=======================================================

// Clone 會複製一個的 SQL 指令建置建構體，但不包括已設置的資料與指令。
func (w *Wrapper) Clone() *Wrapper {
	return w.cloning(false)
}

// Copy 會複製目前的 SQL 指令建置建構體與已設置的資料與指令。
func (w *Wrapper) Copy() *Wrapper {
	return w.cloning(true)
}

// Bind 會設置資料的映射目的地，這樣就能在 SQL 指令執行後將資料映射到某個變數、記憶體指標。
func (w *Wrapper) Bind(destination interface{}) *Wrapper {
	w.destination = destination
	return w
}

// Query 會回傳最後一次所建置的 SQL 執行指令，這和 `LastQuery` 相同功能但這函式的名稱可能更符合某些場合。
func (w *Wrapper) Query() (query string) {
	query = w.LastQuery
	return
}

// Params returns the last used parameters, it's the same as `LastParams` but the name is more meaningful.
func (w *Wrapper) Params() (params []interface{}) {
	params = w.LastParams
	return
}

//=======================================================
// 其他函式
//=======================================================

// Migration 會返回一個新的資料表格遷移建構體。
// 主要是基於現有的資料庫連線來提供資料表格與欄位的的操作功能。
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
