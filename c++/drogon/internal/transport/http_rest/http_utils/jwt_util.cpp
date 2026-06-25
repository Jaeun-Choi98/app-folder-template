#include "internal/transport/http_rest/http_utils/jwt_util.h"

namespace app {

JwtUtil::JwtUtil(const std::string& secret, int expireMin)
    : secret_(secret), expireMin_(expireMin) {}

std::string JwtUtil::generate(const std::string& subject) const {
    // TODO: HS256 JWT 생성 (jwt-cpp 등 라이브러리 사용)
    return "";
}

std::optional<std::string> JwtUtil::verify(const std::string& token) const {
    // TODO: JWT 검증 후 subject 반환
    return std::nullopt;
}

} // namespace app
