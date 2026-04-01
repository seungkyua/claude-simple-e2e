-- KCP Gateway 초기 데이터베이스 스키마
-- 실제 테이블 생성은 Gateway의 자동 마이그레이션이 수행한다.
-- 이 파일은 수동 초기화가 필요한 경우에만 사용한다.

-- 초기 관리자 계정 (비밀번호는 반드시 변경할 것)
-- 비밀번호 해시는 bcrypt 'admin123' 의 해시값 (실제 운영 시 변경 필수)
-- INSERT INTO users (username, password_hash, email, role)
-- VALUES ('admin', '$2a$12$LJ3m4ys4LM8JG.Stq5n9muaYNmG7eVW9dBz/DXGpGb7S5mLqKjIXO', 'admin@example.com', 'ADMIN');
