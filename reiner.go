package reiner

// New creates a new database connection which provides the MySQL wrapper functions.
// The first data source name is for the master, the rest are for the slaves, which is used for the read/write split.
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// Check https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html for more information.
func New(dataSourceNames ...interface{}) (*Wrapper, error) {
	var slaves []string
	var master string

	switch len(dataSourceNames) {
	// Query builder mode.
	case 0:
		return &Wrapper{executable: false, Timestamp: &Timestamp{}}, nil
	// One master only.
	case 1:
		master = dataSourceNames[0].(string)
	// Master(s) and the slave(s).
	case 2:
		master = dataSourceNames[0].(string)
		switch v := dataSourceNames[1].(type) {
		// Multiple slaves.
		case []string:
			slaves = v
		// Single slave.
		case string:
			slaves = append(slaves, v)
		}
	}
	d, err := newDatabase(master, slaves)
	if err != nil {
		return &Wrapper{}, err
	}
	return newWrapper(d), nil
}
