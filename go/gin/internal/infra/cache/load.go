package cache

import (
	"encoding/hex"
	"fmt"
	"os"
	"pjt/internal/infra/object"
	"pjt/internal/logger"
	"strconv"
	"strings"

	"github.com/Jaeun-Choi98/modules/utils"
)

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

func (o *Cache) LoadFile(fpath, vpath string) error {
	o.MuFileInfoCache.Lock()
	defer o.MuFileInfoCache.Unlock()

	stat, err := os.Stat(fpath)
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v", fpath, err)
		return err
	}

	maxSize := int64(5 * 1024 * 1024 * 1024)
	if stat.Size() > maxSize {
		logger.Infof("[Cache] file too large: %d > %d", stat.Size(), maxSize)
		return fmt.Errorf("file too large: %d > %d", stat.Size(), maxSize)
	}

	data, err := os.ReadFile(fpath)
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v", fpath, err)
		return err
	}

	verBytes, err := os.ReadFile(vpath)
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
		return err
	}

	version := strings.Split(string(verBytes), ".")
	major, err := strconv.Atoi(version[0])
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
		return fmt.Errorf("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
	}
	minor, err := strconv.Atoi(version[1])
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
		return fmt.Errorf("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
	}
	tmp, err := hex.DecodeString(version[2])
	var expectedCheckSum [32]byte
	for i := 0; i < 32; i++ {
		expectedCheckSum[i] = tmp[i]
	}

	// 파일 무결성 검증
	vf, err := os.Open(fpath)
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v")
		return fmt.Errorf("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
	}

	calcCheckSum, err := utils.CalculateChecksumFromFile(vf)
	if err != nil {
		logger.Infof("[Cache] Failed to LoadFile: %v, error: %v")
		return fmt.Errorf("[Cache] Failed to LoadFile: %v, error: %v", vpath, err)
	}

	if expectedCheckSum != calcCheckSum {
		logger.Infof("[Cache] Fialed to LoadFile: %v, diff checksum value: %v(expected) - %v(calc)",
			vpath, expectedCheckSum, calcCheckSum)
		return fmt.Errorf("[Cache] Fialed to LoadFile: %v, diff checksum value: %v(expected) - %v(calc)",
			vpath, expectedCheckSum, calcCheckSum)
	}

	fileInfo := &object.FileInfo{
		Data:         data,
		FileSize:     int64(len(data)),
		Filepath:     fpath,
		MajorVersion: major,
		MinorVersion: minor,
		CheckSum:     expectedCheckSum,
	}

	o.FileInfoCache[fpath] = fileInfo
	return nil
}
