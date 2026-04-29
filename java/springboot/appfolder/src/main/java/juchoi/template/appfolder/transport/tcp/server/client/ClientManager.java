package juchoi.template.appfolder.transport.tcp.server.client;

import java.util.concurrent.ConcurrentHashMap;

// TcpConfig에서 @Bean으로 등록
public class ClientManager {

    private final ConcurrentHashMap<Integer, TcpClient> clients = new ConcurrentHashMap<>();

    public void register(int id, TcpClient client) { clients.put(id, client); }
    public void unregister(int id)                  { clients.remove(id); }
    public TcpClient get(int id)                    { return clients.get(id); }
    public int count()                              { return clients.size(); }
}
