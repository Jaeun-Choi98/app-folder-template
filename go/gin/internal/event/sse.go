package event

type SysStt struct {
	ServerTime  string `json:"svrTime"`
	DBState     int8   `json:"dbComm"`
	TCPListener int8   `json:"tcpListener"`
}

func NewSseEventSysStt(sysstt *SysStt) *SseEvent {
	return &SseEvent{
		Event: "sysstt",
		Data:  sysstt,
	}
}
