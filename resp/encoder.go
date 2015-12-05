package resp

import (
	"io"
	"sync"
)

const digitbuflen = 20

var (
	endOfLine  = []byte{'\r', '\n'}
	encoderNil = []byte("$-1\r\n")
	digits     = []byte("0123456789")
)

func intToBytes(v int) []byte {
	buf := make([]byte, digitbuflen)

	i := len(buf)

	for v >= 10 {
		i--
		buf[i] = digits[v%10]
		v = v / 10
	}

	i--
	buf[i] = digits[v%10]

	return buf[i:]
}

// Encoder provides the Encode() method for encoding directly to an io.Writer.
type Encoder struct {
	w   io.Writer
	buf []byte
	mu  *sync.Mutex
}

// NewEncoder creates and returns a *Encoder value with the given io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{
		w:   w,
		buf: []byte{},
		mu:  new(sync.Mutex),
	}
	return e
}

// Encode marshals the given argument into a RESP message and pushes the output
// to the given writer.
func (e *Encoder) Encode(v interface{}) error {
	return e.writeEncoded(e.w, v)
}

func (e *Encoder) writeEncoded(w io.Writer, data interface{}) (err error) {

	var b []byte

	switch v := data.(type) {

	case []byte:
		n := intToBytes(len(v))

		b = make([]byte, 0, 1+len(n)+2+len(v)+2)

		b = append(b, BulkHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)
		b = append(b, v...)
		b = append(b, endOfLine...)

	case string:
		q := []byte(v)

		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, StringHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case error:
		q := []byte(v.Error())

		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, ErrorHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case int:
		q := intToBytes(int(v))
		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, IntegerHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

	case [][]byte:
		n := intToBytes(len(v))

		b = make([]byte, 0, 1+len(n)+2)
		b = append(b, ArrayHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)

		for i := range v {
			q := intToBytes(len(v[i]))

			z := make([]byte, 0, 1+len(q)+2+len(v[i])+2)

			z = append(z, BulkHeader)
			z = append(z, q...)
			z = append(z, endOfLine...)
			z = append(z, v[i]...)
			z = append(z, endOfLine...)

			b = append(b, z...)
		}

	case []string:
		q := intToBytes(len(v))

		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, ArrayHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

		for i := range v {
			p := []byte(v[i])

			z := make([]byte, 0, 1+len(p)+2)
			z = append(z, StringHeader)
			z = append(z, p...)
			z = append(z, endOfLine...)

			b = append(b, z...)
		}

	case []int:
		n := intToBytes(len(v))

		b = make([]byte, 0, 1+len(n)+2)
		b = append(b, ArrayHeader)
		b = append(b, n...)
		b = append(b, endOfLine...)

		for i := range v {
			m := intToBytes(v[i])

			z := make([]byte, 0, 1+len(m)+2)
			z = append(z, IntegerHeader)
			z = append(z, m...)
			z = append(z, endOfLine...)

			b = append(b, z...)
		}

	case []interface{}:
		q := intToBytes(len(v))

		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, ArrayHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

		e.buf = append(e.buf, b...)

		if w != nil {
			e.mu.Lock()
			w.Write(e.buf)
			e.buf = []byte{}
			e.mu.Unlock()
		}

		for i := range v {
			if err = e.writeEncoded(w, v[i]); err != nil {
				return err
			}
		}

		return nil

	case *Message:
		switch v.Type {
		case ErrorHeader:
			return e.writeEncoded(w, v.Error)
		case IntegerHeader:
			return e.writeEncoded(w, int(v.Integer))
		case BulkHeader:
			// case for "$-1\r\n"
			if v.IsNil {
				return e.writeEncoded(w, nil)
			}
			return e.writeEncoded(w, v.Bytes)
		case StringHeader:
			return e.writeEncoded(w, v.Status)
		case ArrayHeader:
			return e.writeEncoded(w, v.Array)
		default:
			return ErrInvalidHeader
		}

	case []*Message:
		q := intToBytes(len(v))

		b = make([]byte, 0, 1+len(q)+2)
		b = append(b, ArrayHeader)
		b = append(b, q...)
		b = append(b, endOfLine...)

		e.buf = append(e.buf, b...)
		b = []byte("")

		if w != nil {
			e.mu.Lock()
			w.Write(e.buf)
			e.buf = []byte{}
			e.mu.Unlock()
		}

		for _, msg := range v {
			e.writeEncoded(w, msg)
		}

	case nil:
		b = make([]byte, 0, len(encoderNil))
		b = append(b, encoderNil...)

	default:
		return ErrInvalidInput
	}

	e.buf = append(e.buf, b...)

	if w != nil {
		e.mu.Lock()
		w.Write(e.buf)
		e.buf = []byte{}
		e.mu.Unlock()
	}

	return nil
}

// Marshal returns the RESP encoding of v. At this moment, it only works with
// string, int, []byte, nil and []interface{} types.
func Marshal(v interface{}) ([]byte, error) {

	switch t := v.(type) {
	case string:
		// If the user sends a string, we convert it to byte to make it binary
		// safe.
		v = []byte(t)
	}

	e := NewEncoder(nil)

	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return e.buf, nil
}
