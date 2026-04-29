package juchoi.template.appfolder.worker;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.CacheDataEvent;
import juchoi.template.appfolder.infra.cache.Cache;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;

// WorkerConfig에서 @Bean으로 등록 — @Scheduled은 클래스에 남겨두면 @EnableScheduling이 처리
@Slf4j
@RequiredArgsConstructor
public class CacheWorker {

    private final DbHandler dbHandler;
    private final Cache cache;
    private final EventBus eventBus;

    @Scheduled(fixedDelay = 1000)
    public void sync() {
        try {
            dbHandler.ping();
            eventBus.publish("cache.data", new CacheDataEvent(cache.snapshot()));
        } catch (Exception e) {
            log.warn("[Cache Worker] sync failed: {}", e.getMessage());
        }
    }

    public void shutdown() {
        log.info("[Cache Worker] shutting down");
    }
}
