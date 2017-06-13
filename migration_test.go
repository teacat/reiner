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
	assert := assert.New(t)
	migration.
		Drop("test_table")
	assert.Equal("DROP TABLE `test_table`", migration.LastQuery)
}

func TestMigrationInsert(t *testing.T) {

}

func TestMigrationDataTypes(t *testing.T) {
	assert := assert.New(t)
	migration.
		Column("test").TinyInt(1).
		Column("test2").SmallInt(1).
		Column("test3").MediumInt(1).
		Column("test4").Int(1).
		Column("test5").BigInt(1).
		Column("test6").Char(1).
		Column("test7").Varchar(1).
		Column("test8").Binary(1).
		Column("test9").VarBinary(1).
		Column("test10").Bit(1).
		Column("test11").TinyText().
		Column("test12").Text().
		Column("test13").MediumText().
		Column("test14").LongText().
		Column("test15").TinyBlob().
		Column("test16").Blob().
		Column("test17").MediumBlob().
		Column("test18").LongBlob().
		Column("test19").Date().
		Column("test20").DateTime().
		Column("test21").Time().
		Column("test22").Timestamp().
		Column("test23").Year().
		Column("test24").Double(2, 1).
		Column("test25").Decimal(2, 1).
		Column("test26").Float(2, 1).
		Column("test27").Float(1).
		Column("test28").Enum("1", "2", "3", "A", "B", "C").
		Column("test29").Set("1", "2", "3", "A", "B", "C").
		Create("test_table1")
	assert.Equal("CREATE TABLE `test_table1` (`test` TINYINT(1) NOT NULL , `test2` SMALLINT(1) NOT NULL , `test3` MEDIUMINT(1) NOT NULL , `test4` INT(1) NOT NULL , `test5` BIGINT(1) NOT NULL , `test6` CHAR(1) NOT NULL , `test7` VARCHAR(1) NOT NULL , `test8` BINARY(1) NOT NULL , `test9` VARBINARY(1) NOT NULL , `test10` BIT(1) NOT NULL , `test11` TINYTEXT NOT NULL , `test12` TEXT NOT NULL , `test13` MEDIUMTEXT NOT NULL , `test14` LONGTEXT NOT NULL , `test15` TINYBLOB NOT NULL , `test16` BLOB NOT NULL , `test17` MEDIUMBLOB NOT NULL , `test18` LONGBLOB NOT NULL , `test19` DATE NOT NULL , `test20` DATETIME NOT NULL , `test21` TIME NOT NULL , `test22` TIMESTAMP NOT NULL , `test23` YEAR NOT NULL , `test24` DOUBLE(2, 1) NOT NULL , `test25` DECIMAL(2, 1) NOT NULL , `test26` FLOAT(2, 1) NOT NULL , `test27` FLOAT(1) NOT NULL , `test28` ENUM('1', '2', '3', 'A', 'B', 'C') NOT NULL , `test29` SET('1', '2', '3', 'A', 'B', 'C') NOT NULL) ENGINE=INNODB", migration.LastQuery)
}

func TestMigrationTableType(t *testing.T) {
	assert := assert.New(t)
	migration.Column("test").Varchar(32).MyISAM().Create("test_myisam_table")
	assert.Equal("CREATE TABLE `test_myisam_table` (`test` VARCHAR(32) NOT NULL) ENGINE=MYISAM", migration.LastQuery)
	migration.Column("test").Varchar(32).InnoDB().Create("test_innodb_table")
	assert.Equal("CREATE TABLE `test_innodb_table` (`test` VARCHAR(32) NOT NULL) ENGINE=INNODB", migration.LastQuery)
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
