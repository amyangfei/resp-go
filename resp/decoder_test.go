package resp

import (
	"bytes"
	"testing"
)

func TestDecodeString(t *testing.T) {
	var encoded []byte
	var msgQ []*Message
	var pos int
	var err error

	// Simple "OK" string
	encoded = []byte("+OK\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Status != "OK" {
		t.Error("error string result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// String with a special character
	encoded = []byte("+OK\r +OK\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Status != "OK\r +OK" {
		t.Error("error string result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// several consecutive simple strings
	encoded = []byte("+OK\r\n+QUEUED\r\n+QUEUED\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 3 {
		t.Error("should contains three messages")
	} else if msgQ[0].Status != "OK" || msgQ[1].Status != "QUEUED" || msgQ[2].Status != "QUEUED" {
		t.Error("error string result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// simple string not complete
	encoded = []byte("+A quite long string")
	msgQ, pos, err = Decode(encoded)
	if err != ErrCrlfNotFound {
		t.Errorf("should return ErrCrlfNotFound error, not: %v", err)
	} else if len(msgQ) != 0 {
		t.Error("should contains no message")
	} else if pos != 0 {
		t.Error("error new consume pos")
	}
	encoded = append(encoded, []byte(" meets end\r\n")...)
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Status != "A quite long string meets end" {
		t.Error("error string result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// several consecutive simple strings, but the last one is not complete
	encoded = []byte("+OK\r\n+QUEUED\r\n+QUEUED")
	msgQ, pos, err = Decode(encoded)
	if err != ErrCrlfNotFound {
		t.Errorf("should return ErrCrlfNotFound error, not: %v", err)
	} else if len(msgQ) != 2 {
		t.Error("should contains three messages")
	} else if msgQ[0].Status != "OK" || msgQ[1].Status != "QUEUED" {
		t.Error("error string result")
	} else if pos != len(encoded)-7 {
		t.Error("error new consume pos")
	}
}

func TestDecodeError(t *testing.T) {
	var encoded []byte
	var msgQ []*Message
	var pos int
	var err error

	// Simple error message
	encoded = []byte("-ERR wrong number of arguments for 'hgetall' command\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Error.Error() != "ERR wrong number of arguments for 'hgetall' command" {
		t.Error("error of error result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}
}

func TestDecodeInteger(t *testing.T) {
	var encoded []byte
	var msgQ []*Message
	var pos int
	var err error

	// Positive integer
	encoded = []byte(":100\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Integer != 100 {
		t.Error("error of integer result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Negative integer
	encoded = []byte(":-100\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Integer != -100 {
		t.Error("error of integer result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// invalid integer
	encoded = []byte(":10.1\r\n")
	msgQ, pos, err = Decode(encoded)
	if err == nil {
		t.Error("error expected")
	} else if len(msgQ) != 0 {
		t.Error("should contains no message")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// consecutive integer
	encoded = []byte(":10\r\n:11.1\r\n:12\r\n")
	msgQ, pos, err = Decode(encoded)
	if err == nil {
		t.Error("error expected")
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Integer != 10 {
		t.Error("error of integer result")
	} else if pos != len(":10\r\n:11.1\r\n") {
		t.Error("error new consume pos")
	}
	encoded = encoded[pos:]
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if msgQ[0].Integer != 12 {
		t.Error("error of integer result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}
}

func TestDecodeBulk(t *testing.T) {
	var encoded []byte
	var msgQ []*Message
	var pos int
	var err error

	// string "hello"
	encoded = []byte("$5\r\nhello\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if bytes.Equal(msgQ[0].Bytes, []byte("hello")) == false {
		t.Error("error bulk result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// string "$10\r\nhello\r\ngo\r\n"
	encoded = []byte("$9\r\nhello\r\ngo\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if bytes.Equal(msgQ[0].Bytes, []byte("hello\r\ngo")) == false {
		t.Error("error bulk result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// empty string
	encoded = []byte("$0\r\n\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if bytes.Equal(msgQ[0].Bytes, []byte("")) == false {
		t.Error("error bulk result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Null bulk string
	encoded = []byte("$-1\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if !msgQ[0].IsNil {
		t.Error("error bulk result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// UTF-8 string
	encoded = []byte("$3\r\n✓\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if bytes.Equal(msgQ[0].Bytes, []byte("✓")) == false {
		t.Error("error bulk result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}
}

func TestArrayDecode(t *testing.T) {
	var encoded []byte
	var msgQ []*Message
	var pos int
	var err error

	// Zero elements
	encoded = []byte("*0\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 0 {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Null array
	encoded = []byte("*-1\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if !msgQ[0].IsNil {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Array with two bulk string
	encoded = []byte("*2\r\n$3\r\nget\r\n$5\r\nhello\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 2 {
		t.Error("error array length")
	} else if string(msgQ[0].Array[0].Bytes) != "get" ||
		string(msgQ[0].Array[1].Bytes) != "hello" {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Array with three integers
	encoded = []byte("*3\r\n:11\r\n:12\r\n:13\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 3 {
		t.Error("error array length")
	} else if msgQ[0].Array[0].Integer != 11 ||
		msgQ[0].Array[1].Integer != 12 ||
		msgQ[0].Array[2].Integer != 13 {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Array with two simple string
	encoded = []byte("*2\r\n+OK\r\n+OK\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 2 {
		t.Error("error array length")
	} else if msgQ[0].Array[0].Status != "OK" || msgQ[0].Array[1].Status != "OK" {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Array of two integers and two bulk string
	encoded = []byte("*4\r\n:11\r\n$5\r\nhello\r\n$2\r\ngo\r\n:13\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 4 {
		t.Error("error array length")
	} else if msgQ[0].Array[0].Integer != 11 ||
		string(msgQ[0].Array[1].Bytes) != "hello" ||
		string(msgQ[0].Array[2].Bytes) != "go" ||
		msgQ[0].Array[3].Integer != 13 {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	}

	// Array of two arrays
	encoded = []byte("*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+foo\r\n-bar\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 2 {
		t.Error("error array length")
	} else if len(msgQ[0].Array[0].Array) != 3 || len(msgQ[0].Array[1].Array) != 2 {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	} else {
		arr1 := msgQ[0].Array[0]
		arr2 := msgQ[0].Array[1]
		if arr1.Array[0].Integer != 1 || arr1.Array[1].Integer != 2 || arr1.Array[2].Integer != 3 {
			t.Error("error array result")
		}
		if arr2.Array[0].Status != "foo" || arr2.Array[1].Error.Error() != "bar" {
			t.Error("error array result")
		}
	}

	// Array of a bulk string and an array
	encoded = []byte("*2\r\n$1\r\n0\r\n*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 2 {
		t.Error("error array length")
	} else if msgQ[0].Array[0].Type != BulkHeader || len(msgQ[0].Array[1].Array) != 2 {
		t.Error("error array result")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	} else {
		arr1 := msgQ[0].Array[0]
		arr2 := msgQ[0].Array[1]
		if string(arr1.Bytes) != "0" {
			t.Error("error array result")
		}
		if string(arr2.Array[0].Bytes) != "foo" || string(arr2.Array[1].Bytes) != "bar" {
			t.Error("error array result")
		}
	}

	// Array with nil
	encoded = []byte("*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n")
	msgQ, pos, err = Decode(encoded)
	if err != nil {
		t.Error(err)
	} else if len(msgQ) != 1 {
		t.Error("should contains one message")
	} else if len(msgQ[0].Array) != 3 {
		t.Error("error array length")
	} else if pos != len(encoded) {
		t.Error("error new consume pos")
	} else {
		if string(msgQ[0].Array[0].Bytes) != "foo" {
			t.Error("error array result")
		}
		if !msgQ[0].Array[1].IsNil {
			t.Error("error array result")
		}
		if string(msgQ[0].Array[2].Bytes) != "bar" {
			t.Error("error array result")
		}
	}
}
