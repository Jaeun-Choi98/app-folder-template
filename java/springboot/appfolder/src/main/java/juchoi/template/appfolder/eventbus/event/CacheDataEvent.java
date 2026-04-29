package juchoi.template.appfolder.eventbus.event;

import lombok.RequiredArgsConstructor;

// Published on "cache.data", "system.state", "system.event" topics.
@RequiredArgsConstructor
public class CacheDataEvent {
    private final String payload;

    public String getPayload() { return payload; }
}
