package reiner

// New 會建立一個新的連線並且提供 MySQL 的指令包覆程式，當沒有傳入參數時，
// 會變成 SQL 指令建置模式，而這個情況下你可以透過 Reiner 來建立 SQL 查詢指令。
//     .New()
// 如果要連接到一個資料庫，第一個參數則需要傳入主要資料庫的 DSN（資料來源名稱）。
//     .New("root:root@/master")
// 若需要讀／寫區分，第一個參數則是為主要資料庫的 DSN，而其餘的則是 Slave 資料庫的 DSN。
//     .New("root:root@/master", root:root@/slave")
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// 查看 https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html 了解更多資訊。
func New(dataSourceNames ...interface{}) (*Builder, error) {
	var slaves []string
	var master string

	switch len(dataSourceNames) {
	// SQL 指令建置模式。
	case 0:
		return &Builder{executable: false, Timestamp: &Timestamp{}}, nil
	// 單個主要資料庫連線。
	case 1:
		master = dataSourceNames[0].(string)
	// 主從資料庫。
	case 2:
		master = dataSourceNames[0].(string)
		switch v := dataSourceNames[1].(type) {
		// 多個 Slaves。
		case []string:
			slaves = v
		// 單個 Slave。
		case string:
			slaves = append(slaves, v)
		}
	}
	d, err := newDatabase(master, slaves)
	if err != nil {
		return &Builder{}, err
	}
	return newBuilder(d), nil
}
