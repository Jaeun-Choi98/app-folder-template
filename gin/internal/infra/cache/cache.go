package cache

import (
	"errors"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"sync"
)

/**
* cache는 객체의 상태를 추적하고 관리함.
* -> 객체의 일관성과 무결성을 유지해야 함. ( 객체 간에 관계가 있을 경우에는 연관된 객체의 일관성도 유지해야 함 )
*
* cache는 인메모리DB로써의 역할을 함.
* -> 논리적인 view를 제공해야 함. ( DBMS에서 처리하는게 아닌 서버에서 처리한다. )
*
* 추가적으로 cache는 http, tcp, ... 등 각 통신 세션 상태도 관리하는 해야 함.
 */
var (
	errNotExistId                  = errors.New("not exists id ( some or one )")
	badCodeForCheckTokenMiddleware = make(map[int]*string)
)

type Cache struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int64]*object.Sample

	/**
	 * - 파일 전송 프로토콜 1안
	 * FileInfoCache는 TCP/IP 바이너리로 파일 다운로드를 할 때, 사용됨 ( TCP Service Layer에서 사용 )
	 * 관련 메서드: Cache.LoadFile(), Cache.ReadChunk()
	 *
	 * FileInfoCache는 Get메서드가 따로 없음.
	 * Get요청마다 객체를 새로 할당하는 것이 비효율적이라 생각.
	 *
	 * - 파일 전송 프로토콜 2안
	 * ClientManager.SendFileToClient() 사용
	 */
	FileInfoCache   map[string]*object.FileInfo // file path -> file info
	MuFileInfoCache sync.RWMutex

	/**
	 * http 로그인 사용자 세션
	 */
	LoginSessions   map[int]*object.LoginSeesion // memberId -> login session
	muLoginSessions sync.RWMutex
}

func NewCacheMem(dao dbhandler.DBHandlerInterface) (*Cache, error) {

	ram := &Cache{
		dao:           dao,
		LoginSessions: make(map[int]*object.LoginSeesion),
	}

	if err := ram.LoadSampleCache(); err != nil {
		return nil, err
	}
	return ram, nil
}
