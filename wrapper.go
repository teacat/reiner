package reiner

import (
	"database/sql"
	"fmt"
	"strings"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
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

// Wrapper represents a database query wrapper, which contains the database connection.
type Wrapper struct {
	db                 *DB
	isSubQuery         bool
	query              string
	alias              string
	tableName          []string
	conditions         []condition
	havingConditions   []condition
	queryOptions       []string
	destination        interface{}
	joins              map[string]*join
	params             []interface{}
	onDuplicateColumns []string
	lastInsertIDColumn string
	limit              []int
	orders             []order
	groupBy            []string

	// Timestamp is the timestamp tool.
	Timestamp *Timestamp
	// Count is the count of the results, or the affected rows.
	Count int
	//
	TotalCount int
	//
	PageLimit int
	//
	TotalPage int
	// LasyQuery is last executed query.
	LastQuery string
	// LastInsertID is the last insert ID.
	LastInsertID int
	// LastInsertIDs is the last insert IDs.
	LastInsertIDs []int
	// LastRow is the `*sql.Rows` from the last result.
	LastRows *sql.Rows
	// LastRow is the `*sql.Row` from the last result.
	LastRow *sql.Row
}

// newWrapper creates a new database function wrapper by the passed database connection.
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{db: db, Timestamp: &Timestamp{}}
}

// clean cleans the last executed result.
func (w *Wrapper) clean() {
	w.tableName = []string{}
	w.params = []interface{}{}
	w.onDuplicateColumns = []string{}
	w.groupBy = []string{}
	w.joins = map[string]*join{}
	w.orders = []order{}
	w.conditions = []condition{}
	w.havingConditions = []condition{}
	w.limit = []int{}
	w.query = ""
}

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
		if len(v.params) > 0 {
			w.params = append(w.params, v.params...)
		}
	case function:
		if len(v.values) > 0 {
			w.params = append(w.params, v.values...)
		}
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
	case function:
		param = v.query
	case nil:
		param = "NULL"
	default:
		param = "?"
	}
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

	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			columns += fmt.Sprintf("%s, ", column)
			values += fmt.Sprintf("%s, ", w.bindParam(value))
		}
		values = fmt.Sprintf("(%s)", trim(values))

	case []map[string]interface{}:
		for index, single := range realData {
			var currentValues string
			for column, value := range single {
				// Get the column names from the first data set only.
				if index == 0 {
					columns += fmt.Sprintf("%s, ", column)
				}
				currentValues += fmt.Sprintf("%s, ", w.bindParam(value))
			}
			values += fmt.Sprintf("(%s), ", trim(currentValues))
		}
		values = trim(values)
	}
	columns = trim(columns)
	query = fmt.Sprintf("%s %sINTO %s (%s) VALUES %s ", operator, options, w.tableName[0], columns, values)
	return
}

// Table specifies the name of the table.
func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

// Insert builds and executes the insert query.
func (w *Wrapper) Insert(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	w.buildQuery()
	return
}

// InsertMulti builds and executes a single insert query with the many rows.
func (w *Wrapper) InsertMulti(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	w.buildQuery()
	return
}

