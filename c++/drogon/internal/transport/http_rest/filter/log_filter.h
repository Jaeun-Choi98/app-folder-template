#pragma once

#include <drogon/HttpFilter.h>

namespace app {

class LogFilter : public drogon::HttpFilter<LogFilter> {
public:
    void doFilter(const drogon::HttpRequestPtr& req,
                  drogon::FilterCallback&&       fcb,
                  drogon::FilterChainCallback&&  fccb) override;
};

} // namespace app
