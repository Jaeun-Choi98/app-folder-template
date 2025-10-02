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
flowchart LR
    subgraph backend
        direction RL
        
        subgraph subprocess
            direction RL
            health@{ shape: lin-rect, label: "monitoring process" }
            log@{ shape: lin-rect, label: "log manager process" }
        end
        
        subgraph EventBus["EventBus Instance"]
            direction TB
            topics["Topics:<br/>- tcpclient.send<br/>- system.state<br/>- cache.data"]
            topics@{ shape: hex }
        end
        
        tcp(tcp thread)
        api(api thread)
        worker(tcp worker thread)
        cache(cache mem)
        dblayer(db layer)
        dbmanager(db manager)
        
        %% API 흐름
        api -->|subscribe: cache.data| EventBus
        api -->|subscribe: system.state| EventBus
        api -->|publish: tcpclient.send| EventBus
        api --- cache
        
        %% TCP 흐름
        tcp -->|publish: tcpclient.send| EventBus
        tcp --- cache
        worker --> tcp

        %% Worker 흐름
        EventBus -->|subscribe: tcpclient.send| worker
        
        %% Cache 흐름
        cache -->|publish: cache.data| EventBus
        cache -->|publish: system.state| EventBus
        cache --- dblayer
        cache cache@--> |state trace|cache
        cache@{ animate: true }
        
        %% DB 흐름
        dblayer ---> DB@{ shape: cyl, label: "Database" }
        dbmanager --->|log,tableManage| dblayer
    end
    
    subgraph tcp_client["TCP Client"]
        client[client]
        tcp <--->|protocol| client
    end
    
    subgraph web_client["Web Client"]
        chrome[browser]
        api sse@-->|SSE stream| chrome
        api <-->|REST API| chrome
        sse@{ animate: true }
    end
    health --> |publish: system.state| EventBus
```
