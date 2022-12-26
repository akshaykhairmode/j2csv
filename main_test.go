package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/akshaykhairmode/j2csv/parser"
	"github.com/rs/zerolog"
)

func TestGenerateFile(t *testing.T) {
	rowCount := 500
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

func BenchmarkParse(b *testing.B) {

	inp, err := os.Open("500")
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
		parser.NewParser(bufio.NewReader(inp), csv.NewWriter(out), &zerolog.Logger{}).Process("")
	}
}
