package resp

import (
	"bytes"
	"errors"
	"testing"
)

func decodeToMsg(buf []byte) (*Message, error) {
	msg, pos, err := Decode(buf)
	if err != nil {
		return nil, err
	} else if len(msg) != 1 {
		return nil, errors.New("should contains only one message")
	} else if pos != len(buf) {
		return nil, errors.New("wrong consume position")
	}
	return msg[0], nil
}

func decodeEncodeTest(buf []byte, target string, t *testing.T) {
	msg, err := decodeToMsg(buf)
	if err != nil {
		t.Error(err)
	}
	newbuf, err := Marshal(msg)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(newbuf, []byte(target)) == false {
		t.Logf("%v newbuf is: %v, should be: %v", string(newbuf), newbuf, []byte(target))
		t.Error(errTestFailed)
	}
}

func TestEncodeDecodeString(t *testing.T) {
	var buf []byte
	var err error
	var target string = "$3\r\nFoo\r\n"

	if buf, err = Marshal("Foo"); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeError(t *testing.T) {
	var buf []byte
	var err error
	var target string = "-Fatal error\r\n"

	if buf, err = Marshal(errors.New("Fatal error")); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeInteger(t *testing.T) {
	var buf []byte
	var err error
	var target string = ":123\r\n"

	if buf, err = Marshal(123); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeBulk(t *testing.T) {
	var buf []byte
	var err error
	var target string = "$3\r\n♥\r\n"

	if buf, err = Marshal([]byte("♥")); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeArray(t *testing.T) {
	var buf []byte
	var err error
	var target string = "*2\r\n+Foo\r\n+Bar\r\n"

	if buf, err = Marshal([]interface{}{"Foo", "Bar"}); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeMixedArray(t *testing.T) {
	var buf []byte
	var err error
	var target string = "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*3\r\n$3\r\nFoo\r\n-Bar\r\n+Baz\r\n"

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

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeZeroArray(t *testing.T) {
	var buf []byte
	var err error
	var target string = "*0\r\n"

	if buf, err = Marshal([]interface{}{}); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte(target)) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, target, t)
}

func TestEncodeDecodeNil(t *testing.T) {
	var buf []byte
	var err error
	var m *Message

	m = new(Message)
	m.SetNil()
	m.Type = BulkHeader

	if buf, err = Marshal(m); err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(buf, []byte("$-1\r\n")) == false {
		t.Fatal(errTestFailed)
	}

	decodeEncodeTest(buf, "$-1\r\n", t)
}
