package ram

import "pjt/internal/infra/object"

// 공유 자원 동기화를 위한 새로운 메모리 할당
func (r *Ram) GetOprtModels() map[int]object.Sample {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return nil
}
