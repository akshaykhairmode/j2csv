package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/akshaykhairmode/j2csv/parser"
	"github.com/rs/zerolog"
)

const (
	arrayFilePath  = "test-files/array.json"
	objectFilePath = "test-files/object.txt"
)

func init() {
	// GenerateFile(true)
	// GenerateFile(false)
}

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
		jsonObj := fmt.Sprintf(`{"fname":"John","Age":%d,"Location":"Australia","lname":"doe","createdAt":%d,"updatedAt":%d}`, i, i, i)
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
		scanner := bufio.NewScanner(inp)
		decoder := json.NewDecoder(inp)
		p := parser.NewParser(scanner, csv.NewWriter(out), decoder, &zerolog.Logger{})
		p.ProcessArray("")
		inp.Reset()
		out.Reset()
	}
}

func BenchmarkParseArrayWithPool(b *testing.B) {

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
		scanner := bufio.NewScanner(inp)
		decoder := json.NewDecoder(inp)
		p := parser.NewParser(scanner, csv.NewWriter(out), decoder, &zerolog.Logger{}).EnablePool()
		p.ProcessArray("")
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
		inp := bytes.NewBuffer(dt)
		out := bytes.NewBuffer(nil)
		scanner := bufio.NewScanner(inp)
		decoder := json.NewDecoder(inp)
		p := parser.NewParser(scanner, csv.NewWriter(out), decoder, &zerolog.Logger{})
		p.ProcessObjects("")
		inp.Reset()
		out.Reset()
	}
}

func BenchmarkParseObjectsWithPool(b *testing.B) {

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
		inp := bytes.NewBuffer(dt)
		out := bytes.NewBuffer(nil)
		scanner := bufio.NewScanner(inp)
		decoder := json.NewDecoder(inp)
		p := parser.NewParser(scanner, csv.NewWriter(out), decoder, &zerolog.Logger{}).EnablePool()
		p.ProcessObjects("")
		inp.Reset()
		out.Reset()
	}
}
