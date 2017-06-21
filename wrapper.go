package reiner

import (
	"database/sql"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// Wrapper represents a database connection.
type Wrapper struct {
	db         *DB
	isSubQuery bool
	tableName  []string

	destination interface{}

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

func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

func (w *Wrapper) Insert(data interface{}) (err error) {
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
