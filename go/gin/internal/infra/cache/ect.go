package cache

import (
	"fmt"
	"io"
)

func (o *Cache) ReadChunk(filepath string, chunkIdx, chunkSize int) ([]byte, error) {
	o.MuFileInfoCache.RLock()
	defer o.MuFileInfoCache.RUnlock() // ← 읽기 전체를 보호

	cache, exists := o.FileInfoCache[filepath]
	if !exists {
		return nil, fmt.Errorf("file cache not found: %s", filepath)
	}

	offset := chunkIdx * chunkSize
	end := offset + chunkSize

	if offset >= len(cache.Data) {
		return nil, io.EOF
	}

	if end > len(cache.Data) {
		end = len(cache.Data)
	}

	chunk := make([]byte, end-offset)
	copy(chunk, cache.Data[offset:end])

	return chunk, nil
}
