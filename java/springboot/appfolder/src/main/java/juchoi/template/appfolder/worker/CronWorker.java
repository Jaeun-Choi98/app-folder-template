package juchoi.template.appfolder.worker;

import juchoi.template.appfolder.db.handler.DbHandler;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;

// WorkerConfig에서 @Bean으로 등록
@Slf4j
@RequiredArgsConstructor
public class CronWorker {

    private final DbHandler dbHandler;

    @Scheduled(cron = "0 0 0 * * *")
    public void dailyTask() {
        log.info("[Cron Worker] running daily task");
    }

    @Scheduled(fixedDelay = 30_000)
    public void heartbeat() {
        try {
            dbHandler.ping();
        } catch (Exception e) {
            log.warn("[Cron Worker] DB heartbeat failed: {}", e.getMessage());
        }
    }
}