// Replace builds and executes the replace query just like what `Insert` does.
func (w *Wrapper) Replace(data interface{}) (err error) {
	w.query = w.buildInsert("REPLACE", data)
	w.buildQuery()
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

// OnDuplicate stores the columns which would be updated when the inserted row has duplicated.
func (w *Wrapper) OnDuplicate(columns []string, lastInsertID ...string) *Wrapper {
	w.onDuplicateColumns = columns
	if len(lastInsertID) != 0 {
		w.lastInsertIDColumn = lastInsertID[0]
	}
	return w
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

// Update updates the rows with the specified data.
func (w *Wrapper) Update(data interface{}) (err error) {
	w.query = w.buildUpdate(data)
	w.buildQuery()
	return
}

// Limit creates the limit statement in the end of the query.
func (w *Wrapper) Limit(from int, count ...int) *Wrapper {
	if len(count) == 0 {
		w.limit = []int{from}
	} else {
		w.limit = []int{from, count[0]}
	}
	return w
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

// Get gets the specified columns of the rows from the specifed database table.
func (w *Wrapper) Get(columns ...string) (err error) {
	w.query = w.buildSelect(columns...)
	w.buildQuery()
	return
}

// GetOne gets the specified columns of a single row from the specifed database table.
func (w *Wrapper) GetOne(columns ...string) (err error) {
	w.query = w.buildSelect(columns...)
	w.buildQuery()
	return
}

// GetValue gets the value of the specified column of the rows, you'll get the slice of the values if you didn't specify `LIMIT 1`.
func (w *Wrapper) GetValue(column string) (err error) {
	err = w.Get(fmt.Sprintf("%s AS Value", column))
	return
}

// Paginate acts the same as `Get` but with the automatically page caculation.
// Make sure you have specified the `PageLimit` (Default: 20) to limit the rows of a page.
func (w *Wrapper) Paginate(pageCount int, columns ...string) (err error) {
	err = w.Limit(w.PageLimit*(pageCount-1), w.PageLimit).Get(columns...)
	w.TotalPage = w.TotalCount / w.PageLimit
	return
}

// RawQuery executes the passed raw query and binds the passed values to the prepared statments.
func (w *Wrapper) RawQuery(query string, values ...interface{}) (err error) {
	w.query = query
	w.LastQuery = w.query
	w.bindParams(values)
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

// buildConditions cooks the conditions to the `WHERE` or the `HAVING` statement.
func (w *Wrapper) buildConditions(conditions []condition) (query string) {
	for i, v := range conditions {
		if i != 0 {
			query += fmt.Sprintf("%s ", v.connector)
		}

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

// buildDelete builds and executes the delete query.
func (w *Wrapper) buildDelete(tableNames ...string) (query string) {
	query += fmt.Sprintf("DELETE FROM %s ", strings.Join(tableNames, ", "))
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
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query

	if !w.isSubQuery {
		w.clean()
	}
}

// Delete deletes the row(s), use it with the `Where` condition so your whole table won't be wiped.
// It's very important alright? Cause .. you know ..fuck.
func (w *Wrapper) Delete() (err error) {
	w.query = w.buildDelete(w.tableName...)
	w.buildQuery()
	return
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

// OrderBy orders the getting result based on a single column (or the fields) with the specified sorting like `ASC` and `DESC`.
func (w *Wrapper) OrderBy(column string, args ...interface{}) *Wrapper {
	w.orders = append(w.orders, order{
		column: column,
		args:   args,
	})
	return w
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

// GroupBy groups the columns when executing the query.
func (w *Wrapper) GroupBy(columns ...string) *Wrapper {
	w.groupBy = columns
	return w
}

// saveJoin saves the table joining information.
func (w *Wrapper) saveJoin(table interface{}, typ string, condition string) {
	switch v := table.(type) {
	case *Wrapper:
		w.joins[v.query] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	case string:
		w.joins[v] = &join{
			typ:       typ,
			table:     table,
			condition: condition,
		}
	}
}

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

// saveJoinCondition saves a table join condition to the specifed table joining information.
func (w *Wrapper) saveJoinCondition(connector string, table interface{}, args ...interface{}) {
	switch v := table.(type) {
	case *Wrapper:
		w.joins[v.query].conditions = append(w.joins[v.query].conditions, condition{
			args:      args,
			connector: connector,
		})
	case string:
		w.joins[v].conditions = append(w.joins[v].conditions, condition{
			args:      args,
			connector: connector,
		})
	}
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

// JoinWhere creates a `HAVING AND` statement for the specified joining table.
func (w *Wrapper) JoinHaving(table interface{}, args ...interface{}) *Wrapper {
	//w.saveJoinCondition("AND", table, args...)
	//return w
}

// JoinOrWhere creates a `HAVING OR` statement for the specified joining table.
func (w *Wrapper) JoinOrHaving(table interface{}, args ...interface{}) *Wrapper {
	//w.saveJoinCondition("OR", table, args...)
	//return w
}

// buildJoin builds the join statement.
func (w *Wrapper) buildJoin() (query string) {
	if len(w.joins) == 0 {
		return
	}

	for _, v := range w.joins {
		query += fmt.Sprintf("%s ", v.typ)
		switch d := v.table.(type) {
		case *Wrapper:
			query += fmt.Sprintf("%s AS %s ON ", w.bindParam(d), d.alias)
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
	w.query = w.buildSelect()
	w.buildQuery()
	return
}

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

// Begin starts a transcation.
func (w *Wrapper) Begin() *Wrapper {
	return w
}

// Rollback rolls the changes back to where the transaction started.
func (w *Wrapper) Rollback() *Wrapper {
	return w
}

// Commit commits the current transaction, so the changes will be saved into the database permanently.
func (w *Wrapper) Commit() *Wrapper {
	return w
}

// SetLockMethod sets the lock method before locked the tables, it could be `WRITE` or `READ`.
func (w *Wrapper) SetLockMethod() *Wrapper {
	return w
}

// Lock locks the tables with the specified lock method.
func (w *Wrapper) Lock(tableNames ...string) *Wrapper {
	return w
}

// Unlock unlocks the locked tables.
func (w *Wrapper) Unlock(tableNames ...string) *Wrapper {
	return w
}

// SetQueryOption sets the query options like `SQL_NO_CACHE`.
func (w *Wrapper) SetQueryOption(options ...string) *Wrapper {
	return w
}

// SetTrace sets the trace mode as on or off, so you can get the traces by accessing the `Trace` property.
func (w *Wrapper) SetTrace(enable bool) *Wrapper {
	return w
}

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
