package reiner

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

var (
	// ErrInvalidPointer occurred when the pointer of the destination is invalid.
	ErrInvalidPointer = errors.New("reiner: The destination of the result is an invalid pointer")
	// ErrIncorrectDataType occurred when the insert, update data is not a `map[string]interface` type.
	ErrIncorrectDataType = errors.New("reiner: The data type must be a `map[string]interface`")
	// ErrUnbegunTransaction occurred when the transation function was called before calling `Begin()`.
	ErrUnbegunTransaction = errors.New("reiner: Calling the transaction function without `Begin()`")
	// ErrNoTable occurred when the table wasn't specified yet.
	ErrNoTable = errors.New("reiner: No table was specified")
)

// Function represents a database function like `SHA(?)` or `NOW()`.
type Function struct {
	query  string
	values []interface{}
}

// condition represents a WHERE or HAVING condition.
type condition struct {
	args      []interface{}
	connector string
}

// order represents a ORDER statement.
type order struct {
	column string
	args   []interface{}
}

// join represents a table joining infromation.
type join struct {
	typ        string
	table      interface{}
	condition  string
	conditions []condition
}

// Trace represents a trace of the executed query.
type Trace struct {
	Query    string
	Duration time.Duration
	Stacks   []map[string]interface{}
	Error    error
}

// Wrapper represents a database query wrapper, which contains the database connection.
type Wrapper struct {
	db *DB
	// executable executes the built queries if it's true.
	executable bool
	// alias is the alias for the table when joining the table as a sub query.
	alias string
	// destination represents a pointer to the destination of the result.
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

	// Timestamp is the timestamp tool.
	Timestamp *Timestamp
	// Traces represents the trace of each executed query.
	Traces []Trace
	//
	TotalCount int
	// PageLimit limits the amount of the rows can be feteched in a page.
	PageLimit int
	// TotalPage represents the count of total page.
	TotalPage int
	// LasyQuery is last executed query.
	LastQuery string
	// LastInsertID is the last insert ID.
	LastInsertID int
	//
	LastParams []interface{}
	//
	LastResult sql.Result
}

// newWrapper creates a new database function wrapper by the passed database connection.
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{executable: true, db: db, Timestamp: &Timestamp{}, joins: make(map[string]*join)}
}

// cloning clones the database wrapper.
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

// cleanAfter cleans the last executed result after the new query was executed.
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

// cleanBefore cleans the last executed result before the new query was executed.
func (w *Wrapper) cleanBefore() {
	w.TotalCount = 0
	w.TotalPage = 0
	w.LastInsertID = 0
	w.LastResult = nil
	w.LastParams = []interface{}{}
	w.count = 0
}

//=======================================================
// Save Functions
//=======================================================

// saveTrace gets the callers and calculates the execution time then save the tracing information.
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

// saveJoin saves the table joining information.
func (w *Wrapper) saveJoin(table interface{}, typ string, condition string) {
	switch v := table.(type) {
	// Sub query joining.
	case *Wrapper:
		w.joins[v.query] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	// Common table joining.
	case string:
		w.joins[v] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	}
}

// saveJoinCondition saves a table join condition to the specifed table joining information.
func (w *Wrapper) saveJoinCondition(connector string, table interface{}, args ...interface{}) {
	switch v := table.(type) {
	// Sub query condition.
	case *Wrapper:
		w.joins[v.query].conditions = append(w.joins[v.query].conditions, condition{
			args:      args,
			connector: connector,
		})
	// Common condition.
	case string:
		w.joins[v].conditions = append(w.joins[v].conditions, condition{
			args:      args,
			connector: connector,
		})
	}
}

// saveCondition stores a condition to the right place.
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
// Param Functions
//=======================================================

// bindParams binds the values of the interface slice to the database, and generates the prepared statement.
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

// bindParam binds the single value to the database and generates the prepared statement.
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

// paramToQuery converts the value to the prepare statement.
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
// Build Functions
//=======================================================

// buildWhere builds the `WHERE` or the `HAVING` statement based on the stored conditons.
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

// buildUpdate builds the `UPDATE` query.
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

// buildLimit builds the `LIMIT` statement.
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

// buildSelect builds the `SELECT` query.
func (w *Wrapper) buildSelect(columns ...string) (query string) {
	beforeOptions, _ := w.buildQueryOptions()

	if len(columns) == 0 {
		query = fmt.Sprintf("SELECT %s* FROM %s ", beforeOptions, w.tableName[0])
	} else {
		query = fmt.Sprintf("SELECT %s%s FROM %s ", beforeOptions, strings.Join(columns, ", "), w.tableName[0])
	}
	return
}

