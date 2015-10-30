package machine

func (c *Cluster) Start(m *Machine) error {
	return m.Start()
}
