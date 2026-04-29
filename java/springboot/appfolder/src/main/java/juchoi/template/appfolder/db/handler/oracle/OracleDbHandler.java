package juchoi.template.appfolder.db.handler.oracle;

import juchoi.template.appfolder.db.entity.SampleEntity;
import juchoi.template.appfolder.db.handler.DbHandler;
import lombok.RequiredArgsConstructor;

import lombok.extern.slf4j.Slf4j;
import org.springframework.jdbc.core.JdbcTemplate;

import java.util.List;

@Slf4j
@RequiredArgsConstructor
public class OracleDbHandler implements DbHandler {

    private final JdbcTemplate jdbc;

    @Override
    public void ping() {
        jdbc.queryForObject("SELECT 1 FROM DUAL", Integer.class);
    }

    @Override
    public void close() {
        log.info("[Oracle] handler closed");
    }

    @Override
    public List<SampleEntity> findAllSamples() {
        return jdbc.query(
                "SELECT id, name, value FROM sample",
                (rs, row) -> new SampleEntity(rs.getLong("id"), rs.getString("name"), rs.getString("value")));
    }

    @Override
    public void insertSample(SampleEntity entity) {
        jdbc.update("INSERT INTO sample (name, value) VALUES (?, ?)", entity.name(), entity.value());
    }
}
