# taskAIO CLI

터미널에서 taskAIO의 프로젝트, 업무, 일정을 조회하고 관리하는 도구입니다.

Linux amd64·arm64와 WSL 2를 지원합니다. 설치할 서버에는 Go나 Docker 없이 실행 파일 하나만 있으면 됩니다. 인터넷이 차단된 환경에서도 사용할 수 있지만, **taskAIO 서버에는 네트워크로 접속할 수 있어야 합니다.**

## 설치

### 인터넷이 연결된 환경

```bash
curl -fsSL https://raw.githubusercontent.com/smilejk930/taskaio-cli/main/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
taskaio --version
```

공개된 최신 릴리즈를 `~/.local/bin/taskaio`에 설치합니다. 새 터미널에서도 사용하려면 위 `export` 줄을 사용하는 셸의 설정 파일(예: `~/.bashrc`)에 추가하세요.

### 오프라인 서버에 배포

**인터넷이 되는 PC에서 파일을 준비하고, 오프라인 서버로 옮겨 설치합니다.** 다음 빌드 명령은 Linux 또는 WSL에서 실행합니다.

**1. 오프라인 서버의 CPU 확인**

오프라인 서버에서 실행합니다.

```bash
uname -m
```

| 결과 | 사용할 실행 파일 |
| --- | --- |
| `x86_64`, `amd64` | `taskaio-linux-amd64` |
| `aarch64`, `arm64` | `taskaio-linux-arm64` |

**2. 인터넷이 되는 PC에서 파일 준비**

Go 1.27 이상과 `make`, Git이 필요합니다. 이미 소스가 있다면 해당 저장소 폴더에서 `make build-all`부터 실행하세요.

```bash
git clone https://github.com/smilejk930/taskaio-cli.git
cd taskaio-cli
make build-all

cd bin
sha256sum taskaio-linux-amd64 > taskaio-linux-amd64.sha256
sha256sum taskaio-linux-arm64 > taskaio-linux-arm64.sha256
```

**3. 파일 2개를 오프라인 서버로 전송**

서버 CPU에 맞는 **실행 파일과 같은 이름의 `.sha256` 파일**을 사내 파일 전송망이나 USB 등으로 옮깁니다. 두 파일은 서버의 같은 폴더에 두세요.

예를 들어 `x86_64` 서버에는 다음 두 파일만 필요합니다.

```text
taskaio-linux-amd64
taskaio-linux-amd64.sha256
```

**4. 오프라인 서버에 설치**

옮긴 파일이 있는 폴더에서 먼저 파일이 손상되지 않았는지 확인합니다.

```bash
sha256sum -c taskaio-linux-amd64.sha256
```

`taskaio-linux-amd64: OK`가 나오면 설치합니다. 아래 명령은 관리자 권한이 필요합니다.

```bash
sudo install -m 0755 taskaio-linux-amd64 /usr/local/bin/taskaio
taskaio --version
```

ARM64 서버라면 위 명령의 `taskaio-linux-amd64`를 모두 `taskaio-linux-arm64`로 바꾸세요.

업데이트할 때도 새 파일을 준비해 같은 순서로 설치하면 됩니다. 기존 서버 주소와 로그인 정보는 유지됩니다.

## taskAIO 서버 연결

1. taskAIO 웹에 로그인한 뒤 **환경 설정 → 개인 액세스 토큰**에서 PAT를 발급합니다.
2. 아래 서버 주소를 실제 taskAIO 주소로 바꿔 실행하고, 안내가 나오면 PAT를 입력합니다.

```bash
taskaio --base-url http://192.168.0.10:3001 auth login
taskaio auth status
```

taskAIO와 CLI를 같은 서버에서 실행한다면 `http://127.0.0.1:3001`을 사용할 수 있습니다. 포트는 실제 웹 서버 설정에 맞추세요. 주소 뒤에 `/api/v1`은 붙이지 않습니다.

로그인에 성공하면 서버 주소와 PAT가 기본적으로 `~/.config/taskaio/config.yaml`에 저장됩니다. 이후에는 서버 주소를 매번 입력하지 않아도 됩니다.

자동화용 환경변수와 추가 설정은 [.env.example](.env.example)과 [config.example.yaml](config.example.yaml)을 참고하세요.

## 기본 사용법

기본 출력은 JSON입니다. 사람이 읽기 쉬운 표로 보려면 `--output table`을 붙이세요. 목록 전체를 가져오려면 `--all`을 사용합니다.

```bash
# 프로젝트 목록에서 사용할 프로젝트 ID 확인
taskaio projects list --output table

# 업무 목록과 요약 조회
taskaio tasks list --project <projectId> --output table
taskaio tasks summary --project <projectId> --output table

# 이번 주 업무, 진행 중인 업무 조회
taskaio tasks list --project <projectId> --week 이번주
taskaio tasks list --project <projectId> --status 진행중

# 업무 등록 및 완료 처리
taskaio tasks create --project <projectId> --title "개발서버 배포"
taskaio tasks update <taskId> --status done

# 일정 목록 조회
taskaio schedules list --output table
```

`<projectId>`와 `<taskId>`는 조회한 실제 ID로 바꿉니다. 실행 가능한 작업은 로그인한 사용자의 서버 권한에 따릅니다.

프로젝트·업무·일정의 등록, 수정, 삭제와 세부 옵션은 도움말에서 확인할 수 있습니다.

```bash
taskaio --help
taskaio projects --help
taskaio tasks list --help
taskaio schedules --help
```

## 개발

```bash
make build       # bin/taskaio 생성
go test ./...    # 테스트
go vet ./...     # 정적 검사
```

라이선스: [MIT](LICENSE)
