package juchoi.template.appfolder.healthcheck;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.CacheDataEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;

// LifecycleConfig에서 @Bean으로 등록
@Slf4j
@RequiredArgsConstructor
public class SystemMonitoring {

    private final DbHandler dbHandler;
    private final EventBus eventBus;

    @Scheduled(fixedDelay = 5000)
    public void check() {
        try {
            dbHandler.ping();
            eventBus.publish("system.state", new CacheDataEvent("ok"));
        } catch (Exception e) {
            log.warn("[Monitoring] DB unreachable: {}", e.getMessage());
            eventBus.publish("system.state", new CacheDataEvent("db_down"));
        }
    }
}
