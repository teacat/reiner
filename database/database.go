package database

import (
	"database/sql"
	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// DB represents a database connection.
type DB struct {
	connection *sql.DB
	isSubQuery bool

	// Count is the count of the results, or the affected rows.
	Count int
	// LasyQuery is last executed query.
	LastQuery string
}

// New creates a new database connection which provides the MySQL wrapper functions.
// The first data source name is for the master, the rest are for the slaves, which is used for the read/write split.
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// Check https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html for more information.
func New(dataSourceName string, slaveDataSourceNames ...[]string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return &DB{}, err
	}
	if err = db.Ping(); err != nil {
		return &DB{}, err
	}

	return &DB{connection: db}, err
}

// Insert inserts the data to the specified table.
func (d *DB) Insert(tableName string, data interface{}) (lastInsertID int, err error) {
	//d.buildInsert(tableName, data, "INSERT")
	return
}

// OnDuplicate specifies the `ON DUPLICATE KEY UPDATE` statement for the SQL queries.
func (d *DB) OnDuplicate(columns []string, lastInsertID string) *DB {
	return d
}

// InsertMulti inserts the multiple data into the same table at the same time.
func (d *DB) InsertMulti(tableName string, data interface{}) (lastInsertIDs []int, err error) {
	return
}

// Where specifies the `WHERE (AND)` statement for the SQL queries.
func (d *DB) Where(property string, value interface{}, operator string) *DB {
	return d
}

// OrWhere specifies the `WHERE (OR)` statement for the SQL queries.
func (d *DB) OrWhere(property string, value interface{}, operator string) *DB {
	return d
}

// Update updates the specified table with the data.
func (d *DB) Update(tableName string, data interface{}) (err error) {
	return
}

// Limit limits the how much rows of the result we can get.
func (d *DB) Limit(count int) *DB {
	return d
}

// Bind binds the struct, map, slice to the result.
func (d *DB) Bind(model interface{}) *DB {
	return d
}

// Get gets the data from the specified table
// and mapping it to the specified slice.
func (d *DB) Get(tableName string, columns string) (err error) {
	return
}

// GetOne gets the data from the specified table with only one row,
// and it'll mapping to a single struct or a map not a slice.
func (d *DB) GetOne(tableName string, columns string) (err error) {
	return
}

// GetValue gets the value of the single column from the specified table,
// and mapping it to the specified variable.
func (d *DB) GetValue(tableName string, column string) (err error) {
	return
}

// Paginate is a simple wrapper to `Get()` with a automatically page caculate function,
// you have to specify the `db.PageLimit` before you use the function.
//     currentPage := 2
//     db.PageLimit = 20
//     db.Bind(&books).Paginate("books", 2)
func (d *DB) Paginate(tableName string, paging int) (err error) {
	return
}

// RawQuery executes a raw query and mapping the result to the binded struct, map or a slice.
func (d *DB) RawQuery(query string, data ...interface{}) (err error) {
	return
}

// RawQueryOne executes a raw query and get the first result then mapping it to the binded struct, or the map.
func (d *DB) RawQueryOne(query string, data ...interface{}) (err error) {
	return
}

// RawQueryValue executes a raw query and get the value of the column from the result,
// and mapping it to a binded variable.
func (d *DB) RawQueryValue(query string, data ...interface{}) (err error) {
	return
}

// Having specifies the `HAVING (AND)` statement for the SQL queries.
func (d *DB) Having(property string, value interface{}, operator string) *DB {
	return d
}

// OrHaving specifies the `HAVING (OR)` statement for the SQL queries.
func (d *DB) OrHaving(property string, value interface{}, operator string) *DB {
	return d
}

// Delete deletes the specified row from the specified table.
func (d *DB) Delete(tableName string) (affected bool, err error) {
	return
}

// OrderBy specifies the `ORDER BY` statement for the SQL queries.
func (d *DB) OrderBy(values ...string) *DB {
	return d
}

// GroupBy specifies the `GROUP BY` statement for the SQL queries.
func (d *DB) GroupBy(column string) *DB {
	return d
}

// Join joins the specified table to the current query, it could be a sub query.
func (d *DB) Join(tableName string, condition string, direction string) *DB {
	return d
}

// JoinWhere specifies the WHERE (AND) statement for the JOIN condition.
func (d *DB) JoinWhere() *DB {
	return d
}

// JoinOrWhere specifies the WHERE (OR) statement for the JOIN condition.
func (d *DB) JoinOrWhere() *DB {
	return d
}

// SubQuery creates a new wrapper for the sub query, thr quries in the wrapper won't be executed.
func (d *DB) SubQuery() *DB {
	return d
}

// Has returns true when there's a row was found,
// it's useful when checking the user does exist or not (or validating the password does match or not).
func (d *DB) Has() (has bool) {
	return
}

// Now returns generated interval function as an insert/update function.
func (d *DB) Now() {

}

// Func returns a specified MySQL function as an insert/update function.
func (d *DB) Func() {

}

// Interval returns generated interval function as an insert/update function.
func (d *DB) Interval() {

}

// Increment returns generated increment function as an insert/update function.
func (d *DB) Increment() {

}

// Decrement returns generated decrement function as an insert/update function.
func (d *DB) Decrement() {

}

// Disconnect disconnects the connection to the database.
func (d *DB) Disconnect() {

}

// Ping pings to the database.
func (d *DB) Ping() (err error) {
	return
}

// Connect connects to the database, it's useful when the connection lost.
func (d *DB) Connect() (err error) {
	return
}

// Begin starts a transaction.
func (d *DB) Begin() *DB {
	return d
}

// Rollback rollbacks a uncommited transaction.
func (d *DB) Rollback() (err error) {
	return
}

// Commit commits a transaction.
func (d *DB) Commit() (err error) {
	return
}

// SetLockMethod sets the lock method
func (d *DB) SetLockMethod(method ...string) *DB {
	return d
}

// Lock locks the specified tables with the specified lock method.
// It'll automatically unlocks the previous lock.
func (d *DB) Lock(tableNames ...string) (err error) {
	return
}

// Unlock unlocks the table lock.
func (d *DB) Unlock() (err error) {
	return
}

// SetQueryOption sets the option for the SQL queries.
func (d *DB) SetQueryOption(options ...string) {

}

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (d *DB) Migration() *Migration {
	return &Migration{connection: d.connection}
}
