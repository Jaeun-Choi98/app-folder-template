package juchoi.template.appfolder.transport.http.controller;

import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.CacheDataEvent;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

// WebConfig에서 @Bean으로 등록
@RestController
@RequestMapping("/sse")
public class SseController {

    private final Map<Integer, SseEmitter> emitters = new ConcurrentHashMap<>();
    private final AtomicInteger idGen = new AtomicInteger(0);

    public SseController(EventBus eventBus) {
        eventBus.subscribe("cache.data",   (CacheDataEvent e) -> broadcast("cache.data",   e.getPayload()));
        eventBus.subscribe("system.state", (CacheDataEvent e) -> broadcast("system.state", e.getPayload()));
        eventBus.subscribe("system.event", (CacheDataEvent e) -> broadcast("system.event", e.getPayload()));
    }

    @GetMapping(value = "/connect", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public SseEmitter connect() {
        int id = idGen.incrementAndGet();
        SseEmitter emitter = new SseEmitter(0L);
        emitters.put(id, emitter);
        emitter.onCompletion(() -> emitters.remove(id));
        emitter.onTimeout(()    -> emitters.remove(id));
        emitter.onError(e       -> emitters.remove(id));
        return emitter;
    }

    @PostMapping("/send")
    public void sendAll(@RequestBody String payload) {
        broadcast("message", payload);
    }

    @PostMapping("/send/{id}")
    public void sendToUser(@PathVariable int id, @RequestBody String payload) {
        SseEmitter emitter = emitters.get(id);
        if (emitter == null) return;
        try {
            emitter.send(SseEmitter.event().name("message").data(payload));
        } catch (IOException e) {
            emitters.remove(id);
        }
    }

    private void broadcast(String eventName, String payload) {
        emitters.forEach((id, emitter) -> {
            try {
                emitter.send(SseEmitter.event().name(eventName).data(payload));
            } catch (IOException e) {
                emitters.remove(id);
            }
        });
    }
}
