package main

import (
	"os"
	"strconv"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	rowCount := 50000
	f, err := os.OpenFile(strconv.Itoa(rowCount)+".csv", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	jsonObj := `{"fname":"John","Age":10,"Location":"Australia","lname":"doe","createdAt":1671888271,"updatedAt":1671888271}`

	f.WriteString("[")
	for i := 0; i < rowCount; i++ {
		f.WriteString(jsonObj)
		if i < rowCount-1 {
			f.WriteString(",")
			f.WriteString("\n")
		}
	}
	f.WriteString("]")
}
