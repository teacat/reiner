package reiner

import "testing"

func BenchmarkInsert(b *testing.B) {
	for i := 0; i < 1000; i++ {
		rw.Table("Users").Insert(map[string]interface{}{
			"Username": i,
			"Password": "test",
			"Age":      64,
		})
	}

}
