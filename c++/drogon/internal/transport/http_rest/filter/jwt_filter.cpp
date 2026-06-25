#include "internal/transport/http_rest/filter/jwt_filter.h"
#include "internal/transport/http_rest/http_utils/jwt_util.h"
#include "internal/transport/http_rest/http_utils/http_error.h"

namespace app {

void JwtFilter::setJwtUtil(std::shared_ptr<JwtUtil> jwt) {
    jwt_ = std::move(jwt);
}

void JwtFilter::doFilter(const drogon::HttpRequestPtr& req,
                         drogon::FilterCallback&&       fcb,
                         drogon::FilterChainCallback&&  fccb) {
    auto cookie = req->getCookie("token");
    if (cookie.empty()) {
        fcb(makeHttpError(drogon::k401Unauthorized, "missing token"));
        return;
    }

    auto subject = jwt_ ? jwt_->verify(cookie) : std::nullopt;
    if (!subject) {
        fcb(makeHttpError(drogon::k401Unauthorized, "invalid token"));
        return;
    }

    // 인증된 subject를 요청 속성에 저장
    req->addHeader("X-Auth-Subject", *subject);
    fccb();
}

} // namespace app
