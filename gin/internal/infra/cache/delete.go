package cache

func (c *Cache) DelLoginSession(id int, acc, ref bool) {
	c.muLoginSessions.Lock()
	defer c.muLoginSessions.Unlock()

	if _, exists := c.LoginSessions[id]; exists {
		if acc {
			c.LoginSessions[id].Acc = nil
			delete(badCodeForCheckTokenMiddleware, id)
		}
		if ref {
			c.LoginSessions[id].Ref = nil
		}
	}
}
