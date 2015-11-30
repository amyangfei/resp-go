## resp-go

[![Build Status](https://travis-ci.org/amyangfei/resp-go.svg?branch=master)](https://travis-ci.org/amyangfei/resp-go)

This package is used for decoding data in RESP format from raw byte array.

The resp-go supports continuous RESP format data which is often seen in redis pipeline scene. Besides it also supports decoding in lazy mode, which means the package can cache incomplete RESP data and wait for the rest data flow.

## Installation

Use `go get`, it's simple

```bash
go get github.com/amyangfei/resp-go/resp
```

## Usage
resp provides an `Decode()` function that takes a byte array contains one or more RESP messages and returns an array of `*Message`, the latest consumption postion and error. The `Message` represents a RESP object.

For example, we have a byte array of four RESP messages, which represents a pipeline operation in redis:

```bash
>>> MULTI
>>> GET a
>>> LRANGE 0 -1
>>> EXEC
```

We can `Decode` the resp data array as following:

```go
encoded := []byte("*1\r\n$5\r\nMULTI\r\n" +
    "*2\r\n$3\r\nGET\r\n$1\r\na\r\n" +
    "*4\r\n$6\r\nLRANGE\r\n$1\r\nl\r\n$1\r\n0\r\n$2\r\n-1\r\n" +
    "*1\r\n$4\r\nEXEC\r\n")
// msgQ is an  array with four *Message object
// pos equals to len(encoded) which means we have consumed all RESP data in encoded
// and later consumption will start from pos.
msgQ, pos, err := resp.Decode(encoded)
```

## Acknowledgment
This package is inspired by [xiam/resp](https://github.com/xiam/resp)

If you want to encode/decode RESP data one by one or from a stream, it's better to use `xiam/resp`
