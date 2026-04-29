package juchoi.template.appfolder.config;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.infra.cache.Cache;
import juchoi.template.appfolder.worker.CacheWorker;
import juchoi.template.appfolder.worker.CronWorker;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

// 워커 빈 등록 — @Scheduled은 클래스에 남겨두고 @EnableScheduling이 처리
@Configuration
public class WorkerConfig {

    @Bean
    public CacheWorker cacheWorker(DbHandler dbHandler, Cache cache, EventBus eventBus) {
        return new CacheWorker(dbHandler, cache, eventBus);
    }

    @Bean
    public CronWorker cronWorker(DbHandler dbHandler) {
        return new CronWorker(dbHandler);
    }
}
