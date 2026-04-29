package juchoi.template.appfolder.app;

import juchoi.template.appfolder.transport.tcp.server.TcpServer;
import juchoi.template.appfolder.worker.CacheWorker;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.SmartLifecycle;

/**
 * 시작/종료 순서 제어 — LifecycleConfig에서 @Bean으로 등록.
 * @Component 없음: 직접 new로 생성하지 않고 Spring 컨텍스트가 관리.
 */
@Slf4j
@RequiredArgsConstructor
public class AppLifecycle implements SmartLifecycle {

    private volatile boolean running = false;

    private final TcpServer tcpServer;
    private final CacheWorker cacheWorker;

    @Override
    public void start() {
        log.info("[App] Starting...");
        tcpServer.start();
        running = true;
        log.info("[App] Started");
    }

    @Override
    public void stop() {
        log.info("[App] Shutting down...");
        tcpServer.shutdown();
        cacheWorker.shutdown();
        running = false;
        log.info("[App] Terminated");
    }

    @Override
    public boolean isRunning() {
        return running;
    }

    @Override
    public int getPhase() {
        return Integer.MAX_VALUE;
    }
}
