package reiner

import (
	"fmt"
	"strings"
)

// Migration represents a table migration.
type Migration struct {
	connection *DB
	table      table
	columns    []column

	// LasyQuery is last executed query.
	LastQuery string
}

// table contains the information of the current table.
type table struct {
	name        string
	comment     string
	primaryKeys []key
	indexKeys   []key
	uniqueKeys  []key
	foreignKeys []key
	engineType  string
}

// column contains the information of a single column.
type column struct {
	name          string
	dataType      string
	length        interface{}
	comment       string
	unsigned      bool
	primary       bool
	unique        bool
	foreign       string
	autoIncrement bool
	defaultValue  interface{}
	nullable      bool
	extras        bool
}

// key represents an index information.
type key struct {
	name          string
	columns       []string
	targetColumns []string
	onUpdate      string
	onDelete      string
}

// newMigration creates a new table migration by the passed database connection.
func newMigration(db *DB) *Migration {
	return &Migration{connection: db}
}

// TinyInt sets the data type of the latest column as `tinyint`.
func (m *Migration) TinyInt(length int) *Migration {
	return m.setColumnType("tinyint", length)
}

// SmallInt sets the data type of the latest column as `smallint`.
func (m *Migration) SmallInt(length int) *Migration {
	return m.setColumnType("smallint", length)
}

// MediumInt sets the data type of the latest column as `mediumint`.
func (m *Migration) MediumInt(length int) *Migration {
	return m.setColumnType("mediumint", length)
}

// Int sets the data type of the latest column as `int`.
func (m *Migration) Int(length int) *Migration {
	return m.setColumnType("int", length)
}

// BigInt sets the data type of the latest column as `bigint`.
func (m *Migration) BigInt(length int) *Migration {
	return m.setColumnType("bigint", length)
}

// Char sets the data type of the latest column as `char`.
func (m *Migration) Char(length int) *Migration {
	return m.setColumnType("char", length)
}

// Varchar sets the data type of the latest column as `varchar`.
func (m *Migration) Varchar(length int) *Migration {
	return m.setColumnType("varchar", length)
}

// TinyText sets the data type of the latest column as `tinytext`.
func (m *Migration) TinyText() *Migration {
	return m.setColumnType("tinytext")
}

// Text sets the data type of the latest column as `text`.
func (m *Migration) Text() *Migration {
	return m.setColumnType("text")
}

// MediumText sets the data type of the latest column as `mediumtext`.
func (m *Migration) MediumText() *Migration {
	return m.setColumnType("mediumtext")
}

// LongText sets the data type of the latest column as `longtext`.
func (m *Migration) LongText() *Migration {
	return m.setColumnType("longtext")
}

// Binary sets the data type of the latest column as `binary`.
func (m *Migration) Binary(length int) *Migration {
	return m.setColumnType("binary", length)
}

// VarBinary sets the data type of the latest column as `varbinary`.
func (m *Migration) VarBinary(length int) *Migration {
	return m.setColumnType("varbinary", length)
}

// Bit sets the data type of the latest column as `bit`.
func (m *Migration) Bit(length int) *Migration {
	return m.setColumnType("bit", length)
}

// TinyBlob sets the data type of the latest column as `tinyblob`.
func (m *Migration) TinyBlob() *Migration {
	return m.setColumnType("tinyblob")
}

// Blob sets the data type of the latest column as `blob`.
func (m *Migration) Blob() *Migration {
	return m.setColumnType("blob")
}

// MediumBlob sets the data type of the latest column as `mediumblob`.
func (m *Migration) MediumBlob() *Migration {
	return m.setColumnType("mediumblob")
}

// LongBlob sets the data type of the latest column as `longblob`.
func (m *Migration) LongBlob() *Migration {
	return m.setColumnType("longblob")
}

// Date sets the data type of the latest column as `date`.
func (m *Migration) Date() *Migration {
	return m.setColumnType("date")
}

// DateTime sets the data type of the latest column as `datetime`.
func (m *Migration) DateTime() *Migration {
	return m.setColumnType("dateTime")
}

// Time sets the data type of the latest column as `time`.
func (m *Migration) Time() *Migration {
	return m.setColumnType("time")
}

// Timestamp sets the data type of the latest column as `timestamp`.
func (m *Migration) Timestamp() *Migration {
	return m.setColumnType("timestamp")
}

// Year sets the data type of the latest column as `year`.
func (m *Migration) Year() *Migration {
	return m.setColumnType("year")
}

