package cli

import (
	"fmt"
	"os"
	"syscall"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// loginCmd 은 Gateway 서버에 로그인하여 토큰을 저장하는 커맨드이다
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Gateway 서버에 로그인",
	Long: `서버 URL은 설정 파일(~/.kcp/kcp-config.yaml)에서 읽는다.
사용자명과 비밀번호를 입력받아 인증 토큰을 발급받고 설정 파일에 저장한다.

설정 파일 위치 변경:
  kcp login --config /path/to/config.yaml`,
	RunE: runLogin,
}

// logoutCmd 은 저장된 인증 토큰을 삭제하는 커맨드이다
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "로그아웃 (토큰 삭제)",
	Long:  "설정 파일에서 인증 토큰을 제거한다.",
	RunE:  runLogout,
}

// loginResponse 는 로그인 API 응답 구조체이다
type loginResponse struct {
	Token string `json:"token"`
}

// runLogin 은 설정 파일에서 서버 URL을 읽고, 사용자 입력으로 로그인을 수행한다
func runLogin(_ *cobra.Command, _ []string) error {
	cfgPath := config.ResolvePath(cfgFile)

	// 설정 파일 로드 (없으면 기본값 생성)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("설정 파일 로드 실패: %w", err)
	}

	fmt.Printf("서버 URL: %s\n", cfg.ServerURL)

	// 사용자명 입력
	var username string
	fmt.Print("사용자명: ")
	if _, err := fmt.Scan(&username); err != nil {
		return fmt.Errorf("사용자명 입력 실패: %w", err)
	}

	// 비밀번호 입력 (화면에 표시하지 않음)
	fmt.Print("비밀번호: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // 줄바꿈 (ReadPassword는 개행을 출력하지 않으므로)
	if err != nil {
		return fmt.Errorf("비밀번호 입력 실패: %w", err)
	}
	password := string(passwordBytes)

	// Gateway 로그인 API 호출
	client := sdk.NewClient(cfg.ServerURL)
	body := map[string]string{
		"username": username,
		"password": password,
		"authType": cfg.AuthType,
	}
	var resp loginResponse
	if err := client.Post("/auth/login", body, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "로그인 실패: %v\n", err)
		return err
	}

	// 설정 파일에 토큰 저장 (username은 저장하지 않음)
	cfg.Token = resp.Token
	if err := config.Save(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "설정 저장 실패: %v\n", err)
		return err
	}

	fmt.Printf("로그인 성공 (%s)\n", username)
	fmt.Printf("설정 파일: %s\n", cfgPath)
	return nil
}

// runLogout 은 설정 파일에서 토큰을 제거한다
func runLogout(_ *cobra.Command, _ []string) error {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Println("이미 로그아웃 상태입니다.")
		return nil
	}

	if cfg.Token == "" {
		fmt.Println("이미 로그아웃 상태입니다.")
		return nil
	}

	// 토큰 제거 후 저장
	cfg.Token = ""
	if err := config.Save(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "설정 저장 실패: %v\n", err)
		return err
	}

	fmt.Println("로그아웃 완료")
	return nil
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
