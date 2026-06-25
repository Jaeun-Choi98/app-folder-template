#include "internal/transport/http_rest/filter/log_filter.h"
#include "internal/logger/logger.h"

namespace app {

void LogFilter::doFilter(const drogon::HttpRequestPtr& req,
                         drogon::FilterCallback&&       fcb,
                         drogon::FilterChainCallback&&  fccb) {
    Logger::instance().info(
        req->getMethodString() + std::string(" ") + req->getPath());
    fccb();
}

} // namespace app
