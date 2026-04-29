package juchoi.template.appfolder.db.handler;

import juchoi.template.appfolder.db.entity.SampleEntity;

import java.util.List;

/**
 * Mirrors Go's DBHandlerInterface.
 * Add new query methods here and implement in each driver class (maria/, oracle/).
 */
public interface DbHandler {
    void ping();
    void close();
    List<SampleEntity> findAllSamples();
    void insertSample(SampleEntity entity);
}
