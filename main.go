package main

import (
	"github.com/TeaMeow/Reiner/database"
	"github.com/TeaMeow/Reiner/wrapper"
)

func main() {

}

//
func New(dataSourceNames ...interface{}) *wrapper.Wrapper {
	var masters, slaves []string
	// One master only
	if len(dataSourceNames) == 1 {
		masters = append(masters, v)
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
		switch dataSourceNames[1].(type) {
		// Multiple slaves.
		case []string:
			slaves = v
		// Single slave.
		case string:
			slaves = append(slaves, v)
		}
	}
	d := database.New(masters, slaves)
	return wrapper.New(d)
}
