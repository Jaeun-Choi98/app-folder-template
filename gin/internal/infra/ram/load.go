package ram

/**
 * using mapper
 */

func (r *Ram) LoadSampleCache() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}
