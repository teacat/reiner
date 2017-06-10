package main

type DB struct {
	Count     int
	LastQuery string
}

//
func (d *DB) Insert(tableName string, data interface{}) {

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
func (d *DB) RawQueryOne() {

}

//
func (d *DB) RawQueryValue() {

}

//
func (d *DB) Having() {

}

//
func (d *DB) OrHaving() {

}

//
func (d *DB) Delete() {

}

//
func (d *DB) OrderBy() {

}

//
func (d *DB) GroupBy() {

}

//
func (d *DB) Join() {

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
func (d *DB) SetLockMethod() {

}

//
func (d *DB) Lock() {

}

//
func (d *DB) Unlock() {

}

//
func (d *DB) SetQueryOption() {

}

//
func (d *DB) Migration() {

}
