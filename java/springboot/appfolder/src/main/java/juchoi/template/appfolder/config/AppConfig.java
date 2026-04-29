package juchoi.template.appfolder.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableScheduling;

import juchoi.template.appfolder.app.AppLifecycle;
import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.healthcheck.SystemMonitoring;
import juchoi.template.appfolder.transport.tcp.server.TcpServer;
import juchoi.template.appfolder.worker.CacheWorker;

// 스케줄러 활성화 — CacheWorker, CronWorker의 @Scheduled 메서드를 처리
@Configuration
@EnableScheduling
public class AppConfig {
    @Bean
    public AppLifecycle appLifecycle(TcpServer tcpServer, CacheWorker cacheWorker) {
        return new AppLifecycle(tcpServer, cacheWorker);
    }

    @Bean
    public SystemMonitoring systemMonitoring(DbHandler dbHandler, EventBus eventBus) {
        return new SystemMonitoring(dbHandler, eventBus);
    }
}
