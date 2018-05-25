package reiner

import (
	"database/sql"
	"strings"
)

// connection 重現了一個資料庫的連線。
type connection struct {
	db             *sql.DB
	tx             *sql.Tx
	lastCheck      int
	isHealth       bool
	dataSourceName string
}

// DB 是一個擁有許多連線的資料庫來源。
type DB struct {
	slaves         []*connection
	master         *connection
	hasSlave       bool
	lastSlaveIndex int
}

// openDatabase 會開啟一個新的資料庫連線。
func openDatabase(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return db, err
	}
	if err = db.Ping(); err != nil {
		return db, err
	}
	return db, nil
}

// newDatabase 會建立一個新的資料庫，當有主從來源時會替這個資料庫建立多個連線。
// 如果僅有單個主要來源的話則會建立一個最主要的連線。
func newDatabase(master string, slaves []string) (*DB, error) {
	d := &DB{}
	// 如果沒有主從來源就建立一個最主要的連線。
	if len(slaves) == 0 {
		db, err := openDatabase(master)
		if err != nil {
			return d, err
		}
		d.master = &connection{
			db:             db,
			dataSourceName: master,
		}
		return d, nil
	}
	d.hasSlave = true
	// 連線到 Slave 資料庫。
	for _, v := range slaves {
		db, err := openDatabase(v)
		if err != nil {
			return d, err
		}
		d.slaves = append(d.slaves, &connection{
			db:             db,
			dataSourceName: v,
		})
	}
	return d, nil
}

// roundRobin 會輪詢資料庫連線來避免不斷地呼叫同個資料庫連線。
// 簡單來說就是個簡易型的負載平衡器。
func (d *DB) roundRobin(pool []*connection, currentIndex int) (index int) {
	length := len(pool) - 1
	index = currentIndex + 1
	if index > length {
		index = 0
	}
	return
}

// getSlave 會取得一個可用的 Slave 資料庫連線。
func (d *DB) getSlave() (db *sql.DB) {
	index := d.roundRobin(d.slaves, d.lastSlaveIndex)
	db = d.slaves[index].db
	// Set the last index.
	d.lastSlaveIndex = index
	return
}

// getDB 會基於 SQL 查詢指令來取得一個適用的資料庫連線，這會被用在讀／寫區分的資料庫上。
// BUG(r): Reimplemented needed.
func (d *DB) getDB(query ...string) (db *sql.DB) {
	if len(query) == 0 || !d.hasSlave {
		db = d.master.db
		return
	}
	action := strings.Split(query[0], " ")[0]
	switch action {
	case "SELECT":
		db = d.getSlave()
	default:
		db = d.master.db
	}
	return
}

// Begin 會基於目前的資料庫連線來開始一段新的交易過程。
func (d *DB) Begin() (*sql.Tx, error) {
	return d.master.db.Begin()
}

// Rollback 會回溯交易時所發生的事情。
func (d *DB) Rollback() error {
	if d.master.tx == nil {
		return ErrUnbegunTransaction
	}
	err := d.master.tx.Rollback()
	if err != nil {
		return err
	}
	d.master.tx = nil
	return nil
}

// Commit 會結束一個交易過程並保存其變更為永久資料。
func (d *DB) Commit() error {
	if d.master.tx == nil {
		return ErrUnbegunTransaction
	}
	err := d.master.tx.Commit()
	if err != nil {
		return err
	}
	d.master.tx = nil
	return nil
}

// Ping 會以 ping 來檢查所有的資料庫連線（包括 Slave 連線）。
func (d *DB) Ping() error {
	var err error
	err = d.master.db.Ping()
	if err != nil {
		return err
	}
	for _, v := range d.slaves {
		err = v.db.Ping()
		if err != nil {
			return err
		}
	}
	return nil
}

// Disconnect 會斷開所有連線（包括 Slave 連線）。
func (d *DB) Disconnect() error {
	var err error
	err = d.master.db.Close()
	if err != nil {
		return err
	}
	for _, v := range d.slaves {
		err = v.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Connect 會重新連接所有資料庫連線（包括 Slave 連線）。
func (d *DB) Connect() error {
	db, err := sql.Open("mysql", d.master.dataSourceName)
	if err != nil {
		return err
	}
	d.master.db = db
	for k, v := range d.slaves {
		db, err := sql.Open("mysql", v.dataSourceName)
		if err != nil {
			return err
		}
		d.slaves[k].db = db
	}
	return nil
}

// Prepare 會準備 SQL 查詢指令。
func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	if d.master.tx != nil {
		return d.master.tx.Prepare(query)
	}
	return d.getDB(query).Prepare(query)
}

// Exec 會執行 SQL 查詢指令並且回傳一個原生結果表示影響的行列數和插入的編號。
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	if d.master.tx != nil {
		return d.master.tx.Exec(query, args...)
	}
	return d.getDB(query).Exec(query, args...)
}

// Query 會執行 SQL 查詢指令並且回傳一個原生的行列結果供後續掃描列出。
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.master.tx != nil {
		return d.master.tx.Query(query, args...)
	}
	return d.getDB(query).Query(query, args...)
}
