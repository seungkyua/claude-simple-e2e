package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// auditCmd 은 감사 로그 관련 상위 커맨드이다
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "감사 로그 관리",
}

// 감사 로그 조회 필터 플래그
var (
	auditUser   string
	auditAction string
	auditFrom   string
	auditTo     string
)

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "감사 로그 조회",
	Long:  "사용자, 액션, 기간 필터를 이용하여 감사 로그를 조회한다.",
	RunE:  runAuditList,
}

// auditEntry 는 감사 로그 항목을 나타낸다
type auditEntry struct {
	ID        string `json:"id"`
	User      string `json:"user"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Timestamp string `json:"timestamp"`
	Detail    string `json:"detail"`
}

func runAuditList(_ *cobra.Command, _ []string) error {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "설정 로드 실패: %v\n", err)
		return err
	}

	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))

	// 필터 파라미터를 쿼리 경로에 추가한다
	path := "/api/v1/audit/logs?"
	params := ""
	if auditUser != "" {
		params += "user=" + auditUser + "&"
	}
	if auditAction != "" {
		params += "action=" + auditAction + "&"
	}
	if auditFrom != "" {
		params += "from=" + auditFrom + "&"
	}
	if auditTo != "" {
		params += "to=" + auditTo + "&"
	}
	path += params

	var entries []auditEntry
	if err := client.Get(path, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "감사 로그 조회 실패: %v\n", err)
		return err
	}

	headers := []string{"ID", "사용자", "액션", "리소스", "시간"}
	var rows [][]string
	for _, e := range entries {
		rows = append(rows, []string{e.ID, e.User, e.Action, e.Resource, e.Timestamp})
	}
	formatOutput(outputFormat, headers, rows, entries)
	return nil
}

func init() {
	auditListCmd.Flags().StringVar(&auditUser, "user", "", "사용자 필터")
	auditListCmd.Flags().StringVar(&auditAction, "action", "", "액션 필터")
	auditListCmd.Flags().StringVar(&auditFrom, "from", "", "시작 일시 (RFC3339)")
	auditListCmd.Flags().StringVar(&auditTo, "to", "", "종료 일시 (RFC3339)")

	auditCmd.AddCommand(auditListCmd)
	rootCmd.AddCommand(auditCmd)
}
