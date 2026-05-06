package juchoi.template.appfolder.config;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.infra.cache.Cache;
import juchoi.template.appfolder.service.ApiService;
import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.service.api.ApiServiceImpl;
import juchoi.template.appfolder.service.tcp.TcpServiceImpl;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// 비즈니스 로직 빈 등록 — 인터페이스 타입으로 노출
@Configuration
public class ServiceConfig {

    @Bean
    public ApiService apiService(DbHandler dbHandler, Cache cache) {
        return new ApiServiceImpl(dbHandler, cache);
    }

    @Bean
    public TcpService tcpService(DbHandler dbHandler, Cache cache, EventBus eventBus) {
        return new TcpServiceImpl(dbHandler, cache, eventBus);
    }
}
