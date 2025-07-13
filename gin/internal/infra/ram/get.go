package ram

import (
	"pjt/internal/infra/object"
	"pjt/internal/utils"
)

func (r *Ram) GetOprtModels(ids []int64, all bool) (map[int64]*object.Sample, map[int]struct{}, error) {
	if len(r.SampleCache) == 0 {
		return nil, nil, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	new := make(map[int64]*object.Sample)
	checkUnique := make(map[int]struct{})
	if all {
		for id, obj := range r.SampleCache {
			n := &object.Sample{}
			utils.DeepCopy(obj, n)
			new[id] = n
			checkUnique[n.UniqueKey] = struct{}{}
		}
	} else {
		for _, id := range ids {
			if _, exists := r.SampleCache[id]; !exists {
				return nil, nil, errNotExistId
			}
			n := &object.Sample{}
			utils.DeepCopy(r.SampleCache[id], n)
			new[id] = n
			checkUnique[n.UniqueKey] = struct{}{}
		}
	}

	return new, checkUnique, nil
}
