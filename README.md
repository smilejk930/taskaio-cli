# taskAIO CLI

`taskaio`는 터미널과 자동화 환경에서 taskAIO의 프로젝트, 업무, 일정, 프로젝트 구성원을 관리하기 위한 Go 기반 명령줄 프로그램입니다.

Linux와 WSL에서 동작하며, 별도의 런타임 없이 단일 바이너리로 실행할 수 있습니다. 기본 출력은 자동화와 AI 에이전트가 처리하기 쉬운 JSON입니다.

## 주요 기능

- 프로젝트 목록·상세 조회, 등록, 수정, 삭제
- 프로젝트 구성원 및 역할(`owner`, `manager`, `member`) 조회
- 프로젝트 업무 목록·상세 조회, 등록, 수정, 삭제
- 업무 상태·우선순위·담당자·상위 업무 필터링
- 계층형 업무와 진행률 관리
- 일정 목록·상세 조회, 등록, 수정, 삭제
- JSON 파일 또는 표준 입력을 이용한 등록·수정
- 커서 페이지네이션과 `--all` 전체 조회
- 기본 JSON 출력 및 `--output table` 표 형식 출력
- PAT(Personal Access Token) 인증
- 자동화에 사용할 수 있는 표준 종료 코드

## 지원 환경

- Linux amd64
- Linux arm64
- WSL 2

## 설치

### 릴리즈 바이너리 설치

다음 명령은 운영체제와 CPU 아키텍처를 확인하고, 최신 릴리즈와 체크섬을 내려받아 `~/.local/bin/taskaio`에 설치합니다.

```bash
curl -fsSL https://raw.githubusercontent.com/smilejk930/taskaio-cli/main/install.sh | bash
```

특정 버전이나 설치 경로를 지정할 수도 있습니다.

```bash
TASKAIO_VERSION=v1.2.3 INSTALL_DIR=/usr/local/bin ./install.sh
```

`~/.local/bin`이 `PATH`에 없다면 셸 설정 파일에 다음 내용을 추가합니다.

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 소스에서 빌드

Go 툴체인이 설치된 환경에서 다음 명령을 사용합니다.

```bash
make build
./bin/taskaio --version
```

현재 사용자 계정에 설치하려면 다음 명령을 실행합니다.

```bash
make install
```

Linux amd64와 arm64 바이너리를 모두 빌드하려면 다음 명령을 사용합니다.

```bash
make build-all
```

생성 파일:

- `bin/taskaio-linux-amd64`
- `bin/taskaio-linux-arm64`

### 오프라인 서버에 설치

인터넷에 연결된 빌드 서버에서 대상 서버의 CPU 아키텍처에 맞는 바이너리를 준비합니다.

```bash
make build-all

cd bin
sha256sum taskaio-linux-amd64 > taskaio-linux-amd64.sha256
sha256sum taskaio-linux-arm64 > taskaio-linux-arm64.sha256
```

대상 서버의 CPU 아키텍처는 다음 명령으로 확인할 수 있습니다.

```bash
uname -m
```

아키텍처별로 옮겨야 할 파일은 다음과 같습니다.

| `uname -m` 출력 | 바이너리 | 체크섬 파일 |
| --- | --- | --- |
| `x86_64`, `amd64` | `taskaio-linux-amd64` | `taskaio-linux-amd64.sha256` |
| `aarch64`, `arm64` | `taskaio-linux-arm64` | `taskaio-linux-arm64.sha256` |

해당 바이너리와 체크섬 파일을 사내 파일 전송망, 이동식 저장장치 등의 허용된 방법으로 오프라인 서버의 같은 디렉터리에 옮깁니다. 오프라인 서버에서 체크섬을 검증한 뒤 현재 사용자 계정에 설치합니다. 다음 예시는 Linux amd64 서버를 기준으로 합니다.

```bash
sha256sum -c taskaio-linux-amd64.sha256

mkdir -p "$HOME/.local/bin"
install -m 0755 taskaio-linux-amd64 "$HOME/.local/bin/taskaio"

export PATH="$HOME/.local/bin:$PATH"
taskaio --version
```

Linux arm64 서버에서는 위 명령의 `taskaio-linux-amd64`를 `taskaio-linux-arm64`로 바꿉니다.

모든 사용자에게 시스템 공용으로 설치하려면 관리자 권한으로 `/usr/local/bin`에 설치합니다.

```bash
sudo install -m 0755 taskaio-linux-amd64 /usr/local/bin/taskaio
taskaio --version
```

