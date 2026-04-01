#!/bin/bash
# KCP PostgreSQL 컨테이너 실행 스크립트
# nerdctl을 사용하여 PostgreSQL 16을 컨테이너로 실행한다

CONTAINER_NAME="kcp-postgres"
PG_PORT="${PG_PORT:-5432}"
PG_USER="${PG_USER:-kcp}"
PG_PASSWORD="${PG_PASSWORD:-kcppassword}"
PG_DB="${PG_DB:-kcp}"

# 기존 컨테이너 확인 및 제거
if nerdctl ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "기존 컨테이너 제거 중: ${CONTAINER_NAME}"
    nerdctl rm -f ${CONTAINER_NAME}
fi

echo "PostgreSQL 컨테이너 시작 중..."
nerdctl run -d \
    --name ${CONTAINER_NAME} \
    -p ${PG_PORT}:5432 \
    -e POSTGRES_USER=${PG_USER} \
    -e POSTGRES_PASSWORD=${PG_PASSWORD} \
    -e POSTGRES_DB=${PG_DB} \
    -v kcp-pgdata:/var/lib/postgresql/data \
    postgres:16-alpine

echo "PostgreSQL 컨테이너 시작 완료"
echo "연결 URL: postgresql://${PG_USER}:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable"
