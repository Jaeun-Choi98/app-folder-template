# third_party

외부 의존성을 이 디렉터리에 배치합니다.
CMakeLists.txt가 이 경로를 자동으로 인식하여 `add_subdirectory()`로 빌드합니다.

## 설치 방법

프로젝트 루트(`c++/drogon/`)에서 아래 명령을 실행합니다:

```bash
# Drogon (HTTP 프레임워크)
git clone --depth 1 --recurse-submodules https://github.com/drogonframework/drogon.git third_party/drogon

# standalone Asio (header-only, TCP 비동기)
git clone --depth 1 https://github.com/chriskohlhoff/asio.git third_party/asio

# SOCI (DB 접근)
git clone --depth 1 https://github.com/SOCI/soci.git third_party/soci

# hiredis (Redis 클라이언트)
git clone --depth 1 https://github.com/redis/hiredis.git third_party/hiredis
```

## 배치 후 구조

```
third_party/
├── drogon/          # Drogon 프레임워크 소스
│   ├── CMakeLists.txt
│   └── ...
├── asio/            # standalone Asio (header-only)
│   └── asio/
│       └── include/
│           ├── asio.hpp      ← CMake가 이 파일을 찾음
│           └── asio/
├── soci/            # SOCI DB 라이브러리 소스
│   ├── CMakeLists.txt
│   └── ...
└── hiredis/         # hiredis Redis 클라이언트 소스
    ├── CMakeLists.txt
    └── ...
```

## 헤더 참조 경로

코드에서 아래와 같이 include합니다:

```cpp
#include <asio.hpp>              // standalone Asio
#include <drogon/drogon.h>       // Drogon
#include <soci/soci.h>           // SOCI
#include <hiredis/hiredis.h>     // hiredis
```

## 참고

- `third_party/` 디렉터리는 `.gitignore`에 추가하거나, git submodule로 관리할 수 있습니다.
- 시스템에 이미 설치된 패키지가 있다면 `third_party/`에 넣지 않아도 됩니다.
  CMakeLists.txt가 `third_party/` → 시스템 패키지 순서로 탐색합니다.
