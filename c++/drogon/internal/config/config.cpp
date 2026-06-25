#include "internal/config/config.h"

#include <fstream>
#include <sstream>
#include <algorithm>
#include <stdexcept>

namespace app {

std::unordered_map<std::string, std::string>
Config::parseIni(const std::string& path) {
    std::ifstream file(path);
    if (!file.is_open())
        throw std::runtime_error("cannot open config file: " + path);

    std::unordered_map<std::string, std::string> kv;
    std::string line, section;

    while (std::getline(file, line)) {
        // trim
        line.erase(0, line.find_first_not_of(" \t\r\n"));
        line.erase(line.find_last_not_of(" \t\r\n") + 1);

        if (line.empty() || line[0] == '#' || line[0] == ';')
            continue;

        if (line.front() == '[' && line.back() == ']') {
            section = line.substr(1, line.size() - 2);
            continue;
        }

        auto eq = line.find('=');
        if (eq == std::string::npos) continue;

        std::string key = line.substr(0, eq);
        std::string val = line.substr(eq + 1);
        key.erase(key.find_last_not_of(" \t") + 1);
        val.erase(0, val.find_first_not_of(" \t"));

        std::string fullKey = section.empty() ? key : section + "." + key;
        kv[fullKey] = val;
    }
    return kv;
}

Config Config::load(const std::string& path) {
    auto kv = parseIni(path);
    Config c;

    auto get = [&](const std::string& k) -> std::string {
        auto it = kv.find(k);
        return it != kv.end() ? it->second : "";
    };
    auto getInt = [&](const std::string& k, int def) -> int {
        auto s = get(k);
        return s.empty() ? def : std::stoi(s);
    };

    c.serverHost   = get("SERVER.HOST");
    c.serverPort   = getInt("SERVER.PORT", 8080);
    c.tcpHost      = get("TCP.HOST");
    c.tcpPort      = getInt("TCP.PORT", 9090);
    c.dbType       = get("DB.TYPE");
    c.dbHost       = get("DB.HOST");
    c.dbPort       = getInt("DB.PORT", 3306);
    c.dbUser       = get("DB.USER");
    c.dbPassword   = get("DB.PASSWORD");
    c.dbName       = get("DB.NAME");
    c.redisHost    = get("REDIS.HOST");
    c.redisPort    = getInt("REDIS.PORT", 6379);
    c.redisPassword = get("REDIS.PASSWORD");
    c.jwtSecret    = get("JWT.SECRET");
    c.jwtExpireMin = getInt("JWT.EXPIRE_MIN", 60);

    return c;
}

} // namespace app
