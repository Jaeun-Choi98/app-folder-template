package juchoi.template.appfolder.service;

// Mirrors Go's service.TCPServiceInterface
public interface TcpService {
    void handle0x01(int clientId);
    void handle0xAA(int clientId);
}
