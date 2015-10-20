package machine

func (c *Cluster) List() []*Machine {
	return c.machines
}

func (c *Cluster) Get(name string) (*Machine, bool) {
	machines := c.List()
	for _, m := range machines {
		if m.Name() == name {
			return m, true
		}
	}
	return nil, false
}
