package ram

func (r *Ram) LoadSampleCache() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}
