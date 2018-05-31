package pool

type pool struct {
	tokens chan struct{}
}

// Creates a new  pool of set size
func NewPool(size int) *pool {
	p := &pool{
		tokens: make(chan struct{}, size),
	}
	for _, t := range make([]struct{}, size) {
		p.tokens <- t
	}
	return p
}

// Claims a token from pool or blocks until one is available
func (wp *pool) Claim() {
	<-wp.tokens
}

// Frees a token from pool
func (wp *pool) Release() {
	wp.tokens <- struct{}{}
}
