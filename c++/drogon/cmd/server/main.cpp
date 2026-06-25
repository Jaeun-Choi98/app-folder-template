#include "internal/app/app.h"
#include "internal/container/container.h"
#include "internal/logger/logger.h"

#include <exception>
#include <iostream>

int main() {
    try {
        auto container = app::Container::build("env.ini");
        app::Application application(std::move(container));

        application.start();
        application.waitForSignal();
        application.shutdown();

        return 0;
    } catch (const std::exception& e) {
        std::cerr << "fatal: " << e.what() << std::endl;
        return 1;
    }
}
