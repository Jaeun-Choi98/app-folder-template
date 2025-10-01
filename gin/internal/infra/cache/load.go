package cache

/**
 * using mapper ( db layer )
 *
 *
 */

func (r *Cache) LoadSampleCache() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}
