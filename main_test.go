package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	rowCount := 5000000
	f, err := os.OpenFile(strconv.Itoa(rowCount), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	f.WriteString("[")
	for i := 0; i < rowCount; i++ {
		jsonObj := fmt.Sprintf(`{"fname":"John","Age":%d,"Location":"Australia","lname":"doe","createdAt":%d,"updatedAt":%d}`, i, i, i)
		f.WriteString(jsonObj)
		if i < rowCount-1 {
			f.WriteString(",")
			f.WriteString("\n")
		}
	}
	f.WriteString("]")
}
