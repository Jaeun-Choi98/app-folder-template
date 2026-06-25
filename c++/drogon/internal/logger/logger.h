#pragma once

#include <string>
#include <fstream>
#include <mutex>

namespace app {

enum class LogLevel { Debug, Info, Warn, Error };

class Logger {
public:
    static Logger& instance();

    void init(const std::string& baseDir);
    void info(const std::string& msg);
    void debug(const std::string& msg);
    void warn(const std::string& msg);
    void error(const std::string& msg);
    void cleanOldLogs(int keepDays = 3);

private:
    Logger() = default;
    void write(LogLevel level, const std::string& msg);
    std::string currentLogDir() const;

    std::string baseDir_;
    std::mutex  mtx_;
};

} // namespace app
