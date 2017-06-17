package wrapper

import (
	"fmt"

	"github.com/TeaMeow/Reiner/database"
	"github.com/TeaMeow/Reiner/migration"
	"github.com/TeaMeow/Reiner/toolkit"
	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// Wrapper represents a database connection.
type Wrapper struct {
	db         *database.DB
	isSubQuery bool
	dest       interface{}

	// Count is the count of the results, or the affected rows.
	Count int
	// LasyQuery is last executed query.
	LastQuery string
}

// New creates a new database connection which provides the MySQL wrapper functions.
// The first data source name is for the master, the rest are for the slaves, which is used for the read/write split.
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// Check https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html for more information.
func New(db *database.DB) *Wrapper {
	return &Wrapper{db: db}
}

// Insert inserts the data to the specified table.
func (w *Wrapper) Insert(tableName string, data interface{}) (lastInsertID int, err error) {
	//d.buildInsert(tableName, data, "INSERT")

	var columnQuery, valueQuery string
	var values []interface{}

	convertedData := convertor(data)
	for k, v := range convertedData {
		columnQuery += fmt.Sprintf("`%s`, ", k)
		valueQuery += "?, "
		values = append(values, v)
	}

	res, err := w.db.Exec(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, toolkit.Trim(columnQuery), toolkit.Trim(valueQuery)), values...)
	if err != nil {
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		return
	}
	lastInsertID = int(id)
	return
}

// OnDuplicate specifies the `ON DUPLICATE KEY UPDATE` statement for the SQL queries.
func (w *Wrapper) OnDuplicate(columns []string, lastInsertID string) *Wrapper {
	return w
}

// InsertMulti inserts the multiple data into the same table at the same time.
func (w *Wrapper) InsertMulti(tableName string, data interface{}) (lastInsertIDs []int, err error) {
	return
}

// Where specifies the `WHERE (AND)` statement for the SQL queries.
func (w *Wrapper) Where(property string, value interface{}, operator string) *Wrapper {
	return w
}

// OrWhere specifies the `WHERE (OR)` statement for the SQL queries.
func (w *Wrapper) OrWhere(property string, value interface{}, operator string) *Wrapper {
	return w
}

// Update updates the specified table with the data.
func (w *Wrapper) Update(tableName string, data interface{}) (err error) {
	return
}

// Limit limits the how much rows of the result we can get.
func (w *Wrapper) Limit(count int) *Wrapper {
	return w
}

// Bind binds the struct, map, slice to the result.
func (w *Wrapper) Bind(dest interface{}) *Wrapper {
	w.dest = dest
	return w
}

// Get gets the data from the specified table
// and mapping it to the specified slice.
func (w *Wrapper) Get(tableName string, columns ...string) (err error) {
	//rows, err := d.connection.Query(fmt.Sprintf("SELECT * FROM `%s`", tableName))

	return
}

// GetOne gets the data from the specified table with only one row,
// and it'll mapping to a single struct or a map not a slice.
func (w *Wrapper) GetOne(tableName string, columns string) (err error) {
	return
}

// GetValue gets the value of the single column from the specified table,
// and mapping it to the specified variable.
func (w *Wrapper) GetValue(tableName string, column string) (err error) {
	return
}

// Paginate is a simple wrapper to `Get()` with a automatically page caculate function,
// you have to specify the `db.PageLimit` before you use the function.
//     currentPage := 2
//     db.PageLimit = 20
//     db.Bind(&books).Paginate("books", 2)
func (w *Wrapper) Paginate(tableName string, paging int) (err error) {
	return
}

// RawQuery executes a raw query and mapping the result to the binded struct, map or a slice.
func (w *Wrapper) RawQuery(query string, data ...interface{}) (err error) {
	return
}

// RawQueryOne executes a raw query and get the first result then mapping it to the binded struct, or the map.
func (w *Wrapper) RawQueryOne(query string, data ...interface{}) (err error) {
	return
}

// RawQueryValue executes a raw query and get the value of the column from the result,
// and mapping it to a binded variable.
func (w *Wrapper) RawQueryValue(query string, data ...interface{}) (err error) {
	return
}

// Having specifies the `HAVING (AND)` statement for the SQL queries.
func (w *Wrapper) Having(property string, value interface{}, operator string) *Wrapper {
	return w
}

// OrHaving specifies the `HAVING (OR)` statement for the SQL queries.
func (w *Wrapper) OrHaving(property string, value interface{}, operator string) *Wrapper {
	return w
}

// Delete deletes the specified row from the specified table.
func (w *Wrapper) Delete(tableName string) (affected bool, err error) {
	return
}

// OrderBy specifies the `ORDER BY` statement for the SQL queries.
func (w *Wrapper) OrderBy(values ...string) *Wrapper {
	return w
}

// GroupBy specifies the `GROUP BY` statement for the SQL queries.
func (w *Wrapper) GroupBy(column string) *Wrapper {
	return w
}

// Join joins the specified table to the current query, it could be a sub query.
func (w *Wrapper) Join(tableName string, condition string, direction string) *Wrapper {
	return w
}

// JoinWhere specifies the WHERE (AND) statement for the JOIN condition.
func (w *Wrapper) JoinWhere() *Wrapper {
	return w
}

// JoinOrWhere specifies the WHERE (OR) statement for the JOIN condition.
func (w *Wrapper) JoinOrWhere() *Wrapper {
	return w
}

// SubQuery creates a new wrapper for the sub query, thr quries in the wrapper won't be executed.
func (w *Wrapper) SubQuery() *Wrapper {
	return w
}

// Has returns true when there's a row was found,
// it's useful when checking the user does exist or not (or validating the password does match or not).
func (w *Wrapper) Has() (has bool) {
	return
}

// Now returns generated interval function as an insert/update function.
func (w *Wrapper) Now() {

}

// Func returns a specified MySQL function as an insert/update function.
func (w *Wrapper) Func() {

}

// Interval returns generated interval function as an insert/update function.
func (w *Wrapper) Interval() {

}

// Increment returns generated increment function as an insert/update function.
func (w *Wrapper) Increment() {

}

// Decrement returns generated decrement function as an insert/update function.
func (w *Wrapper) Decrement() {

}

// Disconnect disconnects the connection to the database.
func (w *Wrapper) Disconnect() {

}

// Ping pings to the database.
func (w *Wrapper) Ping() (err error) {
	return
}

// Connect connects to the database, it's useful when the connection lost.
func (w *Wrapper) Connect() (err error) {
	return
}

// Begin starts a transaction.
func (w *Wrapper) Begin() *Wrapper {
	return w
}

// Rollback rollbacks a uncommited transaction.
func (w *Wrapper) Rollback() (err error) {
	return
}

// Commit commits a transaction.
func (w *Wrapper) Commit() (err error) {
	return
}

// SetLockMethod sets the lock method
func (w *Wrapper) SetLockMethod(method ...string) *Wrapper {
	return w
}

// Lock locks the specified tables with the specified lock method.
// It'll automatically unlocks the previous lock.
func (w *Wrapper) Lock(tableNames ...string) (err error) {
	return
}

// Unlock unlocks the table lock.
func (w *Wrapper) Unlock() (err error) {
	return
}

// SetQueryOption sets the option for the SQL queries.
func (w *Wrapper) SetQueryOption(options ...string) {

}

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *migration.Migration {
	return migration.New(w.db)
}

// convertor converts anything to a `map[string]interface{}` type,
// so we could get the column names and pass the data to the `Exec()` function easily.
func convertor(model interface{}) (result map[string]interface{}) {
	result = model.(map[string]interface{})
	return
}
