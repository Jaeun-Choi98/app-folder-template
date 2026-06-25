#include "internal/worker/cron_worker.h"
#include "internal/logger/logger.h"

#include <chrono>

namespace app {

void CronWorker::start() {
    running_ = true;
    thread_  = std::thread(&CronWorker::run, this);
}

void CronWorker::stop() {
    running_ = false;
    if (thread_.joinable()) thread_.join();
}

void CronWorker::run() {
    while (running_) {
        // TODO: 스케줄된 작업 실행
        std::this_thread::sleep_for(std::chrono::seconds(60));
    }
}

} // namespace app
