package riddick

import (
	"path/filepath"
)

type store struct {
	a                                                *allocator
	root, levels, records, nodes, pageSize, minUsage uint32
}

func newStore(a *allocator) (*store, error) {
	i := a.toc["DSDB"]
	r, err := a.GetBlock(i)
	if err != nil {
		return nil, err
	}
	s := &store{a: a}
	rn, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	s.root = rn

	l, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	s.levels = l
	rc, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	s.records = rc

	nodes, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	s.nodes = nodes

	pageSize, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	s.pageSize = pageSize
	s.minUsage = 2 * s.pageSize
	return s, nil
}

func (s *store) find(pattern string) ([]*entry, error) {
	var o []*entry
	terr := s.a.traverse(s.root, func(e *entry) error {
		ok, err := filepath.Match(pattern, e.filename)
		if err != nil {
			return err
		}
		if ok {
			ce := &entry{
				filename: e.filename,
				code:     e.code,
				typeCode: e.typeCode,
				data:     make([]byte, len(e.data)),
			}
			copy(ce.data, e.data)
			o = append(o, ce)
		}
		return nil
	})
	if terr != nil {
		return nil, terr
	}
	return o, nil
}
