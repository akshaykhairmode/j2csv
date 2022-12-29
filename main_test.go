package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

const (
	arrayFilePath  = "test-files/array.json"
	objectFilePath = "test-files/object.txt"
)

func TestGenerate(t *testing.T) {
	GenerateFile(true)
	GenerateFile(false)
}

func GenerateFile(isArray bool) {
	rowCount := 5000
	var fpath string

	if isArray {
		fpath = arrayFilePath
	} else {
		fpath = objectFilePath
	}

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if isArray {
		f.WriteString("[")
	}

	for i := 0; i < rowCount; i++ {
		var jsonObj string
		if isArray {
			jsonObj = fmt.Sprintf(`{"fname":"John","Age":%d,"Location":"Australia","lname":"doe","createdAt":%d,"updatedAt":%d}`, 1672325049, 1672325049, 1672325049)
		} else {
			jsonObj = fmt.Sprintf(`
			//This is single line comment
			{
				"fname": "John",
				"Age": %d,
				"Location": "Australia",
				"lname": "doe",
				"createdAt": %d,
				"updatedAt": %d
			  } /* this is multi
			   line comment */ `, i, i, i)
		}

		f.WriteString(jsonObj)
		if i < rowCount-1 {
			if isArray {
				f.WriteString(",")
			}
			f.WriteString("\n")
		}
	}
	if isArray {
		f.WriteString("]")
	}
}

func BenchmarkParseArray(b *testing.B) {

	inp, err := os.Open(arrayFilePath)
	if err != nil {
		panic(err)
	}
	dt, err := io.ReadAll(inp)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inp := bytes.NewBuffer(dt)
		out := bytes.NewBuffer(nil)
		processArray(csv.NewWriter(out), inp, &zerolog.Logger{}, "")
		inp.Reset()
		out.Reset()
	}
}

func BenchmarkParseObjects(b *testing.B) {

	inp, err := os.Open(objectFilePath)
	if err != nil {
		panic(err)
	}
	dt, err := io.ReadAll(inp)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inp := bufio.NewReader(bytes.NewBuffer(dt))
		out := bytes.NewBuffer(nil)
		processObjects(csv.NewWriter(out), inp, &zerolog.Logger{}, "")
		out.Reset()
	}
}
