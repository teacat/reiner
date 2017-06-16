package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// DB represents a database connection.
type DB struct {
	connection *sql.DB
	isSubQuery bool
	Count      int
	LastQuery  string
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
func (d *DB) Insert(tableName string, data interface{}) {
	if d.isSubQuery {
		return
	}
}

//
func (d *DB) OnDuplicate(columns []string, lastInsertID string) {

}

//
func (d *DB) InsertMulti(tableName string, data interface{}) {

}

//
func (d *DB) Where(property string, value interface{}, operator string) {

}

//
func (d *DB) OrWhere(property string, value interface{}, operator string) {

}

//
func (d *DB) Update(tableName string, data interface{}) {

}

//
func (d *DB) Limit(count int) {

}

//
func (d *DB) Bind(model interface{}) {

}

//
func (d *DB) Get(tableName string, columns string) {

}

//
func (d *DB) GetOne(tableName string, columns string) {

}

//
func (d *DB) GetValue(tableName string, column string) {

}

//
func (d *DB) Paginate(tableName string, paging int) {

}

//
func (d *DB) RawQuery(query string, data ...interface{}) {

}

//
func (d *DB) RawQueryOne(query string, data ...interface{}) {

}

//
func (d *DB) RawQueryValue(query string, data ...interface{}) {

}

//
func (d *DB) Having(property string, value interface{}, operator string) {

}

//
func (d *DB) OrHaving(property string, value interface{}, operator string) {

}

//
func (d *DB) Delete(tableName string) {

}

//
func (d *DB) OrderBy(values ...string) {

}

//
func (d *DB) GroupBy(column string) {

}

//
func (d *DB) Join(tableName string, condition string, direction string) {

}

//
func (d *DB) JoinWhere() {

}

//
func (d *DB) JoinOrWhere() {

}

//
func (d *DB) SubQuery() {

}

//
func (d *DB) Has() {

}

//
func (d *DB) Now() {

}

//
func (d *DB) Func() {

}

//
func (d *DB) Disconnect() {

}

//
func (d *DB) Ping() {

}

//
func (d *DB) Connect() {

}

//
func (d *DB) Begin() {

}

//
func (d *DB) Rollback() {

}

//
func (d *DB) Commit() {

}

//
func (d *DB) SetLockMethod(method ...string) {

}

//
func (d *DB) Lock(tableNames ...string) {

}

//
func (d *DB) Unlock() {

}

//
func (d *DB) SetQueryOption(options ...string) {

}

//
func (d *DB) Migration() *Migration {
	return &Migration{connection: d.connection}
}
