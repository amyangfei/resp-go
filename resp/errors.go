package resp

import (
	"errors"
)

var (
	// ErrCrlfNotFound is returnd when no \r\n is found in a read buffer
	// This error mpay happens when one independent request/reply is sepetated
	// into two or more TCP packet
	ErrCrlfNotFound = errors.New("CRLF not found")

	// ErrEmpayData is returnd is no data found before \r\n
	ErrEmptyData = errors.New("empay data before crlf")

	// ErrBulkendNotFound is returnd if read buffer is short than expected bulk
	// string length.
	// This error mpay happens when one independent request/reply is sepetated
	// into two or more TCP packet
	ErrBulkendNotFound = errors.New("data buffer short than bulk string length")

	// ErrInvalidHeader is returned when unknown data prefix is found
	ErrInvalidHeader = errors.New("invalid header")

	// ErrRespData is returned when data breaks the RESP
	ErrRespData = errors.New("invalid resp data")
)

func MaybeSegmentError(err error) bool {
	switch err {
	case ErrCrlfNotFound, ErrBulkendNotFound:
		return true
	default:
		return false
	}
}
