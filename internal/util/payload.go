package util

import (
	"strconv"
	"strings"
)

type Payload string

func PayloadFromText(text string) Payload {
	payload := strings.TrimSpace(text)

	return Payload(payload)
}

func (p Payload) AsIndex() (int, bool) {
	i, err := strconv.Atoi(string(p))
	if err != nil {
		return 0, false
	}

	return i, true
}

func (p Payload) String() string {
	return string(p)
}

func (p Payload) IsEmpty() bool {
	return len(p) == 0
}
