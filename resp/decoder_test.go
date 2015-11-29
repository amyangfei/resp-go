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
