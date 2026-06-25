#include "internal/db/db_handler/db_handler.h"
#include "internal/db/db_handler/maria/maria_handler.h"
#include "internal/db/db_handler/oracle/oracle_handler.h"
#include "internal/config/config.h"

#include <stdexcept>

namespace app {

std::unique_ptr<DbHandler> DbHandler::create(const Config& cfg) {
    if (cfg.dbType == "maria" || cfg.dbType == "mysql")
        return std::make_unique<MariaHandler>(cfg);
    if (cfg.dbType == "oracle")
        return std::make_unique<OracleHandler>(cfg);

    throw std::runtime_error("unsupported db type: " + cfg.dbType);
}

} // namespace app
