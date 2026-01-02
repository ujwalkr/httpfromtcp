package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log/slog"
	"strconv"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateDone    parseState = "done"
	StateHeaders parseState = "headers"
	StateBody    parseState = "body"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parseState
	Body        string
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valStr, exists := headers.Get(name)

	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrorRequestInState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	// fmt.Printf("%s\n", string(b))
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 && string(httpParts[0]) != "HTTP" && string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool {
	val := getInt(r.Headers, "content-length", 0)
	return val > 0
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ErrorRequestInState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				r.state = StateDone
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			slog.Info("State#Done", "Body:", r.Body)
			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer
		default:
			panic("Something went wrogn")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		slog.Info("Read Request", "State", request.state)

		n, err := reader.Read(buf[bufLen:])

		if err != nil {
			return nil, err
		}

		bufLen += n
		fmt.Printf("%s\n", buf[:bufLen])
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
