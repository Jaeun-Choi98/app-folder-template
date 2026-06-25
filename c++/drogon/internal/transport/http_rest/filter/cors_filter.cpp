#include "internal/transport/http_rest/filter/cors_filter.h"

namespace app {

void CorsFilter::doFilter(const drogon::HttpRequestPtr& req,
                          drogon::FilterCallback&&       fcb,
                          drogon::FilterChainCallback&&  fccb) {
    if (req->method() == drogon::Options) {
        auto resp = drogon::HttpResponse::newHttpResponse();
        resp->addHeader("Access-Control-Allow-Origin", "*");
        resp->addHeader("Access-Control-Allow-Methods",
                        "GET, POST, PUT, DELETE, OPTIONS");
        resp->addHeader("Access-Control-Allow-Headers",
                        "Content-Type, Authorization");
        resp->addHeader("Access-Control-Allow-Credentials", "true");
        fcb(resp);
        return;
    }
    fccb();
}

} // namespace app
