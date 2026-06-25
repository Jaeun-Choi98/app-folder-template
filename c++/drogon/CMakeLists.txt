cmake_minimum_required(VERSION 3.16)
project(appfolder CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Threads REQUIRED)

# ── third_party 경로 ───────────────────────────────────
set(THIRD_PARTY_DIR ${CMAKE_SOURCE_DIR}/third_party)

# ── Drogon ─────────────────────────────────────────────
# third_party/drogon/ 에 소스가 있으면 서브디렉터리로 빌드,
# 없으면 시스템 설치된 패키지를 탐색.
if(EXISTS ${THIRD_PARTY_DIR}/drogon/CMakeLists.txt)
    add_subdirectory(${THIRD_PARTY_DIR}/drogon drogon_build)
else()
    find_package(Drogon REQUIRED)
endif()

# ── SOCI ───────────────────────────────────────────────
if(EXISTS ${THIRD_PARTY_DIR}/soci/CMakeLists.txt)
    set(SOCI_TESTS OFF CACHE BOOL "" FORCE)
    set(WITH_MYSQL ON CACHE BOOL "" FORCE)
    add_subdirectory(${THIRD_PARTY_DIR}/soci soci_build)
else()
    find_package(SOCI REQUIRED)
endif()

# ── standalone Asio (header-only) ──────────────────────
if(EXISTS ${THIRD_PARTY_DIR}/asio/asio/include/asio.hpp)
    set(ASIO_INCLUDE_DIR ${THIRD_PARTY_DIR}/asio/asio/include)
else()
    find_path(ASIO_INCLUDE_DIR asio.hpp)
    if(NOT ASIO_INCLUDE_DIR)
        message(FATAL_ERROR
            "asio not found.\n"
            "  → git clone https://github.com/chriskohlhoff/asio.git third_party/asio")
    endif()
endif()

# ── hiredis ────────────────────────────────────────────
if(EXISTS ${THIRD_PARTY_DIR}/hiredis/CMakeLists.txt)
    set(DISABLE_TESTS ON CACHE BOOL "" FORCE)
    add_subdirectory(${THIRD_PARTY_DIR}/hiredis hiredis_build)
    set(HIREDIS_LIB hiredis)
    set(HIREDIS_INCLUDE_DIR ${THIRD_PARTY_DIR}/hiredis)
else()
    find_library(HIREDIS_LIB hiredis)
    find_path(HIREDIS_INCLUDE_DIR hiredis/hiredis.h)
    if(NOT HIREDIS_LIB)
        message(FATAL_ERROR
            "hiredis not found.\n"
            "  → git clone https://github.com/redis/hiredis.git third_party/hiredis")
    endif()
endif()

# ── 소스 수집 ──────────────────────────────────────────
file(GLOB_RECURSE SOURCES
    "cmd/server/*.cpp"
    "internal/**/*.cpp"
)

# ── 실행 파일 ──────────────────────────────────────────
add_executable(${PROJECT_NAME} ${SOURCES})

target_include_directories(${PROJECT_NAME} PRIVATE
    ${CMAKE_SOURCE_DIR}
    ${ASIO_INCLUDE_DIR}
    ${HIREDIS_INCLUDE_DIR}
)

target_compile_definitions(${PROJECT_NAME} PRIVATE
    ASIO_STANDALONE
)

target_link_libraries(${PROJECT_NAME} PRIVATE
    Drogon::Drogon
    SOCI::soci_core
    SOCI::soci_mysql
    Threads::Threads
    ${HIREDIS_LIB}
)