`~/.local/bin`을 계속 `PATH`에 포함하려면 사용하는 셸의 설정 파일(예: `~/.bashrc`)에 다음 내용을 추가합니다.

```bash
grep -qxF 'export PATH="$HOME/.local/bin:$PATH"' "$HOME/.bashrc" \
  || printf '\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$HOME/.bashrc"
source "$HOME/.bashrc"
```

## 빠른 시작

```bash
# 설정 파일 생성
taskaio config init

# PAT 저장 및 인증 확인
taskaio auth login
taskaio auth status

# 프로젝트와 업무 조회
taskaio projects list
taskaio tasks list --project <projectId>

# 일정 조회
taskaio schedules list
```

## 인증

taskAIO 서버에서 발급한 PAT를 사용합니다.

### 대화형 로그인

```bash
taskaio auth login
```

### 표준 입력으로 로그인

CI/CD 또는 자동화 환경에서는 토큰을 표준 입력으로 전달할 수 있습니다.

```bash
printf '%s' "$TASKAIO_TOKEN" | taskaio auth login --token-stdin
```

### 명령 옵션으로 로그인

```bash
taskaio auth login --token "pat_xxxxxxxx"
```

명령 옵션은 셸 기록이나 프로세스 목록에 토큰이 남을 수 있으므로, 자동화 환경에서는 `TASKAIO_TOKEN` 환경변수나 `--token-stdin` 사용을 권장합니다.

현재 인증 정보를 확인하려면 다음 명령을 사용합니다.

```bash
taskaio auth status
taskaio auth status --output table
```

## 환경설정

기본 설정 파일 경로는 다음과 같습니다.

```text
~/.config/taskaio/config.yaml
```

`XDG_CONFIG_HOME`이 설정된 경우에는 `$XDG_CONFIG_HOME/taskaio/config.yaml`을 사용합니다. 설정 디렉터리는 `0700`, 토큰이 저장되는 설정 파일은 `0600` 권한으로 생성됩니다.

예시 설정은 [config.example.yaml](./config.example.yaml)과 [.env.example](./.env.example)을 참고하세요.

```yaml
base_url: http://localhost:3000
token: pat_xxxxxxxxxxxxxxxxxxxxxxxx
output: json
timeout: 30s
```

설정 적용 우선순위는 다음과 같습니다.

1. 명령 옵션
2. `TASKAIO_*` 환경변수
3. 설정 파일
4. 기본값

지원 환경변수:

| 환경변수 | 설명 | 기본값 |
| --- | --- | --- |
| `TASKAIO_BASE_URL` | taskAIO 서버 주소 | `http://localhost:3000` |
| `TASKAIO_TOKEN` | PAT | 없음 |
| `TASKAIO_OUTPUT` | 출력 형식(`json`, `table`) | `json` |
| `TASKAIO_TIMEOUT` | HTTP 요청 제한 시간 | `30s` |
| `TASKAIO_CONFIG` | 설정 파일 경로 | XDG 기본 경로 |

설정 파일을 생성하거나 덮어쓰려면 다음 명령을 사용합니다.

```bash
taskaio config init
taskaio config init --force
```

## 공통 옵션

| 옵션 | 설명 |
| --- | --- |
| `--base-url <URL>` | taskAIO 서버 주소 지정 |
| `--token <PAT>` | 요청에 사용할 PAT 지정 |
| `--config <경로>` | 설정 파일 경로 지정 |
| `--output json\|table` | 출력 형식 지정 |
| `--timeout <시간>` | HTTP 요청 제한 시간 지정(예: `30s`) |
| `-h`, `--help` | 도움말 출력 |
| `-v`, `--version` | 버전 출력 |

## 명령어

### 프로젝트

```bash
# 목록 조회
taskaio projects list
taskaio projects list --search "백엔드" --output table
taskaio projects list --limit 100 --cursor <cursor>
taskaio projects list --all

# 상세 조회
taskaio projects get <projectId>

# 등록
taskaio projects create --name "신규 프로젝트" --description "프로젝트 설명"
taskaio projects create --input project.json
taskaio projects create --input - < project.json

# 수정
taskaio projects update <projectId> --name "변경된 이름"
taskaio projects update <projectId> --input project-update.json

# 삭제
taskaio projects delete <projectId>
taskaio projects delete <projectId> --yes

# 프로젝트 구성원 목록
taskaio projects members list <projectId>
taskaio projects members list <projectId> --all --output table
```

### 업무

업무 목록 명령에는 프로젝트 ID가 필요합니다.

