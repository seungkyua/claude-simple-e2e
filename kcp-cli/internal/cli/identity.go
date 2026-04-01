package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// --- Project 커맨드 ---

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "프로젝트 관리",
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "프로젝트 목록 조회",
	RunE:  runProjectList,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "프로젝트 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 프로젝트 생성 폼 구현")
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "프로젝트 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

// --- User 커맨드 ---

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "사용자 관리",
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "사용자 목록 조회",
	RunE:  runUserList,
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "사용자 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 사용자 생성 폼 구현")
		return nil
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "사용자 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserDelete,
}

// --- Role 커맨드 ---

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "역할 관리",
}

// roleAssign 플래그
var (
	roleUserID    string
	roleProjectID string
	roleRoleID    string
)

var roleAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "사용자에게 역할 부여",
	RunE:  runRoleAssign,
}

var roleRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "사용자에게서 역할 회수",
	RunE:  runRoleRevoke,
}

// newIdentityClient 는 설정 파일을 로드하여 IdentityClient를 생성한다
func newIdentityClient() (sdk.IdentityClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewIdentityClient(client), nil
}

func runProjectList(_ *cobra.Command, _ []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := ic.ListProjects(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "프로젝트 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "설명", "활성"}
	var rows [][]string
	for _, p := range resp.Items {
		enabled := "N"
		if p.Enabled {
			enabled = "Y"
		}
		rows = append(rows, []string{p.ID, p.Name, p.Description, enabled})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runProjectDelete(_ *cobra.Command, args []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := ic.DeleteProject(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "프로젝트 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("프로젝트 삭제 완료: %s\n", args[0])
	return nil
}

func runUserList(_ *cobra.Command, _ []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := ic.ListUsers(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "사용자 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "이메일", "활성"}
	var rows [][]string
	for _, u := range resp.Items {
		enabled := "N"
		if u.Enabled {
			enabled = "Y"
		}
		rows = append(rows, []string{u.ID, u.Name, u.Email, enabled})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runUserDelete(_ *cobra.Command, args []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := ic.DeleteUser(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "사용자 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("사용자 삭제 완료: %s\n", args[0])
	return nil
}

func runRoleAssign(_ *cobra.Command, _ []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := ic.AssignRole(roleUserID, roleProjectID, roleRoleID); err != nil {
		fmt.Fprintf(os.Stderr, "역할 부여 실패: %v\n", err)
		return err
	}
	fmt.Printf("역할 부여 완료: 사용자=%s, 프로젝트=%s, 역할=%s\n", roleUserID, roleProjectID, roleRoleID)
	return nil
}

func runRoleRevoke(_ *cobra.Command, _ []string) error {
	ic, err := newIdentityClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := ic.RevokeRole(roleUserID, roleProjectID, roleRoleID); err != nil {
		fmt.Fprintf(os.Stderr, "역할 회수 실패: %v\n", err)
		return err
	}
	fmt.Printf("역할 회수 완료: 사용자=%s, 프로젝트=%s, 역할=%s\n", roleUserID, roleProjectID, roleRoleID)
	return nil
}

func init() {
	// Role 플래그 등록
	roleAssignCmd.Flags().StringVar(&roleUserID, "user", "", "사용자 ID (필수)")
	roleAssignCmd.Flags().StringVar(&roleProjectID, "project", "", "프로젝트 ID (필수)")
	roleAssignCmd.Flags().StringVar(&roleRoleID, "role", "", "역할 ID (필수)")

	roleRevokeCmd.Flags().StringVar(&roleUserID, "user", "", "사용자 ID (필수)")
	roleRevokeCmd.Flags().StringVar(&roleProjectID, "project", "", "프로젝트 ID (필수)")
	roleRevokeCmd.Flags().StringVar(&roleRoleID, "role", "", "역할 ID (필수)")

	projectCmd.AddCommand(projectListCmd, projectCreateCmd, projectDeleteCmd)
	userCmd.AddCommand(userListCmd, userCreateCmd, userDeleteCmd)
	roleCmd.AddCommand(roleAssignCmd, roleRevokeCmd)

	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(roleCmd)
}
