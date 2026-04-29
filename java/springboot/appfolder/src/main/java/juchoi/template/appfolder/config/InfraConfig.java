package juchoi.template.appfolder.config;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.infra.cache.Cache;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// 인프라 빈 등록 — EventBus, Cache
@Configuration
public class InfraConfig {

    @Bean
    public EventBus eventBus() {
        return new EventBus();
    }

    // Cache 생성 시 DB에서 초기 데이터 로드
    @Bean
    public Cache cache(DbHandler dbHandler) {
        return new Cache(dbHandler);
    }
}
