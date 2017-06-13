package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var migration *Migration

func TestMigrationMain(t *testing.T) {
	migration = db.Migration()
}

func TestMigrationBasic(t *testing.T) {
	assert := assert.New(t)
	migration.
		Column("test").Varchar(32).Primary().
		Create("test_table")
	assert.Equal("CREATE TABLE `test_table` (`test` VARCHAR(32) NOT NULL PRIMARY KEY) ENGINE=INNODB", migration.LastQuery)
}

func TestMigrationDrop(t *testing.T) {

}

func TestMigrationInsert(t *testing.T) {

}

func TestMigrationDataTypes(t *testing.T) {

}

func TestMigrationTableType(t *testing.T) {

}

func TestMigrationDefault(t *testing.T) {

}

func TestMigrationNullable(t *testing.T) {

}

func TestMigrationUnsigned(t *testing.T) {

}

func TestMigrationAutoIncrement(t *testing.T) {

}

func TestMigrationComment(t *testing.T) {

}

func TestMigrationTableComment(t *testing.T) {

}

func TestMigrationPrimaryKey(t *testing.T) {

}

func TestMigrationNamingPrimaryKey(t *testing.T) {

}

func TestMigrationMultiPrimaryKey(t *testing.T) {

}

func TestMigrationUniqueKey(t *testing.T) {

}

func TestMigrationNamingUniqueKey(t *testing.T) {

}

func TestMigrationMultiUniqueKey(t *testing.T) {

}

func TestMigrationAnonymousIndexKey(t *testing.T) {

}

func TestMigrationNamingIndexKey(t *testing.T) {

}

func TestMigrationMixedKeys(t *testing.T) {

}

func TestMigrationForeignKey(t *testing.T) {

}

func TestMigrationMultipleForeignKey(t *testing.T) {

}

func TestMigrationNamingForeignKey(t *testing.T) {

}