// Double sets the data type of the latest column as `double`.
func (m *Migration) Double(length ...int) *Migration {
	return m.setColumnType("double", length)
}

// Decimal sets the data type of the latest column as `decimal`.
//     .Decimal(2, 1)
func (m *Migration) Decimal(length ...int) *Migration {
	return m.setColumnType("decimal", length)
}

// Float sets the data type of the latest column as `float`.
//     .Float(2, 1)
//     .Float(1)
func (m *Migration) Float(length ...int) *Migration {
	return m.setColumnType("float", length)
}

// Enum sets the data type of the latest column as `enum`.
//     .Enum(1, 2, "A", "B")
func (m *Migration) Enum(types ...interface{}) *Migration {
	return m.setColumnType("enum", types)
}

// Set sets the data type of the latest column as `set`.
//     .Set(1, 2, "A", "B")
func (m *Migration) Set(types ...interface{}) *Migration {
	return m.setColumnType("set", types)
}

// Column creates a new column.
func (m *Migration) Column(name string) *Migration {
	m.columns = append(m.columns, column{name: name, defaultValue: false})
	return m
}

// InnoDB sets the engine type of the table as InnoDB.
func (m *Migration) InnoDB() *Migration {
	m.table.engineType = "innodb"
	return m
}

// MyISAM sets the engine type of the table as MyISAM.
func (m *Migration) MyISAM() *Migration {
	m.table.engineType = "myisam"
	return m
}

