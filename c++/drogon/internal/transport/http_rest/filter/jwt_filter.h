#pragma once

#include <drogon/HttpFilter.h>
#include <memory>

namespace app {

class JwtUtil;

class JwtFilter : public drogon::HttpFilter<JwtFilter> {
public:
    void doFilter(const drogon::HttpRequestPtr& req,
                  drogon::FilterCallback&&       fcb,
                  drogon::FilterChainCallback&&  fccb) override;

    void setJwtUtil(std::shared_ptr<JwtUtil> jwt);

private:
    std::shared_ptr<JwtUtil> jwt_;
};

} // namespace app
