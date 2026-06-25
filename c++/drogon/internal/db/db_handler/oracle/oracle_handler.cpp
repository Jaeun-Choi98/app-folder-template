#include "internal/db/db_handler/oracle/oracle_handler.h"
#include "internal/config/config.h"

namespace app {

OracleHandler::OracleHandler(const Config& cfg) {
    connStr_ = "oracle://service=" + cfg.dbName +
               " user=" + cfg.dbUser +
               " password=" + cfg.dbPassword;
}

OracleHandler::~OracleHandler() { disconnect(); }

void OracleHandler::connect() {
    sql_ = std::make_unique<soci::session>(connStr_);
}

void OracleHandler::disconnect() {
    sql_.reset();
}

bool OracleHandler::isConnected() const {
    return sql_ != nullptr;
}

std::vector<SampleEntity> OracleHandler::findAll() {
    std::vector<SampleEntity> result;
    // TODO: SELECT * FROM sample
    return result;
}

std::optional<SampleEntity> OracleHandler::findById(int id) {
    // TODO: SELECT * FROM sample WHERE id = :id
    return std::nullopt;
}

int OracleHandler::insert(const SampleEntity& entity) {
    // TODO: INSERT INTO sample (name, value) VALUES (:name, :value)
    return 0;
}

bool OracleHandler::update(const SampleEntity& entity) {
    // TODO: UPDATE sample SET name = :name, value = :value WHERE id = :id
    return false;
}

bool OracleHandler::remove(int id) {
    // TODO: DELETE FROM sample WHERE id = :id
    return false;
}

void OracleHandler::insertLog(const LogEntity& log) {
    // TODO: INSERT INTO event_log (level, message) VALUES (:level, :message)
}

} // namespace app
