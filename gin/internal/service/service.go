package service

type APIServcieInterface interface {
	Test() string
	Close() error
}

type TCPServiceInterface interface {
	Handle0x010(clientId uint32) error
	Handle0xAA(clientId uint32)
	Close()
}
