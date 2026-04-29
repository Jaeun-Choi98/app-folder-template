package cache

import (
	"pjt/internal/infra/object"

	"github.com/Jaeun-Choi98/modules/utils"
)

func (r *Cache) GetOprtModels(ids []int64, all bool) (map[int64]*object.Sample, map[int]struct{}) {

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
				return nil, nil
			}
			n := &object.Sample{}
			utils.DeepCopy(r.SampleCache[id], n)
			new[id] = n
			checkUnique[n.UniqueKey] = struct{}{}
		}
	}

	return new, checkUnique
}

func CheckAccToken(id int, accToken string) bool {
	token, exists := badCodeForCheckTokenMiddleware[id]
	if !exists {
		return false
	}
	if *token != accToken {
		return false
	}
	return true
}

func (c *Cache) GetLoginSession(id int) *object.LoginSeesion {
	c.muLoginSessions.RLock()
	defer c.muLoginSessions.RUnlock()
	if _, exists := c.LoginSessions[id]; exists {
		n := &object.LoginSeesion{}
		utils.DeepCopy(c.LoginSessions[id], n)
		return n
	}
	return nil
}
