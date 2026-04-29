package juchoi.template.appfolder.eventbus;

import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.Executor;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

import lombok.extern.slf4j.Slf4j;

// InfraConfig에서 @Bean으로 등록
@Slf4j
public class EventBus {

    private final Map<String, List<Consumer<Object>>> subscribers = new ConcurrentHashMap<>();
    private final Executor executor = Executors.newCachedThreadPool();

    @SuppressWarnings("unchecked")
    public <T> void subscribe(String topic, Consumer<T> handler) {
        subscribers.computeIfAbsent(topic, k -> new CopyOnWriteArrayList<>())
                .add(event -> handler.accept((T) event));
    }

    public <T> void publish(String topic, T event) {
        List<Consumer<Object>> handlers = subscribers.getOrDefault(topic, List.of());
        for (Consumer<Object> handler : handlers) {
            CompletableFuture
                    .runAsync(() -> handler.accept(event), executor)
                    .exceptionally(e -> {
                        log.error("[EventBus] handler error on topic '{}'", topic, e);
                        return null;
                    });
        }
    }

    public void close() {
        ((ExecutorService) executor).shutdown();
        try {
            ((ExecutorService) executor).awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException ignored) {
        }
    }
}
