package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var db *DB

func TestMain(t *testing.T) {
	assert := assert.New(t)

	var err error
	db, err = New("root:root@/test?charset=utf8")

	assert.NoError(err)


	subQuery := db.SubQuery("KeepRows")
	subQuery.GroupBy("Col1, Col2, Col3").Get("MyTable", "MIN(RowId) as RowId, Col1, Col2, Col3")
	db.Join(subQuery, "MyTable.RowId = KeepRows.RowId", "LEFT OUTER").Where("KeepRows.RowId", nil).Delete("MyTable")
}





DELETE FROM MyTable
LEFT OUTER JOIN (
   SELECT MIN(RowId) as RowId, Col1, Col2, Col3
   FROM MyTable 
   GROUP BY Col1, Col2, Col3
) as KeepRows ON
   MyTable.RowId = KeepRows.RowId
WHERE
   KeepRows.RowId IS NULL