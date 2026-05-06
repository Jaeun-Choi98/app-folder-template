package juchoi.template.appfolder.service.tcp;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.CacheDataEvent;
import juchoi.template.appfolder.infra.cache.Cache;
import juchoi.template.appfolder.service.TcpService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

// ServiceConfig에서 @Bean으로 등록
@Slf4j
@RequiredArgsConstructor
public class TcpServiceImpl implements TcpService {

    private final DbHandler dbHandler;
    private final Cache cache;
    private final EventBus eventBus;

    @Override
    public void handle0x21(int clientId) {
        log.info("[TCP Service] 0x21 from client {}", clientId);
    }

    @Override
    public void handle0x22(int clientId) {
        log.info("[TCP Service] 0x22 from client {}", clientId);
        eventBus.publish("system.event", new CacheDataEvent("received 0x22 packet"));
    }
}
