package machine

func (c *Cluster) ListAll() []*Machine {
	return c.List(func(m *Machine) bool { return true })
}

func (c *Cluster) List(accept func(m *Machine) bool) []*Machine {
	c.Lock()
	defer c.Unlock()
	ms := make([]*Machine, len(c.machines))
	cnt := 0
	for _, m := range c.machines {
		if accept(m) {
			ms[cnt] = m
			cnt++
		}
	}
	return ms[0:cnt]
}

func (c *Cluster) Get(name string) (*Machine, bool) {
	c.Lock()
	defer c.Unlock()
	m, found := c.machines[name]
	return m, found
}
