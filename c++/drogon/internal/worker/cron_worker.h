#pragma once

#include <atomic>
#include <thread>

namespace app {

class CronWorker {
public:
    void start();
    void stop();

private:
    void run();

    std::thread       thread_;
    std::atomic<bool> running_{false};
};

} // namespace app
