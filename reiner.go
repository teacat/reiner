package reiner

func main() {

}

//
func New(dataSourceNames ...interface{}) (*Wrapper, error) {
	var masters, slaves []string
	// One master only
	if len(dataSourceNames) == 1 {
		masters = append(masters, dataSourceNames[0].(string))
		//slaves = append(slaves, dataSourceNames[0].(string))
		// Master(s) and the slave(s).
	} else if len(dataSourceNames) == 2 {
		switch v := dataSourceNames[0].(type) {
		// Multiple masters.
		case []string:
			masters = v
		// Single master.
		case string:
			masters = append(masters, v)
		}
		switch v := dataSourceNames[1].(type) {
		// Multiple slaves.
		case []string:
			slaves = v
		// Single slave.
		case string:
			slaves = append(slaves, v)
		}
	}
	d, err := newDatabase(masters, slaves)
	if err != nil {
		return &Wrapper{}, err
	}
	return newWrapper(d), nil
}
