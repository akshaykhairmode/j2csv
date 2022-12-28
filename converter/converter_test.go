package converter

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/rs/zerolog"
)

func TestCopy(t *testing.T) {

	dta := make([]byte, 10)
	for i := 0; i < len(dta); i++ {
		dta[i] = byte(i)
	}

	r := bufio.NewReader(bytes.NewReader(dta))

	cc := New(r, 0, &zerolog.Logger{})
	data, err := io.ReadAll(cc)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(dta, data) {
		t.Logf("Expected : %s, Got : %s", dta, data)
		t.Errorf("data does not match")
	}
}
