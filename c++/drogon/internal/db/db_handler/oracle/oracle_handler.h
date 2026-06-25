#pragma once

#include "internal/db/db_handler/db_handler.h"

#include <soci/soci.h>
#include <memory>

namespace app {

class OracleHandler : public DbHandler {
public:
    explicit OracleHandler(const Config& cfg);
    ~OracleHandler() override;

    void connect() override;
    void disconnect() override;
    bool isConnected() const override;

    std::vector<SampleEntity>          findAll() override;
    std::optional<SampleEntity>        findById(int id) override;
    int                                insert(const SampleEntity& entity) override;
    bool                               update(const SampleEntity& entity) override;
    bool                               remove(int id) override;
    void                               insertLog(const LogEntity& log) override;

private:
    std::string connStr_;
    std::unique_ptr<soci::session> sql_;
};

} // namespace app