```bash
# 목록 조회
taskaio tasks list --project <projectId>
taskaio tasks list --project <projectId> --status in_progress
taskaio tasks list --project <projectId> --priority high --assignee <userId>
taskaio tasks list --project <projectId> --parent <parentTaskId>
taskaio tasks list --project <projectId> --all

# 상세 조회
taskaio tasks get <taskId>

# 등록
taskaio tasks create --project <projectId> --title "REST API 구현" --priority high
taskaio tasks create --project <projectId> --input task.json

# 수정
taskaio tasks update <taskId> --progress 100 --status done
taskaio tasks update <taskId> --input task-update.json

# 삭제
taskaio tasks delete <taskId> --yes
```

일반 사용자는 본인이 담당자로 지정된 업무만 등록·수정·삭제할 수 있습니다. 시스템 관리자는 모든 프로젝트의 업무를 관리할 수 있습니다.

### 일정

```bash
# 목록 조회
taskaio schedules list
taskaio schedules list --from 2026-08-01 --to 2026-08-31
taskaio schedules list --type member_leave --output table
taskaio schedules list --all

# 상세 조회
taskaio schedules get <scheduleId>

# 등록
taskaio schedules create \
  --name "연차" \
  --start-date 2026-09-01 \
  --end-date 2026-09-02 \
  --type member_leave
taskaio schedules create --input schedule.json

# 수정
taskaio schedules update <scheduleId> --name "오후 반차"
taskaio schedules update <scheduleId> --input schedule-update.json

# 삭제
taskaio schedules delete <scheduleId> --yes
```

일정 권한은 프로젝트, 일정 유형, 대상 팀원을 기준으로 나누지 않습니다. 인증된 사용자는 모든 일정에 대해 조회·등록·수정·삭제할 수 있습니다. 일정 유형과 대상 팀원은 일정 정보를 분류하기 위한 필드입니다.

### 버전과 도움말

```bash
taskaio version
taskaio --version
taskaio help
taskaio projects --help
taskaio tasks create --help
```

## JSON 입력

등록과 수정 명령은 필드 옵션 대신 `--input <파일|->`을 사용할 수 있습니다.

```bash
taskaio projects create --input project.json
taskaio tasks create --project <projectId> --input - < task.json
taskaio schedules update <scheduleId> --input schedule-update.json
```

`--input`과 요청 본문을 구성하는 필드 옵션은 함께 사용할 수 없습니다. 단, 업무 등록의 `--project`처럼 요청 대상을 지정하는 옵션은 `--input`과 함께 사용합니다.

업무 등록 예시:

```json
{
  "title": "CLI 문서 작성",
  "description": "사용 예제와 설치 방법을 정리합니다.",
  "status": "in_progress",
  "priority": "medium",
  "assigneeId": "user-id",
  "startDate": "2026-08-26",
  "endDate": "2026-08-28",
  "progress": 50
}
```

## 출력

기본 JSON 결과는 표준 출력(`stdout`)으로 전달됩니다. 오류와 진단 메시지는 표준 오류(`stderr`)로 분리되므로 파이프라인에서 안전하게 사용할 수 있습니다.

```bash
taskaio projects list | jq '.[] | .id'
taskaio tasks list --project <projectId> --all > tasks.json
taskaio schedules list --output table
```

## 종료 코드

| 종료 코드 | 의미 | 예시 |
| --- | --- | --- |
| `0` | 성공 | 명령 정상 완료 |
| `2` | 입력 또는 설정 오류 | 잘못된 옵션, 필수값 누락, 400/422 응답, 설정 오류 |
| `3` | 인증 실패 | PAT 누락 또는 401 응답 |
| `4` | 권한 부족 | 403 응답 |
| `5` | 리소스 없음 또는 접근 불가 | 404 응답 |
| `6` | 충돌 | 409 응답 |
| `7` | API, 서버 또는 네트워크 오류 | 5xx 응답, 연결 실패, 제한 시간 초과 |

셸 스크립트에서는 종료 코드를 이용해 오류를 구분할 수 있습니다.

```bash
if ! taskaio auth status >/dev/null; then
  code=$?
  echo "인증 확인 실패: $code" >&2
fi
```

## 개발

```bash
go test ./...
go vet ./...
make build-all
```

릴리즈는 `v*` 형식의 Git 태그가 푸시되면 GitHub Actions와 GoReleaser를 통해 생성됩니다. Linux amd64/arm64 압축 파일과 `checksums.txt`가 함께 배포됩니다.

## 라이선스

MIT License
