package juchoi.template.appfolder.config;

import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.transport.tcp.server.TcpServer;
import juchoi.template.appfolder.transport.tcp.server.client.ClientManager;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// TCP 전송 계층 빈 등록
@Configuration
public class TcpConfig {

    @Bean
    public ClientManager clientManager() {
        return new ClientManager();
    }

    @Bean
    public TcpServer tcpServer(
            @Value("${tcp.port:9000}") int port,
            ClientManager clientManager,
            TcpService tcpService,
            EventBus eventBus) {
        return new TcpServer(port, clientManager, tcpService, eventBus);
    }
}
