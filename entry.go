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
	"golang.org/x/text/encoding/unicode"
)

var utf16be = unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
var utf16beDec = utf16be.NewDecoder()
var utf16beEnc = utf16be.NewEncoder()

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

func (e *entry) len() (int, error) {
	l := 4 + len(e.filename) + 8
	switch e.typeCode {
	case "bool":
		l++
	case "long", "shor", "type":
		l += 4
	case "blob":
		l += 4 + len(e.data)
	case "ustr":
		s, err := utf16beDec.Bytes(e.data)
		if err != nil {
			return 0, err
		}
		l += 4 + len(s)
	case "comp", "dutc":
		l += 8
	}
	return l, nil
}

func (e *entry) decodeIloc() (uint32, uint32, error) {
	if e.code != "Iloc" {
		return 0, 0, fmt.Errorf("reading %s from %s", "Iloc", e.code)
	}
	r := bytes.NewReader(e.data)
	var x, y uint32
	err := binary.Read(r, binary.BigEndian, &x)
	if err != nil {
		return 0, 0, err
	}
	err = binary.Read(r, binary.BigEndian, &y)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

func (e *entry) Encode() ([]byte, error) {
	var o bytes.Buffer
	name, err := utf16beEnc.String(e.filename)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&o, binary.BigEndian, uint32(len(name)))
	if err != nil {
		return nil, err
	}
	_, err = o.WriteString(name)
	if err != nil {
		return nil, err
	}
	o.WriteString(e.code)
	o.WriteString(e.typeCode)
	switch e.typeCode {
	case "bool", "long", "shor", "comp", "dutc":
		o.Write(e.data)
	case "blob", "ustr":
		err = binary.Write(&o, binary.BigEndian, uint32(len(e.data)))
		if err != nil {
			return nil, err
		}
		o.Write(e.data)
	default:
		return nil, fmt.Errorf("unkown type code %s", e.typeCode)
	}
	return o.Bytes(), nil
}
