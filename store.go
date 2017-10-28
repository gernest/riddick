package riddick

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
