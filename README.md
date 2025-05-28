# app-folder-template

I want to manage and handle the project more efficiently by standardizing the app folder structure. By doing so, I aim to help others easily understand and work with the system.

++ examples using gorilla no longer be writed. reference the gin folder

```
pjt/ <br>
├── cmd/              # 애플리케이션 엔트리 포인트
│ └── server/
│ └── main.go
│ └── log/            # 로그 파일
├── internal/         # 애플리케이션 내부 코드
│ ├── logger/         # 커스텀 로거
│ ├── config/         # 설정
│ ├── container/      # 의존성 컨테이너
│ ├── app/            # 애플리케이션 시작/종료 로직
│ ├── service/        # 비즈니스 로직
│ ├── model/          # 전역적으로 사용하는 모델
│ ├── db/             # 데이터 접근 계층
│ │ ├── db-model/     # 데이터베이스 엔티티
│ │ ├── db-handler/   # 데이터 액세스 객체
│ │ └── mapper/       # 엔티티-모델 매핑 -> entity 폴더로 병합
│ ├── server/         # 실행 프로세스 관련 코드
│ │ ├── http/         # HTTP 서버 관련
│ │ │ ├── controller/ # HTTP 컨트롤러
│ │ │ ├── http-model/ # HTTP 요청/응답 모델
│ │ │ ├── mapper/     # 모델-DTO 매핑 -> model 폴더로 병합 예정 
│ │ │ └── middleware/ # HTTP 미들웨어
│ │ ├── tcp/          # TCP 서버 관련
│ │ │ ├── handler/    # TCP 핸들러
│ │ │ ├── protocol/   # TCP 프로토콜 정의
│ │ │ └── mapper/     # 프로토콜-모델 매핑 -> protocol 폴더로 병합 예정
│ │ └── eventbus/     # 내부 이벤트 버스
│ └── infra/          # 메모리 관련 코드
│   ├── ram/          # 메모리
│   └── shm/          # 공유 메모리
└── pkg/              # 외부 라이브러리 코드
```
