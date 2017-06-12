package main

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	connection *sql.DB
	table      table
	columns    []column
	LastQuery  string
}

type table struct {
	name        string
	comment     string
	primaryKeys []keys
	indexKeys   []keys
	uniqueKeys  []keys
	foreignKeys []keys
	engineType  string
}

type column struct {
	name          string
	dataType      string
	length        interface{}
	comment       string
	unsigned      bool
	primary       bool
	unique        bool
	index         bool
	foreign       string
	autoIncrement bool
	defaultValue  string
	nullable      bool
	extras        bool
}

type keys struct {
	name    string
	columns []string
}

// FKEY

func (m *Migration) setColumnType(dataType string, arg ...interface{}) *Migration {
	m.columns[len(m.columns)-1].dataType = dataType
	if len(arg) == 1 {
		m.columns[len(m.columns)-1].length = arg[0]
	}

	return m
}

func (m *Migration) TinyInt(length int) *Migration {
	return m.setColumnType("tinyint", length)
}

func (m *Migration) SmallInt(length int) *Migration {
	return m.setColumnType("smallint", length)
}

func (m *Migration) MediumInt(length int) *Migration {
	return m.setColumnType("mediumint", length)
}

func (m *Migration) Int(length int) *Migration {
	return m.setColumnType("int", length)
}

func (m *Migration) BigInt(length int) *Migration {
	return m.setColumnType("bigint", length)
}

func (m *Migration) Char(length int) *Migration {
	return m.setColumnType("char", length)
}

func (m *Migration) Varchar(length int) *Migration {
	return m.setColumnType("varchar", length)
}

func (m *Migration) TinyText() *Migration {
	return m.setColumnType("tinytext")
}

func (m *Migration) Text() *Migration {
	return m.setColumnType("text")
}

func (m *Migration) MediumText() *Migration {
	return m.setColumnType("mediumtext")
}

func (m *Migration) LongText() *Migration {
	return m.setColumnType("longtext")
}

func (m *Migration) Binary() *Migration {
	return m.setColumnType("binary")
}

func (m *Migration) VarBinary() *Migration {
	return m.setColumnType("varbinary")
}

func (m *Migration) Bit() *Migration {
	return m.setColumnType("bit")
}

func (m *Migration) Blob() *Migration {
	return m.setColumnType("blob")
}

func (m *Migration) MediumBlob() *Migration {
	return m.setColumnType("mediumblob")
}

func (m *Migration) LongBlob() *Migration {
	return m.setColumnType("longblob")
}

func (m *Migration) Date() *Migration {
	return m.setColumnType("date")
}

func (m *Migration) DateTime() *Migration {
	return m.setColumnType("dateTime")
}

func (m *Migration) Time() *Migration {
	return m.setColumnType("time")
}

func (m *Migration) Timestamp() *Migration {
	return m.setColumnType("timestamp")
}

func (m *Migration) Year() *Migration {
	return m.setColumnType("year")
}

func (m *Migration) Double(length []int) *Migration {
	return m.setColumnType("double", length)
}

func (m *Migration) Decimal(length []int) *Migration {
	return m.setColumnType("decimal", length)
}

func (m *Migration) Float(length []int) *Migration {
	return m.setColumnType("float", length)
}

func (m *Migration) Enum(types []string) *Migration {
	return m.setColumnType("enum", types)
}

func (m *Migration) Set(types []string) *Migration {
	return m.setColumnType("set", types)
}

func (m *Migration) Column(name string) *Migration {
	m.columns = append(m.columns, column{name: name})
	return m
}

func (m *Migration) InnoDB() *Migration {
	m.table.engineType = "innodb"
	return m
}

func (m *Migration) MyISAM() *Migration {
	m.table.engineType = "myisam"
	return m
}

//
//    Primary()
//    Primary([]string{"column1", "column2"})
//    Primary("primary_keys", []string{"column1", "column2"})
//
func (m *Migration) Primary(args ...interface{}) *Migration {
	switch len(args) {
	// Primary()
	case 0:
		m.columns[len(m.columns)-1].primary = true
	// Primary([]string{"column1", "column2"})
	case 1:
	// Primary("primary_keys", []string{"column1", "column2"})
	case 2:
		m.table.primaryKeys = append(m.table.primaryKeys, keys{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

func (m *Migration) Unique(args ...interface{}) *Migration {
	switch len(args) {
	// Unique()
	case 0:
		m.columns[len(m.columns)-1].primary = true
	// Unique("primary_keys", []string{"column1", "column2"})
	case 2:
		m.table.primaryKeys = append(m.table.primaryKeys, keys{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

func (m *Migration) Index(args ...interface{}) *Migration {
	return m
}

func (m *Migration) Foreign(args ...interface{}) *Migration {
	return m
}

func (m *Migration) Nullable() *Migration {
	m.columns[len(m.columns)-1].defaultValue = ""
	m.columns[len(m.columns)-1].nullable = true
	return m
}

func (m *Migration) Unsigned() *Migration {
	m.columns[len(m.columns)-1].unsigned = true
	return m
}

func (m *Migration) Comment(text string) *Migration {
	m.columns[len(m.columns)-1].comment = text
	return m
}

func (m *Migration) Default(value string) *Migration {
	m.columns[len(m.columns)-1].defaultValue = value
	return m
}

func (m *Migration) AutoIncrement() *Migration {
	m.columns[len(m.columns)-1].autoIncrement = true
	return m
}

func (m *Migration) Create(tableName string, comment string) {
	m.table.name = tableName
	m.table.comment = comment
}

func (m *Migration) Drop(tableNames ...string) *Migration {
	for _, name := range tableNames {
		query := fmt.Sprintf("DROP TABLE `%s`", name)
		m.connection.Exec(query)

		m.LastQuery = query
	}
	return m
}
