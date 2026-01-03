package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type Reponse struct{}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) *headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error")
	default:
		return fmt.Errorf("unrecognized status code")
	}
	statusLine = fmt.Append(statusLine, "\r\n")
	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	b := []byte{}
	headers.ForEach(func(k, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}
