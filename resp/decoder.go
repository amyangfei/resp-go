package resp

import (
	"bytes"
	"errors"
	"strconv"
)

const (
	CR = '\r'
	LF = '\n'
)

type Decoder struct {
	src         []byte
	pos         int
	msgQ        []*Message
	msgStartPos int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		src: data,
		pos: 0,
	}
}

func (d *Decoder) next(bufmsg *Message) error {
	lineType, line, err := ParseLine(d.src[d.pos:], -1)
	if err != nil {
		return err
	}
	var msg *Message
	if bufmsg != nil {
		msg = bufmsg
	} else {
		msg = &Message{}
	}

	switch lineType {
	case StringHeader:
		msg.Type = StringHeader
		msg.Status = string(line)
		d.UpdatePos(true, len(line))
		d.AppendNewMsg(msg)
		return nil
	case ErrorHeader:
		msg.Type = ErrorHeader
		msg.Error = errors.New(string(line))
		d.UpdatePos(true, len(line))
		d.AppendNewMsg(msg)
		return nil
	case IntegerHeader:
		msg.Type = IntegerHeader
		msg.Integer, err = strconv.ParseInt(string(line), 10, 64)
		d.UpdatePos(true, len(line))
		if err != nil {
			d.UpdateStartPos(d.pos)
			return err
		}
		d.AppendNewMsg(msg)
		return nil
	case BulkHeader:
		var msgLen int
		if msgLen, err = strconv.Atoi(string(line)); err != nil {
			return err
		}
		// RESP Bulk Strings can also be used in order to signal non-existence
		// of a value, which is known as a Null Bulk String
		if msgLen < 0 {
			msg.Type = BulkHeader
			msg.IsNil = true
			d.UpdatePos(true, len(line))
			d.AppendNewMsg(msg)
			return nil
		}
		d.UpdatePos(true, len(line))
		_, bulkstr, err := ParseLine(d.src[d.pos:], msgLen)
		d.UpdatePos(false, len(bulkstr))
		if err != nil {
			d.UpdateStartPos(d.pos)
			return err
		}
		msg.Type = BulkHeader
		msg.Bytes = bulkstr
		if bufmsg == nil {
			d.AppendNewMsg(msg)
		}
		return nil
	case ArrayHeader:
		var arrLen int
		if arrLen, err = strconv.Atoi(string(line)); err != nil {
			return err
		}
		// The concept of Null Array exists as well, and is an alternative way
		// to specify a Null value (usually the Null Bulk String is used, but
		// for historical reasons we have two formats).
		if arrLen < 0 {
			msg.Type = ArrayHeader
			msg.IsNil = true
			d.UpdatePos(true, len(line))
			d.AppendNewMsg(msg)
			return nil
		}
		msg.Array = make([]*Message, arrLen)
		d.UpdatePos(true, len(line))
		for i := 0; i < arrLen; i++ {
			msg.Array[i] = &Message{}
			if err = d.next(msg.Array[i]); err != nil {
				return err
			}
		}
		if bufmsg == nil {
			d.AppendNewMsg(msg)
		}
		return nil
	}
	return ErrInvalidHeader
}

func (d *Decoder) UpdateStartPos(pos int) {
	d.msgStartPos = pos
}

func (d *Decoder) AppendNewMsg(msg *Message) {
	d.msgQ = append(d.msgQ, msg)
	d.UpdateStartPos(d.pos)
}

func (d *Decoder) UpdatePos(hasHeader bool, msgLen int) {
	d.pos += msgLen + 2
	if hasHeader {
		d.pos += 1
	}
}

func Decode(data []byte) ([]*Message, int, error) {
	d := NewDecoder(data)
	for d.msgStartPos < len(data) {
		err := d.next(nil)
		if err != nil {
			return d.msgQ, d.msgStartPos, err
		}
	}
	return d.msgQ, d.msgStartPos, nil
}

// ParseLine find the CRLF and return lineType and line data.
// If readLen is no less than zero, we don't parse the header type and read
// readLen bytes directly and check if \r\n follows.
func ParseLine(data []byte, readLen int) (lineType byte, line []byte, err error) {
	if readLen >= 0 {
		if len(data) < readLen+2 {
			return 0, nil, ErrBulkNotEnough
		} else {
			if data[readLen] != CR || data[readLen+1] != LF {
				return 0, nil, ErrRespData
			}
			return 0, data[:readLen], nil
		}
	} else {
		i := bytes.IndexByte(data, LF)
		if i < 0 || i == 0 {
			return 0, nil, ErrCrlfNotFound
		} else if data[i-1] != CR {
			return 0, nil, ErrCrlfNotFound
		} else if i == 1 {
			return 0, nil, ErrEmpayData
		}
		return data[0], data[1 : i-1], nil
	}
}
