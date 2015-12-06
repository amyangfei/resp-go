package resp

import (
	"bytes"
	"errors"
	"testing"
)

var (
	errTestFailed    = errors.New("Test failed")
	errErrorExpected = errors.New("An error expected")
)

func TestEncodeString(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal("Foo"); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("$3\r\nFoo\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeError(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal(errors.New("Fatal error")); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("-Fatal error\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeInteger(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal(123); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(":123\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeBulk(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal([]byte("♥")); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("$3\r\n♥\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeArray(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal([]interface{}{"Foo", "Bar"}); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("*2\r\n+Foo\r\n+Bar\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeMixedArray(t *testing.T) {
	var buf []byte
	var err error

	mixed := []interface{}{
		[]interface{}{
			1, 2, 3,
		},
		[]interface{}{
			[]byte("Foo"),
			errors.New("Bar"),
			"Baz",
		},
	}

	if buf, err = Marshal(mixed); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*3\r\n$3\r\nFoo\r\n-Bar\r\n+Baz\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeZeroArray(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal([]interface{}{}); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("*0\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}

func TestEncodeNil(t *testing.T) {
	var buf []byte
	var err error

	if buf, err = Marshal(nil); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("-1\r\n")) == false {
		t.Fatal(errTestFailed)
	}
}