// buildConditions cooks the conditions to the `WHERE` or the `HAVING` statement.
func (w *Wrapper) buildConditions(conditions []condition) (query string) {
	for i, v := range conditions {
		// Add the connector if it's not the first condition.
		if i != 0 {
			query += fmt.Sprintf("%s ", v.connector)
		}

		// Get the type of the column name, it might be a query, or normal column name, or even a sub query.
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

		// Build the condition based on the type.
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

// buildDelete builds and executes the delete query.
func (w *Wrapper) buildDelete(tableNames ...string) (query string) {
	beforeOptions, _ := w.buildQueryOptions()
	query += fmt.Sprintf("DELETE %sFROM %s ", beforeOptions, strings.Join(tableNames, ", "))
	return
}

// buildQueryOptions builds the query options and
// return the two type of the options which is in the start of the query, another is in the end of the query.
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

// buildQuery builds the whole query.
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

// buildOrderBy builds the `ORDERY BY` statement based on the stored orders.
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

// buildGroupBy builds the `GROUP BY` statement.
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

// buildDuplicate builds the `ON DUPLICATE KEY UPDATE` statement.
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

// buildInsert builds the `INSERT INTO` query.
func (w *Wrapper) buildInsert(operator string, data interface{}) (query string) {
	var columns, values string
	beforeOptions, _ := w.buildQueryOptions()

	// Build the query based on the data type.
	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			columns += fmt.Sprintf("%s, ", column)
			values += fmt.Sprintf("%s, ", w.bindParam(value))
		}
		values = fmt.Sprintf("(%s)", trim(values))

	case []map[string]interface{}:
		var columnNames []string
		// Get the column names first, so we can range the map in order.
		for name := range realData[0] {
			columnNames = append(columnNames, name)
			// Build the column name query.
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

// buildJoin builds the join statement.
func (w *Wrapper) buildJoin() (query string) {
	if len(w.joins) == 0 {
		return
	}

	for _, v := range w.joins {
		// The join type (ex: LEFT JOIN, RIGHT JOIN, INNER JOIN).
		query += fmt.Sprintf("%s ", v.typ)
		switch d := v.table.(type) {
		// Sub query.
		case *Wrapper:
			query += fmt.Sprintf("%s AS %s ON ", w.bindParam(d), d.alias)
		// Table name.
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
// Execution.
//=======================================================

// runQuery runs the query with `Query`.
func (w *Wrapper) runQuery() (rows *sql.Rows, err error) {
	w.cleanBefore()
	w.buildQuery()
	w.LastQuery = w.query
	w.LastParams = w.params

	// Calculate the execution time.
	var start time.Time
	if w.tracing {
		start = time.Now()
	}

	// Execute the query if the wrapper is executable.
	if w.executable {
		var stmt *sql.Stmt
		var count int
		var tx *sql.Tx

		// Using a transaction if there's `SQL_CALC_FOUND_ROWS` in the query option,
		// because it requires the same connection.
		for _, v := range w.queryOptions {
			if v == "SQL_CALC_FOUND_ROWS" {
				// Start the transaction.
				tx, err = w.db.Begin()
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// Prepare the query.
				stmt, err = tx.Prepare(w.query)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// Execute the query.
				rows, err = stmt.Query(w.params...)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// Loads the results to the destination.
				// This also closes the `rows` so it won't trigger the `busy buffer` error.
				count, err = load(rows, w.destination)
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// Save the count.
				w.count = count

				// Select the `FOUND_ROWS` to get the total count.
				rows, err = tx.Query("SELECT FOUND_ROWS()")
				if err != nil {
					w.saveTrace(err, w.query, start)
					w.cleanAfter()
					return
				}
				// Scan the result to get the total count.
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
				// Close the whole statement.
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

		// If there're no `SQL_CALC_FOUND_ROWS` setted,
		// we use the normal connection pool.
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

// executeQuery executes the query with `Exec`.
func (w *Wrapper) executeQuery() (res sql.Result, err error) {
	w.cleanBefore()
	w.buildQuery()
	w.LastQuery = w.query
	w.LastParams = w.params

	// Calculate the execution time.
	var start time.Time
	if w.tracing {
		start = time.Now()
	}

	// Execute the query if the wrapper is executable.
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
// Exported Functions
//=======================================================

// Table specifies the name of the table.
func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

//=======================================================
// Select Functions
//=======================================================

// Get gets the specified columns of the rows from the specifed database table.
func (w *Wrapper) Get(columns ...string) (err error) {
	w.query = w.buildSelect(columns...)
	_, err = w.runQuery()
	if err != nil {
		return
	}
	return
}

// GetOne gets the specified columns with the only one row from the specifed database table,
// It's an alias for `.Limit(1).Get()`.
func (w *Wrapper) GetOne(columns ...string) (err error) {
	w.Limit(1)
	w.query = w.buildSelect(columns...)
	_, err = w.runQuery()
	if err != nil {
		return
	}
	return
}

// Paginate acts the same as `Get` but with the automatically page caculation.
// Make sure you have specified the `PageLimit` (Default: 20) to limit the rows of a page.
func (w *Wrapper) Paginate(pageCount int, columns ...string) (err error) {
	err = w.WithTotalCount().Limit(w.PageLimit*(pageCount-1), w.PageLimit).Get(columns...)
	w.TotalPage = w.TotalCount / w.PageLimit
	return
}

// WithTotalCount sets the `SQL_CALC_FOUND_ROWS` query option
// so you can get the total count of the rows after the query was executed.
// This might reduce the executing performance.
func (w *Wrapper) WithTotalCount() *Wrapper {
	w.SetQueryOption("SQL_CALC_FOUND_ROWS")
	return w
}

//=======================================================
// Insert Functions
//=======================================================

// Insert builds and executes the insert query.
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

// InsertMulti builds and executes a single insert query with the many rows.
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

// Delete deletes the row(s), use it with the `Where` condition so your whole table won't be wiped.
// It's very important alright? Cause .. you know ..fuck.
func (w *Wrapper) Delete() (err error) {
	w.query = w.buildDelete(w.tableName...)
	_, err = w.executeQuery()
	return
}

//=======================================================
// Update Functions
//=======================================================

// Replace builds and executes the replace query just like what `Insert` does.
func (w *Wrapper) Replace(data interface{}) (err error) {
	w.query = w.buildInsert("REPLACE", data)
	_, err = w.executeQuery()
	return
}

// Update updates the rows with the specified data.
func (w *Wrapper) Update(data interface{}) (err error) {
	w.query = w.buildUpdate(data)
	_, err = w.executeQuery()
	return
}

// OnDuplicate stores the columns which would be updated when the inserted row has duplicated.
func (w *Wrapper) OnDuplicate(columns []string, lastInsertID ...string) *Wrapper {
	w.onDuplicateColumns = columns
	if len(lastInsertID) != 0 {
		w.lastInsertIDColumn = lastInsertID[0]
	}
	return w
}

//=======================================================
// Limitation Functions
//=======================================================

// Limit creates the limit statement in the end of the query.
func (w *Wrapper) Limit(from int, count ...int) *Wrapper {
	if len(count) == 0 {
		w.limit = []int{from}
	} else {
		w.limit = []int{from, count[0]}
	}
	return w
}

// OrderBy orders the getting result based on a single column (or the fields) with the specified sorting like `ASC` and `DESC`.
func (w *Wrapper) OrderBy(column string, args ...interface{}) *Wrapper {
	w.orders = append(w.orders, order{
		column: column,
		args:   args,
	})
	return w
}

// GroupBy groups the columns when executing the query.
func (w *Wrapper) GroupBy(columns ...string) *Wrapper {
	w.groupBy = columns
	return w
}

//=======================================================
// Query Functions
//=======================================================

// RawQuery executes the passed raw query and binds the passed values to the prepared statments.
func (w *Wrapper) RawQuery(query string, values ...interface{}) (err error) {
	w.query = query
	w.params = values
	_, err = w.runQuery()
	return
}

//=======================================================
// Condition Functions
//=======================================================

// Where adds a new `WHERE AND` condition.
func (w *Wrapper) Where(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "AND", args...)
	return w
}

// OrWhere adds a new `WHERE OR` condition.
func (w *Wrapper) OrWhere(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "OR", args...)
	return w
}

// Having adds a new `HAVING AND` condition.
func (w *Wrapper) Having(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "AND", args...)
	return w
}

// OrHaving adds a new `HAVING OR` condition.
func (w *Wrapper) OrHaving(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "OR", args...)
	return w
}

//=======================================================
// Join Functions
//=======================================================

// LeftJoin left joins a table.
func (w *Wrapper) LeftJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "LEFT JOIN", condition)
	return w
}

// RightJoin right joins a table.
func (w *Wrapper) RightJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "RIGHT JOIN", condition)
	return w
}

// InnerJoin inner joins a table.
func (w *Wrapper) InnerJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "INNER JOIN", condition)
	return w
}

// NaturalJoin natural joins a table.
func (w *Wrapper) NaturalJoin(table interface{}, condition string) *Wrapper {
	w.saveJoin(table, "NATURAL JOIN", condition)
	return w
}

// JoinWhere creates a `WHERE AND` statement for the specified joining table.
func (w *Wrapper) JoinWhere(table interface{}, args ...interface{}) *Wrapper {
	w.saveJoinCondition("AND", table, args...)
	return w
}

// JoinOrWhere creates a `WHERE OR` statement for the specified joining table.
func (w *Wrapper) JoinOrWhere(table interface{}, args ...interface{}) *Wrapper {
	w.saveJoinCondition("OR", table, args...)
	return w
}

// SubQuery converts the current query into a sub query so it won't be executed, and it could be passed to the any other statement like `Where`, the alias is required when passing the sub query to the joining statement.
func (w *Wrapper) SubQuery(alias ...string) *Wrapper {
	newWrapper := &Wrapper{
		executable: false,
	}
	if len(alias) > 0 {
		newWrapper.alias = alias[0]
	}
	return newWrapper
}

// Has returns true when there's a such result, it's useful when you want to validate if there's a duplicated email address, or the usename.
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
// Database Functions
//=======================================================

// Disconnect disconnects the current database conection.
func (w *Wrapper) Disconnect() (err error) {
	err = w.db.Disconnect()
	return
}

// Ping pings the database to make sure it's working.
func (w *Wrapper) Ping() (err error) {
	err = w.db.Ping()
	return
}

// Connect reconnects to the database after the connection has disconnected.
func (w *Wrapper) Connect() (err error) {
	err = w.db.Connect()
	return
}

//=======================================================
// Transactions
//=======================================================

// Begin starts a transcation.
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

// Rollback rolls the changes back to where the transaction started.
func (w *Wrapper) Rollback() error {
	return w.db.Rollback()
}

// Commit commits the current transaction, so the changes will be saved into the database permanently.
func (w *Wrapper) Commit() error {
	return w.db.Commit()
}

//=======================================================
// Helper Functions
//=======================================================

// Count returns the count of the result rows.
func (w *Wrapper) Count() (count int) {
	count = w.count
	return
}

// Func returns a database function, so it won't be treated like a normal data value.
func (w *Wrapper) Func(query string, data ...interface{}) Function {
	return Function{
		query:  query,
		values: data,
	}
}

// Now returns a database function based on the `INTERVAL`.
// The formats can be like `+1Y`, `-2M`, it's possible to chain the units with `Now("+1Y", "-2M")`.
// Here're the supported units: `Y`(Year), `M`(Month), `D`(Day), `W`(Week), `h`(Hour), `m`(Minute), `s`(Second).
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

// SetLockMethod sets the lock method before locked the tables, it could be `WRITE` or `READ`.
func (w *Wrapper) SetLockMethod(method string) *Wrapper {
	w.lockMethod = strings.ToUpper(method)
	return w
}

// Lock locks the tables with the specified lock method.
func (w *Wrapper) Lock(tableNames ...string) (err error) {
	var tables string
	for _, v := range tableNames {
		tables += fmt.Sprintf("%s %s, ", v, w.lockMethod)
	}
	tables = trim(tables)

	err = w.RawQuery(fmt.Sprintf("LOCK TABLES %s", tables))
	return
}

// Unlock unlocks the locked tables.
func (w *Wrapper) Unlock(tableNames ...string) (err error) {
	err = w.RawQuery("UNLOCK TABLES")
	return
}

// SetQueryOption sets the query options like `SQL_NO_CACHE`.
func (w *Wrapper) SetQueryOption(options ...string) *Wrapper {
	w.queryOptions = append(w.queryOptions, options...)
	return w
}

// SetTrace sets the trace mode as on or off, so you can get the traces by accessing the `Trace` property.
func (w *Wrapper) SetTrace(status bool) *Wrapper {
	w.tracing = status
	return w
}

//=======================================================
// Object Functions
//=======================================================

// Clone clones the current database wrapper without the same settings.
func (w *Wrapper) Clone() *Wrapper {
	return w.cloning(false)
}

// Copy copies the current database wrapper with the same settings.
func (w *Wrapper) Copy() *Wrapper {
	return w.cloning(true)
}

// Bind binds the destination of the result.
func (w *Wrapper) Bind(destination interface{}) *Wrapper {
	w.destination = destination
	return w
}

// Query returns the last built query, it's the same as `LastQuery` but the name is more meaningful.
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
// Others
//=======================================================

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
