package cache

import "pjt/internal/infra/object"

func (c *Cache) SetLoginSession(id int, acc, ref *string) {
	c.muLoginSessions.Lock()
	defer c.muLoginSessions.Unlock()

	if _, exists := c.LoginSessions[id]; exists {
		if acc != nil {
			c.LoginSessions[id].Acc = acc
			badCodeForCheckTokenMiddleware[id] = acc
		}
		if ref != nil {
			c.LoginSessions[id].Ref = ref
		}
	} else {
		c.LoginSessions[id] = &object.LoginSeesion{Acc: acc, Ref: ref}
	}
}
