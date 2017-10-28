package riddick

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

type entry struct {
	filename string
	code     string
	typeCode string
	value    interface{}
	data     []byte
}

func (e *entry) ustring() (string, error) {
	if e.value != nil {
		if s, ok := e.value.(string); ok {
			return s, nil
		}
	}
	if e.typeCode != "ustr" {
		return "", fmt.Errorf("reading %s from %s", "ustr", e.typeCode)
	}
	s := string(e.data)
	e.value = s
	return s, nil
}

func (e *entry) bool() (bool, error) {
	if e.value != nil {
		if v, ok := e.value.(bool); ok {
			return v, nil
		}
	}
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
	e.value = ok
	return ok, nil
}

func (e *entry) timestamp() (time.Time, error) {
	if e.value != nil {
		if v, ok := e.value.(time.Time); ok {
			return v, nil
		}
	}
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
