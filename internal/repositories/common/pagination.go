package common

type Pagination struct {
	Limit  int
	Offset int
}

func (p Pagination) Normalize() Pagination {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 50
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}
