# 베이스 이미지
FROM quay.io/centos/centos:stream9

# 작업 디렉토리 설정
WORKDIR /app

# 소스 복사
COPY ./goWeb .
COPY ./goWeb.ini .

# 의존성 설치 및 빌드 go 는 필요 없음 바이너리만 넘길거라
# RUN go mod tidy && go build -o test .

# 실행 명령
CMD ["./goWeb", "-f", "./goWeb.ini"]
