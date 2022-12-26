package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
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
	for i := 0; i < b.N; i++ {
		inp, err := os.Open("500")
		if err != nil {
			panic(err)
		}

		out, err := os.OpenFile("500.out", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		p := parser.NewParser(bufio.NewReader(inp), csv.NewWriter(out), &zerolog.Logger{})
		p.Process("")
		inp.Close()
		out.Close()
		os.Remove("500.out")
	}
}
