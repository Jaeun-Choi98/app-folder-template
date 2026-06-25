#include "internal/db/db_handler/maria/maria_handler.h"
#include "internal/config/config.h"

namespace app {

MariaHandler::MariaHandler(const Config& cfg) {
    connStr_ = "mysql://host=" + cfg.dbHost +
               " port=" + std::to_string(cfg.dbPort) +
               " dbname=" + cfg.dbName +
               " user=" + cfg.dbUser +
               " password=" + cfg.dbPassword;
}

MariaHandler::~MariaHandler() { disconnect(); }

void MariaHandler::connect() {
    sql_ = std::make_unique<soci::session>(connStr_);
}

void MariaHandler::disconnect() {
    sql_.reset();
}

bool MariaHandler::isConnected() const {
    return sql_ != nullptr;
}

std::vector<SampleEntity> MariaHandler::findAll() {
    std::vector<SampleEntity> result;
    // TODO: SELECT * FROM sample
    return result;
}

std::optional<SampleEntity> MariaHandler::findById(int id) {
    // TODO: SELECT * FROM sample WHERE id = :id
    return std::nullopt;
}

int MariaHandler::insert(const SampleEntity& entity) {
    // TODO: INSERT INTO sample (name, value) VALUES (:name, :value)
    return 0;
}

bool MariaHandler::update(const SampleEntity& entity) {
    // TODO: UPDATE sample SET name = :name, value = :value WHERE id = :id
    return false;
}

bool MariaHandler::remove(int id) {
    // TODO: DELETE FROM sample WHERE id = :id
    return false;
}

void MariaHandler::insertLog(const LogEntity& log) {
    // TODO: INSERT INTO event_log (level, message) VALUES (:level, :message)
}

} // namespace app
