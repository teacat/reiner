package reiner

import (
	"database/sql"
	"fmt"
	"strings"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

type tableName string

type condition struct {
	typ       string
	operator  string
	column    string
	subQuery  string
	rawQuery  string
	connector string
	values    []interface{}
}

type join struct {
	tableName  string
	typ        string
	condition  string
	conditions []condition
}

// Wrapper represents a database connection.
type Wrapper struct {
	db           *DB
	isSubQuery   bool
	query        string
	alias        string
	tableName    []string
	queryOptions []string
	destination  interface{}
	joins        map[tableName]join
	params       []interface{}

	//
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
	//
	LastInsertID int
	//
	LastInsertIDs []int
	//
	LastRows *sql.Rows
	//
	LastRow *sql.Row
}

// New creates a new database connection which provides the MySQL wrapper functions.
// The first data source name is for the master, the rest are for the slaves, which is used for the read/write split.
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// Check https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html for more information.
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{db: db}
}

func (w *Wrapper) clean() {
	w.tableName = []string{}
}

func (w *Wrapper) bindParams(params []interface{}) {
	for _, v := range params {
		w.bindParam(v)
	}
}

func (w *Wrapper) bindParam(value interface{}) {
	w.params = append(w.params, value)
}

func (w *Wrapper) buildPair(operator string, value interface{}) (query string) {
	switch v := value.(type) {
	// Is a sub query.
	case *Wrapper:
		subQuery := v.query
		params := v.params
		alias := ""
		if v.alias != "" {
			alias = fmt.Sprintf("%s ", v.alias)
		}
		w.bindParams(params)
		query = fmt.Sprintf("%s (%s) %s", operator, subQuery, alias)
	// Is values.
	default:
		w.bindParam(value)
		query = fmt.Sprintf(" %s ?", operator)
	}
	return
}

func (w *Wrapper) buildCondition(operator string, value []interface{}) {
	switch operator {
	case "NOT IN", "IN":
	case "NOT BETWEEN", "BETWEEN":
	case "NOT EXISTS", "EXISTS":
	default:
	}
}

func (w *Wrapper) buildJoin() {
	if len(w.joins) == 0 {
		return
	}
	for _, v := range w.joins {
		w.query += fmt.Sprintf("%s JOIN %s AS %s ON %s ", v.typ, v.tableName, v.tableName, v.condition)
		for _, c := range v.conditions {
			w.query += fmt.Sprintf("%s %s ", c.connector, c.column)
			w.buildCondition(c.operator, c.values)
		}
	}
	w.query += "xx"
	return
}

func (w *Wrapper) buildQuery(numRows int, data interface{}) {
	w.buildJoin()
	return
}

func (w *Wrapper) buildInsert(operation string, data interface{}) {
	w.query = fmt.Sprintf("%s %s INTO %s", operation, strings.Join(w.queryOptions, ", "), w.tableName[0])
	w.buildQuery(0, data)
	return
}

func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

func (w *Wrapper) Insert(data interface{}) (err error) {
	if w.isSubQuery {
		err = nil
		return
	}
	w.buildInsert("INSERT", data)
	w.LastQuery = w.query
	return
}

func (w *Wrapper) InsertMulti(data interface{}) (err error) {
	return
}

func (w *Wrapper) Replace(data interface{}) (err error) {
	return
}

func (w *Wrapper) Func(query string, data ...interface{}) (err error) {
	return
}

func (w *Wrapper) Now(format ...string) (err error) {
	return
}

func (w *Wrapper) OnDuplicate(columns []string, lastInsertID ...string) *Wrapper {
	return w
}

func (w *Wrapper) Update(data interface{}) (err error) {
	return
}

func (w *Wrapper) Limit(count int) *Wrapper {
	return w
}

func (w *Wrapper) Get(columns ...string) (err error) {
	return
}

func (w *Wrapper) GetOne(columns ...string) (err error) {
	return
}

func (w *Wrapper) GetValue(column string) (err error) {
	return
}

func (w *Wrapper) Paginate(pageCount int) (err error) {
	return
}

func (w *Wrapper) RawQuery(query string, values ...interface{}) (err error) {
	return
}

func (w *Wrapper) RawQueryOne(query string, values ...interface{}) (err error) {
	return
}

func (w *Wrapper) RawQueryValue(query string, values ...interface{}) (err error) {
	return
}

func (w *Wrapper) Having(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) OrHaving(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) Where(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) OrWhere(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereBetween(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereNotBetween(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereIn(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereNotIn(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereNull(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) WhereNotNull(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) Delete() (err error) {
	return
}

func (w *Wrapper) OrderBy(column string, args ...interface{}) *Wrapper { //sorting string, fields ...[]string
	return w
}

func (w *Wrapper) GroupBy(column string) *Wrapper {
	return w
}

func (w *Wrapper) LeftJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) RightJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) InnerJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) NatualJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) CrossJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) JoinWhere(table string, args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) JoinOrWhere(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) SubQuery(alias ...string) *Wrapper {
	return w
}

func (w *Wrapper) Has() (has bool, err error) {
	return
}

func (w *Wrapper) Disconnect() (err error) {
	return
}

func (w *Wrapper) Ping() (err error) {
	return
}

func (w *Wrapper) Connect() (err error) {
	return
}

func (w *Wrapper) Begin() *Wrapper {
	return w
}

func (w *Wrapper) Rollback() *Wrapper {
	return w
}

func (w *Wrapper) Commit() *Wrapper {
	return w
}

func (w *Wrapper) SetLockMethod() *Wrapper {
	return w
}

func (w *Wrapper) Lock() *Wrapper {
	return w
}

func (w *Wrapper) Unlock() *Wrapper {
	return w
}

func (w *Wrapper) SetQueryOption(options ...string) *Wrapper {
	return w
}

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