// Primary makes a column as a primary key when there're no arguments.
//     .Column("id").Primary()
// It groups columns as a primary key group when the argument is a string slice.
//     .Primary([]string{"id", "username"})
// It creates a naming primary key group when the first argument is a string, and the second argument is a string slice.
//     .Primary("pk_group", []string{"id", "username"})
func (m *Migration) Primary(args ...interface{}) *Migration {
	switch len(args) {
	// Primary()
	case 0:
		m.columns[len(m.columns)-1].primary = true

	// Primary([]string{"column1", "column2"})
	case 1:
		m.table.primaryKeys = append(m.table.primaryKeys, key{
			columns: args[0].([]string),
		})

	// Primary("primary_keys", []string{"column1", "column2"})
	case 2:
		m.table.primaryKeys = append(m.table.primaryKeys, key{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

// Unique makes a column as an unique key when there're no arguments.
//     .Column().Unique()
// It groups the columns as an single unique key group when the argument is a string slice.
//     .Unique([]string{"id", "username"})
// It creates a naming unique key group when the first argument is a string, and the second argument is a string slice.
//     .Unique("uk_group", []string{"id", "username"})
func (m *Migration) Unique(args ...interface{}) *Migration {
	switch len(args) {
	// Unique()
	case 0:
		m.columns[len(m.columns)-1].unique = true

	// Unique([]string{"column1", "column2"})
	case 1:
		m.table.uniqueKeys = append(m.table.uniqueKeys, key{
			columns: args[0].([]string),
		})

	// Unique("unique_keys", []string{"column1", "column2"})
	case 2:
		m.table.uniqueKeys = append(m.table.uniqueKeys, key{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

// Index makes a column as an index when there're no arguments.
//     .Column("id").Index()
// It groups columns as an index when the argument is a string slice.
//     .Index([]string{"id", "username"})
// It creates a naming index when the first argument is a string, and the second argument is a string slice.
//     .Index("ik_group", []string{"id", "username"})
func (m *Migration) Index(args ...interface{}) *Migration {
	switch len(args) {
	// Index()
	case 0:
		// Create an anonymous index group and put the latest column into the group.
		m.table.indexKeys = append(m.table.indexKeys, key{
			columns: []string{m.columns[len(m.columns)-1].name},
		})

	// Index([]string{"column1", "column2"})
	case 1:
		// Create the groups for each of the index.
		for _, v := range args[0].([]string) {
			m.table.indexKeys = append(m.table.indexKeys, key{
				name:    v,
				columns: []string{v},
			})
		}

	// Index("index_keys", []string{"column1", "column2"})
	case 2:
		m.table.indexKeys = append(m.table.indexKeys, key{
			name:    args[0].(string),
			columns: args[1].([]string),
		})
	}
	return m
}

// OnUpdate decides what to do to the foreign key parent when updating the child, for example: `NO ACTION`, `SET NULL`, etc.
func (m *Migration) OnUpdate(action string) *Migration {
	m.table.foreignKeys[len(m.table.foreignKeys)-1].onUpdate = action
	return m
}

// OnDelete decides what to do to the foreign key parent when deleting the child, for example: `NO ACTION`, `SET NULL`, etc.
func (m *Migration) OnDelete(action string) *Migration {
	m.table.foreignKeys[len(m.table.foreignKeys)-1].onDelete = action
	return m
}

// Foreign sets the latest column as a foreign key, or creates an anonymous/naming foreign key group.
// To make the last column as a foreign key, try the following code:
//     Columns("id").Foreign("users.id")
// To create an anonymous foreign key group:
//     .Foreign([]string{"id", "username"}, []string{"users.id", "users.username"})
// To create a naming foreign key group, here's how you do it:
//     .Foreign("fk_group", []string{"id", "username"}, []string{"users.id", "users.username"})
func (m *Migration) Foreign(args ...interface{}) *Migration {
	switch len(args) {
	// Foreign("users.id")
	case 1:
		// Split the string by the dot to get the target table from the `table.column` format.
		targetTable := strings.Split(args[0].(string), ".")[0]

		// Scan the current table foreign keys, to find if there's an exist target table or not.
		index := -1
		for k, v := range m.table.foreignKeys {
			// Split the name of the foreign key by the dot to get the name of the target table of the foreign key.
			tableName := strings.Split(v.targetColumns[0], ".")[0]
			if tableName == targetTable {
				index = k
				break
			}
		}
		// If there's the same target table, we merge the new foreign key to the group.
		if index != -1 {
			m.table.foreignKeys[index].columns = append(m.table.foreignKeys[index].columns, m.columns[len(m.columns)-1].name)
			m.table.foreignKeys[index].targetColumns = append(m.table.foreignKeys[index].targetColumns, args[0].(string))
			// Otherwise we create a new group.
		} else {
			m.table.foreignKeys = append(m.table.foreignKeys, key{
				columns:       []string{m.columns[len(m.columns)-1].name},
				targetColumns: []string{args[0].(string)},
			})
		}

	// Foreign([]string{"id", "password"}, []string{"users.id", "users.password"})
	case 2:
		m.table.foreignKeys = append(m.table.foreignKeys, key{
			columns:       args[0].([]string),
			targetColumns: args[1].([]string),
		})

	// Foreign("foreign_keys", []string{"id", "password"}, []string{"users.id", "users.password"})
	case 3:
		m.table.foreignKeys = append(m.table.foreignKeys, key{
			name:          args[0].(string),
			columns:       args[1].([]string),
			targetColumns: args[2].([]string),
		})
	}
	return m
}

// Nullable allows the latest column to be nullable.
func (m *Migration) Nullable() *Migration {
	m.columns[len(m.columns)-1].defaultValue = nil
	m.columns[len(m.columns)-1].nullable = true
	return m
}

// Unsigned allows the latest column to be unsigned.
func (m *Migration) Unsigned() *Migration {
	m.columns[len(m.columns)-1].unsigned = true
	return m
}

// Comment leaves the comment to the latest column.
func (m *Migration) Comment(text string) *Migration {
	m.columns[len(m.columns)-1].comment = text
	return m
}

// Default sets the default value for the latest column,
// it colud be nil, string or int and `CURRENT_TIMESTAMP` or `NOW()`.
func (m *Migration) Default(value interface{}) *Migration {
	m.columns[len(m.columns)-1].defaultValue = value
	return m
}

// AutoIncrement auto increments the latest column.
func (m *Migration) AutoIncrement() *Migration {
	m.columns[len(m.columns)-1].autoIncrement = true
	return m
}

// Table prepares the table to create.
func (m *Migration) Table(tableName string, comment ...string) *Migration {
	// Set the table name.
	m.table.name = tableName
	// And set the table comment if there's one.
	if len(comment) != 0 {
		m.table.comment = comment[0]
	}
	return m
}

// Create builds the query and execute it to create the table with the columns.
func (m *Migration) Create() (err error) {
	// Build the main query.
	query := m.tableBuilder()
	// Execute the main query to create the table and the columns.
	_, err = m.connection.Exec(query)
	// Save the last executed query.
	m.LastQuery = query
	// Clean the current table, columns data.
	m.clean()

	return
}

// Drop drops the specified tables.
func (m *Migration) Drop(tableNames ...string) error {
	// Drop each of the table.
	for _, name := range tableNames {
		// Build the query and execute to drop the table.
		query := fmt.Sprintf("DROP TABLE `%s`", name)
		_, err := m.connection.Exec(query)
		// Save the last executed query.
		m.LastQuery = query
		// Clean the current table, columns data.
		m.clean()
		// Return the error if any.
		if err != nil {
			return err
		}
	}
	return nil
}

// setColumnType sets the data type and the length of the latest column.
func (m *Migration) setColumnType(dataType string, arg ...interface{}) *Migration {
	m.columns[len(m.columns)-1].dataType = dataType
	// Store the length if any.
	if len(arg) == 1 {
		m.columns[len(m.columns)-1].length = arg[0]
	}
	return m
}

// tableBuilder builds the main table query.
func (m *Migration) tableBuilder() (query string) {
	var contentQuery string
	// Build the queries.
	columnQuery := m.columnBuilder()
	foreignQuery := m.indexBuilder("FOREIGN KEY")
	primaryQuery := m.indexBuilder("PRIMARY KEY")
	uniqueQuery := m.indexBuilder("UNIQUE KEY")
	indexQuery := m.indexBuilder("INDEX")
	// The main query.
	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` ", m.table.name)

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
		engineType = "innodb"
	}
	query += fmt.Sprintf("ENGINE=%s, ", strings.ToUpper(engineType))
	// Comment.
	if m.table.comment != "" {
		query += fmt.Sprintf("COMMENT='%s', ", m.table.comment)
	}
	// Remove the unnecessary comma and the space.
	query = trim(query)
	return
}

// clean cleans the previous table information.
func (m *Migration) clean() {
	m.table.comment = ""
	m.table.engineType = ""
	m.table.foreignKeys = []key{}
	m.table.indexKeys = []key{}
	m.table.primaryKeys = []key{}
	m.table.uniqueKeys = []key{}
	m.table.name = ""
	m.columns = []column{}
}

// indexBuilder builds the query for the indexes.
func (m *Migration) indexBuilder(indexName string) (query string) {
	var keys []key
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
			targetColumns = ""
			for _, c := range v.targetColumns {
				// Get the target table name from the target columns. (targetTable.targetColumn)
				targetTable = strings.Split(c, ".")[0]
				// Remove the table name in the column name and build the query.
				targetColumns += fmt.Sprintf("`%s`, ", strings.Split(c, ".")[1])
			}
			// Remove the unnecessary comma and the space.
			targetColumns = trim(targetColumns)
		}

		// The on update/delete actions.
		var onUpdate, onDelete string
		if v.onUpdate != "" {
			onUpdate = fmt.Sprintf(" ON UPDATE %s", v.onUpdate)
		}
		if v.onDelete != "" {
			onDelete = fmt.Sprintf(" ON DELETE %s", v.onDelete)
		}

		// Indexs without the group name.
		if v.name == "" && len(v.targetColumns) == 0 {
			query += fmt.Sprintf("%s (%s), ", indexName, columns)
			// Naming indexes.
		} else if v.name != "" && len(v.targetColumns) == 0 {
			query += fmt.Sprintf("%s `%s` (%s), ", indexName, v.name, columns)
			// Foreign keys without the group name.
		} else if v.name == "" && len(v.targetColumns) != 0 {
			query += fmt.Sprintf("%s (%s) REFERENCES `%s` (%s)%s%s, ", indexName, columns, targetTable, targetColumns, onUpdate, onDelete)
			// Foreign keys.
		} else if v.name != "" && len(v.targetColumns) != 0 {
			query += fmt.Sprintf("%s %s (%s) REFERENCES `%s` (%s)%s%s, ", indexName, v.name, columns, targetTable, targetColumns, onUpdate, onDelete)
		}
	}
	// Remove the unnecessary comma and the space.
	query = trim(query)
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
		// ENUM(1, 2, 3, 4)
		case []int:
			// Converts []int to a string and split by the comma.
			query += fmt.Sprintf("%s(%s) ", dataType, strings.Trim(strings.Join(strings.Split(fmt.Sprint(t), " "), ", "), "[]"))
		// ENUM("A", "B", "C")
		case []string:
			query += fmt.Sprintf("%s(%s) ", dataType, strings.Join(t, ", "))
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
			query += fmt.Sprintf("%s(%s) ", dataType, trim(options))
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
			if t == "CURRENT_TIMESTAMP" || t == "NOW()" {
				query += fmt.Sprintf("DEFAULT %s ", t)
			} else {
				query += fmt.Sprintf("DEFAULT '%s' ", t)
			}
		}

		// Keys.
		if v.primary {
			query += "PRIMARY KEY "
		}
		if v.unique {
			query += "UNIQUE "
		}
		if v.foreign != "" {
			m.table.foreignKeys = append(m.table.foreignKeys, key{
				columns:       []string{v.name},
				targetColumns: []string{v.foreign},
			})
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
