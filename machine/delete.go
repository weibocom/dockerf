package machine

func (c *Cluster) Remove(m *Machine) error {
	return m.Remove()
}
