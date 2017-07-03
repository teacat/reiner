package reiner

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

const (
//Err
)

// function represents a database function like `SHA(?)` or `NOW()`.
type function struct {
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
	Stacks   []string
}

// Wrapper represents a database query wrapper, which contains the database connection.
type Wrapper struct {
	db *DB
	// executable executes the built queries if it's true.
	executable bool
	// builderMode enables when the wrapper is used to build the queries,
	// the queries won't be executed in this mode.
	builderMode bool
	// isSubQuery tells if the wrapper is a sub query or not.
	// The query won't be executed if it's a sub query.
	isSubQuery bool
	// alias is the alias for the table when joining the table as a sub query.
	alias string
	// destination represents a pointer to the destination of the result.
	destination interface{}
	// scanner handles the rows scan function after the result was fetched.
	scanner            func(*sql.Rows)
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

	// Params represents the parameter values of the query.
	Params []interface{}
	// Query represents the builded query.
	Query string
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
	// LastRows is the `*sql.Rows` from the last result.
	LastRows *sql.Rows
	//
	LastResult sql.Result
}

// newWrapper creates a new database function wrapper by the passed database connection.
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{db: db, Timestamp: &Timestamp{}}
}

// clean cleans the last executed result.
func (w *Wrapper) clean() {
	w.LastInsertID = 0
	w.LastResult = nil
	w.queryOptions = []string{}
	w.LastRows = nil
	w.tableName = []string{}
	w.Params = []interface{}{}
	w.onDuplicateColumns = []string{}
	w.groupBy = []string{}
	w.joins = map[string]*join{}
	w.orders = []order{}
	w.conditions = []condition{}
	w.havingConditions = []condition{}
	w.limit = []int{}
	w.destination = nil
	w.scanner = nil
}

//=======================================================
// Save Functions
//=======================================================

