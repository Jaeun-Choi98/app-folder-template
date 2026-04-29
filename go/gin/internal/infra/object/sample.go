package object

type Sample struct {
	Id        int64
	UniqueKey int
}

// 파일
type FileInfo struct {
	Filepath     string
	MajorVersion int
	MinorVersion int
	FileSize     int64
	CheckSum     [32]byte
	Data         []byte
}

// 로그인 세션
type LoginSeesion struct {
	Acc *string // accToken
	Ref *string // refToken
}
