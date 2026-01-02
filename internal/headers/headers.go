package headers

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

var rn = []byte("\r\n")

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}
		if !found {
			return false
		}
	}

	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return string(name), string(value), nil
}

type Headers struct {
	Headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		Headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.Headers[strings.ToLower(name)]
}
func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.Headers[name]; ok {
		h.Headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.Headers[name] = value
	}
}

func (h *Headers) ForEach(fe func(k, v string)) {
	for k, v := range h.Headers {
		fe(k, v)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	log.Printf("Field Line Original : %v\n", string(data))

	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		//EMPTY Header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("Malformed header name")
		}
		read += idx + len(rn)
		h.Set(name, value)
	}

	return read, done, nil
}
