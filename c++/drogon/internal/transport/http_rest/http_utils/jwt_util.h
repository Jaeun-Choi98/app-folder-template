#pragma once

#include <string>
#include <optional>

namespace app {

class JwtUtil {
public:
    explicit JwtUtil(const std::string& secret, int expireMin = 60);

    std::string generate(const std::string& subject) const;
    std::optional<std::string> verify(const std::string& token) const;

private:
    std::string secret_;
    int         expireMin_;
};

} // namespace app
