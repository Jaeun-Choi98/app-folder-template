# app-folder-template

I want to manage and handle the project more efficiently by standardizing the app folder structure. By doing so, I aim to help others easily understand and work with the system.

++ examples using gorilla no longer be writed. reference the gin folder

```
pjt/
├── cmd/              # 애플리케이션 엔트리 포인트
│   └── server/
│       ├── log/
│       └── env.ini
├── internal/         # 애플리케이션 내부 코드
│   ├── logger/       # 커스텀 로거
│   ├── config/       # 설정
│   ├── container/    # 의존성 컨테이너
│   ├── app/          # 애플리케이션 시작/종료 로직
│   ├── service/      # 비즈니스 로직
│   ├── worker/       # 루틴이 필요한 작업
│   ├── reids/        # 레디스
│   ├── eventbus/     # 이벤트 버스
│   ├── healthcheck/  # 시스템 헬스 체크
│   ├── db/           # 데이터 접근 계층
│   │   ├── entity/
│   │   └── db-handler/
│   ├── transport/    # 실행 프로세스 관련 코드
│   │   ├── http/
│   │   └── tcp/
│   └── infra/        # 메모리 관련 코드
│       ├── cache/
│       └── object/
└── pkg/              # 외부 라이브러리 코드
```

## architecture

```mermaid
flowchart TB
    subgraph backend["Backend System"]
        direction TB

        subgraph subprocess["Subprocess Layer"]
            direction LR
            health["Healthcheck<br/>Process"]
            stdlog["Std Log<br/>Manager"]
            dblog["DB Log<br/>Manager"]
        end

        subgraph EventBus["EventBus Instance"]
            topics["Topics:<br/>• tcpclient.send<br/>• system.state<br/>• system.event<br/>• cache.data"]
            topics@{ shape: hex }
        end

        subgraph threads["Transport Layer"]
            direction LR
            api["API Thread<br/>(REST + SSE)"]
            tcp["TCP Thread"]
        end

        subgraph workers["Worker Layer"]
            direction LR
            tcpworker["TCP Worker<br/>Thread"]
            cacheworker["Cache Worker<br/>Thread"]
            cacheworker@{ animate: true }
        end

        subgraph infra["Infrastructure Layer"]
            direction LR
            cache["Cache<br/>Memory"]
            dblayer["DB Layer"]
        end

        %% EventBus 구독 (EventBus --- Subscriber)
        EventBus ---|sub: tcpclient.send| tcpworker
        EventBus ---|sub: cache.data| api
        EventBus ---|sub: system.state| api
        EventBus ---|sub: system.event| api

        %% EventBus 발행 (Publisher --- EventBus)
        api ---|pub: tcpclient.send| EventBus
        cacheworker ---|pub: cache.data| EventBus
        cacheworker ---|pub: system.state| EventBus
        cacheworker ---|pub: system.event| EventBus
        cacheworker -->|state trace| cacheworker
        health ---|pub: system.state| EventBus

        %% 데이터 접근
        api <--> cache
        tcp <--> cache
        tcpworker --> tcp
        cacheworker <--> cache
        cacheworker <--> dblayer
        dblog --> dblayer
    end

    subgraph external["External Systems"]
        direction TB

        subgraph database["Database"]
            DB[("MariaDB /<br/>Oracle /<br/>Tibero")]
        end

        subgraph tcp_client["TCP Client"]
            client["Client<br/>Application"]
        end

        subgraph web_client["Web Client"]
            chrome["Browser"]
        end
    end

    %% 외부 통신
    dblayer <-.->|SQL Query| DB
    tcp <-->|TCP Protocol| client
    api ==>|SSE Stream:<br/>• cache.data<br/>• system.state<br/>• system.event| chrome
    api <-->|REST API| chrome

    style EventBus fill:#e1f5ff
    style backend fill:#f9f9f9
    style external fill:#fff9e6
    style database fill:#ffe6e6
```
