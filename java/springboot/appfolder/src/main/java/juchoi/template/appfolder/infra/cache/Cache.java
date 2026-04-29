package juchoi.template.appfolder.infra.cache;

import com.fasterxml.jackson.databind.ObjectMapper;
import juchoi.template.appfolder.db.entity.SampleEntity;
import juchoi.template.appfolder.db.handler.DbHandler;

import java.util.concurrent.ConcurrentHashMap;

// InfraConfig에서 @Bean으로 등록
public class Cache {

    private final ConcurrentHashMap<Long, SampleEntity> store = new ConcurrentHashMap<>();
    private final ObjectMapper mapper = new ObjectMapper();

    public Cache(DbHandler dbHandler) {
        dbHandler.findAllSamples().forEach(e -> store.put(e.id(), e));
    }

    public void put(SampleEntity entity)  { store.put(entity.id(), entity); }
    public SampleEntity get(Long id)      { return store.get(id); }
    public void remove(Long id)           { store.remove(id); }

    public String snapshot() {
        try {
            return mapper.writeValueAsString(store.values());
        } catch (Exception e) {
            return "[]";
        }
    }
}
