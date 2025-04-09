package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

const CRLF = "\r\n"
const VALIDCHAR = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-.^_`|~"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	if !strings.Contains(str, CRLF) {
		return 0, false, nil
	}

	if strings.Index(str, CRLF) == 0 {
		return 0, true, nil
	}

	firstHeader := strings.Split(str, CRLF)[0]
	line := strings.TrimSpace(firstHeader)
	pair := strings.SplitN(line, ":", 2)
	keyValid := strings.TrimSpace(pair[0])
	if len(keyValid) != len(pair[0]) {
		return 0, false, fmt.Errorf("invalid spacing in header")
	}
	if len(keyValid) == 0 {
		return 0, false, fmt.Errorf("invalid header key")
	}
	for _, char := range keyValid {
		if !strings.Contains(VALIDCHAR, string(char)) {
			return 0, false, fmt.Errorf("invalid character in header key")
		}
	}

	lowerCaseKey := strings.ToLower(keyValid)
	_, ok := h[lowerCaseKey]
	if ok {
		h[lowerCaseKey] += ", " + strings.TrimSpace(pair[1])
	} else {
		h[lowerCaseKey] = strings.TrimSpace(pair[1])
	}

	return len(firstHeader) + 2, false, nil
}
