package main

import (
	"database/sql"
	"fmt"
	"strings"
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
	defaultValue  interface{}
	nullable      bool
	extras        bool
}

type keys struct {
	name          string
	columns       []string
	targetColumns []string
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

func (m *Migration) Enum(types ...interface{}) *Migration {
	return m.setColumnType("enum", types)
}

func (m *Migration) Set(types ...interface{}) *Migration {
	return m.setColumnType("set", types)
}

func (m *Migration) Column(name string) *Migration {
	m.columns = append(m.columns, column{name: name, defaultValue: false})
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

func (m *Migration) Primary(args ...interface{}) *Migration {
	switch len(args) {
	// Primary()
	case 0:
		m.columns[len(m.columns)-1].primary = true
	// Primary([]string{"column1", "column2"})
	case 1:
		m.table.primaryKeys = append(m.table.primaryKeys, keys{
			columns: args[0].([]string),
		})
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
		m.columns[len(m.columns)-1].unique = true
	// Unique([]string{"column1", "column2"})
	case 1:
		m.table.uniqueKeys = append(m.table.uniqueKeys, keys{
			columns: args[0].([]string),
		})
	// Unique("unique_keys", []string{"column1", "column2"})
	case 2:
		m.table.uniqueKeys = append(m.table.uniqueKeys, keys{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

func (m *Migration) Index(args ...interface{}) *Migration {
	switch len(args) {
	// Index()
	case 0:
		m.columns[len(m.columns)-1].index = true
	// Index([]string{"column1", "column2"})
	case 1:
		m.table.indexKeys = append(m.table.indexKeys, keys{
			columns: args[0].([]string),
		})
	// Index("index_keys", []string{"column1", "column2"})
	case 2:
		m.table.indexKeys = append(m.table.indexKeys, keys{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

// .Varchar(32).Foreign("users.id")
// .Foreign("foreign_keys", []string{"id", "password"}, []string{"users.id", "users.password"})

func (m *Migration) Foreign(args ...interface{}) *Migration {
	switch len(args) {
	// Foreign("users.id")
	case 0:
		m.columns[len(m.columns)-1].foreign = args[0].(string)
	// Foreign([]string{"id", "password"}, []string{"users.id", "users.password"})
	case 2:
		m.table.foreignKeys = append(m.table.foreignKeys, keys{
			columns:       args[0].([]string),
			targetColumns: args[1].([]string),
		})
	// Foreign("foreign_keys", []string{"id", "password"}, []string{"users.id", "users.password"})
	case 3:
		m.table.foreignKeys = append(m.table.foreignKeys, keys{
			name:          args[0].(string),
			columns:       args[1].([]string),
			targetColumns: args[2].([]string),
		})
	}
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

func (m *Migration) Default(value interface{}) *Migration {
	m.columns[len(m.columns)-1].defaultValue = value
	return m
}

func (m *Migration) AutoIncrement() *Migration {
	m.columns[len(m.columns)-1].autoIncrement = true
	return m
}

func (m *Migration) Create(tableName string, comment ...string) {
	m.table.name = tableName
	if len(comment) != 0 {
		m.table.comment = comment[0]
	}

	query := m.tableBuilder()

	m.LastQuery = query
}

func (m *Migration) tableBuilder() (query string) {
	var contentQuery string

	columnQuery := m.columnBuilder()
	foreignQuery := m.indexBuilder("FOREIGN KEY")
	primaryQuery := m.indexBuilder("PRIMARY KEY")
	uniqueQuery := m.indexBuilder("UNIQUE KEY")
	indexQuery := m.indexBuilder("INDEX")

	query = fmt.Sprintf("CREATE TABLE `%s` ", m.table.name)

	// Columns, keys.
	if columnQuery != "" {
		contentQuery += fmt.Sprintf("%s, ", columnQuery)
	}
	if foreignQuery != "" {
		contentQuery += fmt.Sprintf("%s, ", foreignQuery)
	}
	if primaryQuery != "" {
		contentQuery += fmt.Sprintf("%s, ", primaryQuery)
	}
	if uniqueQuery != "" {
		contentQuery += fmt.Sprintf("%s, ", uniqueQuery)
	}
	if indexQuery != "" {
		contentQuery += fmt.Sprintf("%s, ", indexQuery)
	}
	if contentQuery != "" {
		query += fmt.Sprintf("(%s) ", trim(contentQuery))
	}

	// Engine type.
	engineType := m.table.engineType
	if engineType == "" {
		engineType = "INNODB"
	}
	query += fmt.Sprintf("ENGINE=%s, ", engineType)

	// Comment.
	if m.table.comment != "" {
		query += fmt.Sprintf("COMMENT='%s', ", m.table.comment)
	}

	// Remove the unnecessary comma and the space.
	query = trim(query)
	return
}

func (m *Migration) indexBuilder(indexName string) (query string) {
	var keys []keys
	var targetTable, targetColumns string

	// Get the key groups by the index name.
	switch indexName {
	case "PRIMARY KEY":
		keys = m.table.primaryKeys
	case "UNIQUE KEY":
		keys = m.table.uniqueKeys
	case "INDEX":
		keys = m.table.indexKeys
	case "FOREIGN KEY":
		keys = m.table.foreignKeys
	}

	// Each index group.
	for _, v := range keys {
		// Build the column query. (`column_1`, `column_2`)
		columns := fmt.Sprintf("`%s`", strings.Join(v.columns, "`,`"))

		// Build the query for the target columns of the foreign keys.
		if len(v.targetColumns) != 0 {
			for _, c := range v.targetColumns {
				// Get the target table name from the target columns. (targetTable.targetColumn)
				targetTable = strings.Split(c, ".")[0]
				// Removed the table name in the column name and build the query.
				targetColumns += fmt.Sprintf("`%s`, ", strings.Split(c, ".")[0])
			}
			// Remove the unnecessary comma and the space.
			targetColumns = trim(targetColumns)
		}

		// Indexs without group name.
		if v.name == "" && len(v.targetColumns) == 0 {
			query += fmt.Sprintf("%s (%s), ", indexName, columns)
			// Naming indexes.
		} else if v.name != "" && len(v.targetColumns) == 0 {
			query += fmt.Sprintf("%s `%s` (%s), ", indexName, v.name, columns)
			// Foreign keys without group name.
		} else if v.name == "" && len(v.targetColumns) != 0 {
			query += fmt.Sprintf("%s (%s) REFERENCES %s (%s), ", indexName, columns, targetTable, targetColumns)
			// Foreign keys.
		} else if v.name != "" && len(v.targetColumns) != 0 {
			query += fmt.Sprintf("CONSTRAINT %s %s (%s) REFERENCES %s (%s), ", v.name, indexName, columns, targetTable, targetColumns)
		}
	}
	// Remove the unnecessary comma and the space.
	query = trim(query)
	return
}

func trim(input string) (result string) {
	if len(input) == 0 {
		result = strings.TrimSpace(input)
	} else {
		result = strings.TrimSpace(input[0 : len(input)-2])
	}
	return
}

// columnBuilder builds the query from the columns.
func (m *Migration) columnBuilder() (query string) {
	for _, v := range m.columns {
		// Column name.
		query += fmt.Sprintf("`%s` ", v.name)

		// Data types.
		dataType := strings.ToUpper(v.dataType)
		switch t := v.length.(type) {
		// VARCHAR(30)
		case int:
			query += fmt.Sprintf("%s(%d) ", dataType, t)
		// FLOAT(1, 2) or ENUM(1, 2, "A", "B")
		case []interface{}:
			// Extracting the options from the length.
			options := ""
			for _, o := range t {
				switch tt := o.(type) {
				case int:
					options += fmt.Sprintf("%d, ", tt)
				case string:
					options += fmt.Sprintf("'%s', ", tt)
				}
			}
			// Trim the comma and the space.
			query += fmt.Sprintf("%s(%s)", dataType, trim(options))
		// DATETIME
		case nil:
			query += fmt.Sprintf("%s ", dataType)
		}

		// Unsigned.
		if v.unsigned {
			query += "UNSIGNED "
		}
		// Nullable.
		if !v.nullable {
			query += "NOT NULL "
		}
		// Auto increment.
		if v.autoIncrement {
			query += "AUTO_INCREMENT "
		}

		// Default value.
		switch t := v.defaultValue.(type) {
		case int:
			query += fmt.Sprintf("DEFAULT %d ", t)
		case nil:
			query += fmt.Sprintf("DEFAULT NULL ")
		case string:
			query += fmt.Sprintf("DEFAULT '%s' ", t)
		}

		// Keys.
		if v.primary {
			query += "PRIMARY KEY "
		}
		if v.unique {
			query += "UNIQUE "
		}
		if v.index {
			query += "INDEX "
		}

		// Comment.
		if v.comment != "" {
			query += fmt.Sprintf("COMMENT '%s'", v.comment)
		}

		// End.
		query += ", "
	}
	// Remove the last unnecessary comma
	query = trim(query)
	return
}

func (m *Migration) Drop(tableNames ...string) *Migration {
	for _, name := range tableNames {
		query := fmt.Sprintf("DROP TABLE `%s`", name)
		m.connection.Exec(query)

		m.LastQuery = query
	}
	return m
}