// saveJoin saves the table joining information.
func (w *Wrapper) saveJoin(table interface{}, typ string, condition string) {
	switch v := table.(type) {
	// Sub query joining.
	case *Wrapper:
		w.joins[v.Query] = &join{
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
		w.joins[v.Query].conditions = append(w.joins[v.Query].conditions, condition{
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
		if len(v.Params) > 0 {
			w.Params = append(w.Params, v.Params...)
		}
	case function:
		if len(v.values) > 0 {
			w.Params = append(w.Params, v.values...)
		}
	case Timestamp:
		w.Params = append(w.Params, v.value)
	default:
		w.Params = append(w.Params, data)
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
				param = fmt.Sprintf("%s", v.Query)
			}
		} else {
			param = fmt.Sprintf("(%s)", v.Query)
		}
	case function:
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
	query = fmt.Sprintf("UPDATE %s SET ", w.tableName[0])
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
	if len(columns) == 0 {
		query = fmt.Sprintf("SELECT * FROM %s ", w.tableName[0])
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s ", strings.Join(columns, ", "), w.tableName[0])
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
	query += fmt.Sprintf("DELETE FROM %s ", strings.Join(tableNames, ", "))
	return
}

// buildQuery builds the whole query.
func (w *Wrapper) buildQuery() {
	w.Query += w.buildDuplicate()
	w.Query += w.buildJoin()
	w.Query += w.buildWhere("WHERE")
	w.Query += w.buildWhere("HAVING")
	w.Query += w.buildOrderBy()
	w.Query += w.buildGroupBy()
	w.Query += w.buildLimit()
	w.Query = strings.TrimSpace(w.Query)
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
	query = trim(query)
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
	query = trim(query)
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
		query += fmt.Sprintf("%s = VALUE(%s), ", v, v)
	}
	query = trim(query)
	return
}

// buildInsert builds the `INSERT INTO` query.
func (w *Wrapper) buildInsert(operator string, data interface{}) (query string) {
	var columns, values, options string
	if len(w.queryOptions) > 0 {
		options = fmt.Sprintf("%s ", strings.Join(w.queryOptions, ", "))
	}

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
	query = fmt.Sprintf("%s %sINTO %s (%s) VALUES %s ", operator, options, w.tableName[0], columns, values)
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

func (w *Wrapper) runQuery() (rows *sql.Rows, err error) {
	w.buildQuery()
	stmt, err := w.db.Prepare(w.Query)
	if err != nil {
		return nil, err
	}
	rows, err = stmt.Query(w.Params...)
	if err != nil {
		return nil, err
	}
	_, err = load(rows, w.destination)
	if err != nil {
		return nil, err
	}
	w.LastRows = rows
	w.LastQuery = w.Query
	if !w.isSubQuery {
		w.clean()
	}
	return
}

func (w *Wrapper) executeQuery() (res sql.Result, err error) {
	w.buildQuery()
	stmt, err := w.db.Prepare(w.Query)
	if err != nil {
		return
	}
	res, err = stmt.Exec(w.Params...)
	if err != nil {
		return
	}
	w.LastResult = res
	w.LastQuery = w.Query
	if !w.isSubQuery {
		w.clean()
	}
	return
}

// Get gets the specified columns of the rows from the specifed database table.
func (w *Wrapper) Get(columns ...string) (err error) {
	w.Query = w.buildSelect(columns...)
	_, err = w.runQuery()
	return
}

// GetOne gets the specified columns of a single row from the specifed database table.
func (w *Wrapper) GetOne(columns ...string) (err error) {
	w.Limit(1)
	w.Query = w.buildSelect(columns...)
	_, err = w.runQuery()
	return
}

// GetValue gets the value of the specified column of the rows, you'll get the slice of the values if you didn't specify `LIMIT 1`.
func (w *Wrapper) GetValue(column string) (err error) {
	w.Query = w.buildSelect(column)
	_, err = w.runQuery()
	return
}

// Paginate acts the same as `Get` but with the automatically page caculation.
// Make sure you have specified the `PageLimit` (Default: 20) to limit the rows of a page.
func (w *Wrapper) Paginate(pageCount int, columns ...string) (err error) {
	err = w.Limit(w.PageLimit*(pageCount-1), w.PageLimit).Get(columns...)
	w.TotalPage = w.TotalCount / w.PageLimit
	return
}

//=======================================================
// Insert Functions
//=======================================================

// Insert builds and executes the insert query.
func (w *Wrapper) Insert(data interface{}) (err error) {
	w.Query = w.buildInsert("INSERT", data)
	res, err := w.executeQuery()
	if err != nil {
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
	w.Query = w.buildInsert("INSERT", data)
	res, err := w.executeQuery()
	if err != nil {
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
	w.Query = w.buildDelete(w.tableName...)
	_, err = w.executeQuery()
	return
}

//=======================================================
// Update Functions
//=======================================================

// Replace builds and executes the replace query just like what `Insert` does.
func (w *Wrapper) Replace(data interface{}) (err error) {
	w.Query = w.buildInsert("REPLACE", data)
	_, err = w.executeQuery()
	return
}

// Update updates the rows with the specified data.
func (w *Wrapper) Update(data interface{}) (err error) {
	w.Query = w.buildUpdate(data)
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
	w.Query = query
	w.Params = values
	_, err = w.runQuery()
	return
}

// RawQueryOne works the same as `GetOne`, and it only gets a single row as the result.
func (w *Wrapper) RawQueryOne(query string, values ...interface{}) (err error) {
	err = w.RawQuery(query, values...)
	return
}

// RawQueryValue works the same as `GetValue`, it gets the value slice instead of a single value if there's no `LIMIT 1` was specifed in the raw query.
func (w *Wrapper) RawQueryValue(query string, values ...interface{}) (err error) {
	err = w.RawQuery(query, values...)
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
		isSubQuery: true,
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
	return
}

// Ping pings the database to make sure it's working.
func (w *Wrapper) Ping() (err error) {
	return
}

// Connect reconnects to the database after the connection has disconnected.
func (w *Wrapper) Connect() (err error) {
	return
}

//=======================================================
// Transactions
//=======================================================

// Begin starts a transcation.
func (w *Wrapper) Begin() (tx *Wrapper, err error) {

	return w, nil
}

// Rollback rolls the changes back to where the transaction started.
func (w *Wrapper) Rollback() bool {

	return false
}

// Commit commits the current transaction, so the changes will be saved into the database permanently.
func (w *Wrapper) Commit() error {
	return nil
}

//=======================================================
// Helper Functions
//=======================================================

// Count returns the count of the result rows.
func (w *Wrapper) Count() (count int) {
	if w.LastRows != nil {
		for w.LastRows.Next() {
			count++
		}
	} else if w.LastResult != nil {
		rowAffected, err := w.LastResult.RowsAffected()
		if err != nil {
			count = 0
		} else {
			count = int(rowAffected)
		}
	} else {
		count = 0
	}
	return
}

// Func returns a database function, so it won't be treated like a normal data value.
func (w *Wrapper) Func(query string, data ...interface{}) function {
	return function{
		query:  query,
		values: data,
	}
}

// Now returns a database function based on the `INTERVAL`.
// The formats can be like `+1Y`, `-2M`, it's possible to chain the units with `Now("+1Y", "-2M")`.
// Here're the supported units: `Y`(Year), `M`(Month), `D`(Day), `W`(Week), `h`(Hour), `m`(Minute), `s`(Second).
func (w *Wrapper) Now(formats ...string) function {
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
	w.queryOptions = options
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

// Copy returns a new database wrapper based on the current configurations. It's useful when you're trying to pass the database wrapper to the goroutines to make sure it's thread safe.
func (w *Wrapper) Copy() *Wrapper {
	return w
}

// Scan scans the rows of the result, and mapping it to the specified variable.
func (w *Wrapper) Scan(handler func(*sql.Rows)) *Wrapper {
	w.scanner = handler
	return w
}

// Bind binds the destination of the result.
func (w *Wrapper) Bind(destination interface{}) *Wrapper {
	w.destination = destination
	return w
}

//=======================================================
// Others
//=======================================================

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
