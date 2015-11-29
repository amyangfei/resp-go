package resp

const (
	// StringHeader is the header used to prefix simple strings (or status
	// messages). String messages are not binary safe.
	StringHeader = '+'
	// ErrorHeader is the header used to prefix error messages.
	ErrorHeader = '-'
	// IntegerHeader is the header used to prefix integers.
	IntegerHeader = ':'
	// BulkHeader is the header used to prefix binary safe messages.
	BulkHeader = '$'
	// ArrayHeader is the header used to prefix an array of messages.
	ArrayHeader = '*'
)

// Message is a representation of a RESP message.
type Message struct {
	Error   error
	Integer int64
	Bytes   []byte
	Status  string
	Array   []*Message
	IsNil   bool
	Type    byte
}

// SetStatus sets a message of type status.
func (m *Message) SetStatus(s string) {
	m.Type = StringHeader
	m.Status = s
}

// SetError sets a message of type error.
func (m *Message) SetError(e error) {
	m.Type = ErrorHeader
	m.Error = e
}

// SetInteger sets a message of type integer.
func (m *Message) SetInteger(i int64) {
	m.Type = IntegerHeader
	m.Integer = i
}

// SetBytes sets a binary safe message.
func (m *Message) SetBytes(b []byte) {
	m.Type = BulkHeader
	m.Bytes = b
}

// SetArray sets a message of type array.
func (m *Message) SetArray(a []*Message) {
	m.Type = ArrayHeader
	m.Array = a
}

// SetNil sets a message as nil.
func (m *Message) SetNil() {
	m.Type = 0
	m.IsNil = true
}

// Interface returns the current value of the message, as an interface.
func (m Message) Interface() interface{} {
	switch m.Type {
	case ErrorHeader:
		return m.Error
	case IntegerHeader:
		return m.Integer
	case BulkHeader:
		return m.Bytes
	case StringHeader:
		return m.Status
	case ArrayHeader:
		return m.Array
	}
	return nil
}
