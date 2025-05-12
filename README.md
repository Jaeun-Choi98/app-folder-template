# app-folder-template

I want to manage and handle the project more efficiently by standardizing the app folder structure. By doing so, I aim to help others easily understand and work with the system.

```
root/ <br>
├── cmd/              # 애플리케이션 엔트리 포인트
│ └── server/
│ └── main.go
├── internal/         # 애플리케이션 내부 코드
│ ├── config/         # 설정
│ ├── container/      # 의존성 컨테이너
│ ├── app/            # 애플리케이션 시작/종료 로직
│ ├── model/          # 도메인 모델
│ ├── service/        # 비즈니스 로직
│ ├── repository/     # 데이터 접근 계층
│ │ ├── entity/       # 데이터베이스 엔티티
│ │ ├── dao/          # 데이터 액세스 객체
│ │ └── mapper/       # 엔티티-모델 매핑
│ ├── transport/      # 전송 프로토콜 관련 코드
│ │ ├── http/         # HTTP 서버 관련
│ │ │ ├── controller/ # HTTP 컨트롤러
│ │ │ ├── dto/        # HTTP 요청/응답 모델
│ │ │ ├── mapper/     # 모델-DTO 매핑
│ │ │ └── middleware/ # HTTP 미들웨어
│ │ ├── tcp/          # TCP 서버 관련
│ │ │ ├── handler/    # TCP 핸들러
│ │ │ ├── protocol/   # TCP 프로토콜 정의
│ │ │ └── mapper/     # 프로토콜-모델 매핑
│ │ └── eventbus/     # 내부 이벤트 버스
│ └── infrastructure/ # 메모리 관련 코드
│ ├── ram/            # 메모리
│ └── shm/            # 공유 메모리
└── pkg/              # 외부 라이브러리 코드
```
