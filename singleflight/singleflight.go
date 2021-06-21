package singleflight

import "sync"

type call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {

	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	c, ok := g.m[key]
	if ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.value, c.err
	}

	c = new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.value, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.value, c.err
}
