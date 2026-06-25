#pragma once

#include <string>
#include <unordered_map>

namespace app {

struct Config {
  // [SERVER]
  std::string serverHost;
  int serverPort = 8080;

  // [TCP]
  std::string tcpHost;
  int tcpPort = 9090;

  // [DB]
  std::string dbType;  // maria | oracle | postgres
  std::string dbHost;
  int dbPort = 3306;
  std::string dbUser;
  std::string dbPassword;
  std::string dbName;

  // [REDIS]
  std::string redisHost;
  int redisPort = 6379;
  std::string redisPassword;

  // [JWT]
  std::string jwtSecret;
  int jwtExpireMin = 60;

  static Config load(const std::string& path);

 private:
  static std::unordered_map<std::string, std::string> parseIni(
      const std::string& path);
};

}  // namespace app
