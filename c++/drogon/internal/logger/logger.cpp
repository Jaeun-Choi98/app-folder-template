#include "internal/logger/logger.h"

#include <chrono>
#include <filesystem>
#include <iomanip>
#include <iostream>
#include <sstream>

namespace fs = std::filesystem;

namespace app {

Logger& Logger::instance() {
    static Logger inst;
    return inst;
}

void Logger::init(const std::string& baseDir) {
    baseDir_ = baseDir;
    fs::create_directories(currentLogDir());
}

std::string Logger::currentLogDir() const {
    auto now = std::chrono::system_clock::now();
    auto t   = std::chrono::system_clock::to_time_t(now);
    std::tm tm{};
#ifdef _WIN32
    localtime_s(&tm, &t);
#else
    localtime_r(&t, &tm);
#endif
    std::ostringstream oss;
    oss << baseDir_ << "/"
        << std::put_time(&tm, "%Y/%m/%d");
    return oss.str();
}

void Logger::write(LogLevel level, const std::string& msg) {
    std::lock_guard lock(mtx_);

    auto dir = currentLogDir();
    fs::create_directories(dir);

    std::string filename = (level == LogLevel::Debug) ? "debug" : "info";
    std::string path = dir + "/" + filename;

    std::ofstream ofs(path, std::ios::app);
    if (!ofs.is_open()) return;

    auto now = std::chrono::system_clock::now();
    auto t   = std::chrono::system_clock::to_time_t(now);
    std::tm tm{};
#ifdef _WIN32
    localtime_s(&tm, &t);
#else
    localtime_r(&t, &tm);
#endif

    static const char* labels[] = {"DEBUG", "INFO", "WARN", "ERROR"};
    ofs << std::put_time(&tm, "%Y-%m-%d %H:%M:%S")
        << " [" << labels[static_cast<int>(level)] << "] "
        << msg << "\n";
}

void Logger::info(const std::string& msg)  { write(LogLevel::Info, msg); }
void Logger::debug(const std::string& msg) { write(LogLevel::Debug, msg); }
void Logger::warn(const std::string& msg)  { write(LogLevel::Warn, msg); }
void Logger::error(const std::string& msg) { write(LogLevel::Error, msg); }

void Logger::cleanOldLogs(int keepDays) {
    // TODO: baseDir_ 아래에서 keepDays 이전 날짜 디렉터리를 삭제
}

} // namespace app
