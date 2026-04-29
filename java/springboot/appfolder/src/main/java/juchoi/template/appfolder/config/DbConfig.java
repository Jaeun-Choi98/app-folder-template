package juchoi.template.appfolder.config;

import javax.sql.DataSource;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.lang.NonNull;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.db.handler.maria.MariaDbHandler;
import juchoi.template.appfolder.db.handler.oracle.OracleDbHandler;

// Mirrors Go's NewDBHandler() switch statement — selects driver from db.type property.
@Configuration
public class DbConfig {

    @Value("${db.type}")
    private String dbType;

    @Bean
    public DbHandler dbHandler(@NonNull DataSource dataSource) {
        return switch (dbType) {
            case "maria", "mysql" -> new MariaDbHandler(new JdbcTemplate(dataSource));
            case "oracle" -> new OracleDbHandler(new JdbcTemplate(dataSource));
            default -> throw new IllegalArgumentException("Unsupported db.type: " + dbType);
        };
    }
}
