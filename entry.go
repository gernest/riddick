package riddick

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mattetti/cocoa"

	"github.com/DHowett/go-plist"
)

type entry struct {
	filename string
	code     string
	typeCode string
	data     []byte
}

func (e *entry) ustring() (string, error) {
	if e.typeCode != "ustr" {
		return "", fmt.Errorf("reading %s from %s", "ustr", e.typeCode)
	}
	s := string(e.data)
	return s, nil
}

func (e *entry) bool() (bool, error) {
	if e.typeCode != "bool" {
		return false, fmt.Errorf("reading %s from %s", "bool", e.typeCode)
	}
	if len(e.data) != 1 {
		return false, errors.New("bool must be a 1 byte")
	}
	b := e.data[0]
	var ok bool
	if b == 0x1 {
		ok = true
	}
	return ok, nil
}

func (e *entry) timestamp() (time.Time, error) {
	if e.typeCode != "dutc" {
		return time.Time{}, fmt.Errorf("reading %s from %s", "bool", e.typeCode)
	}
	diff := 2082844800 * time.Second
	v := int64(binary.BigEndian.Uint64(e.data)) / 65536
	if v != math.MinInt64 {
		ts := time.Unix(v, 0)
		return ts.Add(-diff), nil
	}
	return time.Time{}, nil
}

func (e *entry) plist() (map[string]interface{}, int, error) {
	switch e.code {
	case "bwsp", "lsvp", "lsvP", "icvp":
		o := make(map[string]interface{})
		f, err := plist.Unmarshal(e.data, &o)
		if err != nil {
			return nil, 0, err
		}
		return o, f, nil
	default:
		return nil, 0, fmt.Errorf("%s is not a plist entry", e.code)
	}
}

func (e *entry) bookmark() (*cocoa.BookmarkData, error) {
	if e.code != "pBBk" {
		return nil, fmt.Errorf("reading %s from %s", "pBBk", e.code)
	}
	return cocoa.AliasFromReader(bytes.NewReader(e.data))
}

func (e *entry) len() int {
	l := 4 + len(e.filename) + 8
	switch e.typeCode {
	case "bool":
		l++
	case "long", "shor", "type":
		l += 4
	case "blob":
		l += 4 + len(e.data)
	case "ustr":
		s := utf16be2utf8(e.data)
		l += 4 + len(s)
	case "comp", "dutc":
		l += 8
	}
	return l
}
