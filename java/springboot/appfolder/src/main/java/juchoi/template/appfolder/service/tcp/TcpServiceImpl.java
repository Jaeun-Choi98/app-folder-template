package juchoi.template.appfolder.service.tcp;

import juchoi.template.appfolder.db.handler.DbHandler;
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

    @Override
    public void handle0x01(int clientId) {
        log.info("[TCP Service] 0x01 from client {}", clientId);
    }

    @Override
    public void handle0xAA(int clientId) {
        log.info("[TCP Service] 0xAA from client {}", clientId);
    }
}
