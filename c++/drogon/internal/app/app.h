#pragma once

#include "internal/container/container.h"

namespace app {

class Application {
public:
    explicit Application(Container container);

    void start();
    void shutdown();
    void waitForSignal();

private:
    Container c_;
};

} // namespace app
