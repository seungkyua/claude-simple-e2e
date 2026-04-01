package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// --- Network 커맨드 ---

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "네트워크 관리",
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "네트워크 목록 조회",
	RunE:  runNetworkList,
}

var networkCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "네트워크 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 네트워크 생성 폼 구현")
		return nil
	},
}

var networkDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "네트워크 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runNetworkDelete,
}

// --- Subnet 커맨드 ---

var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "서브넷 관리",
}

var subnetListCmd = &cobra.Command{
	Use:   "list",
	Short: "서브넷 목록 조회",
	RunE:  runSubnetList,
}

var subnetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "서브넷 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 서브넷 생성 폼 구현")
		return nil
	},
}

var subnetDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "서브넷 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runSubnetDelete,
}

// --- Router 커맨드 ---

var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "라우터 관리",
}

var routerListCmd = &cobra.Command{
	Use:   "list",
	Short: "라우터 목록 조회",
	RunE:  runRouterList,
}

var routerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "라우터 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 라우터 생성 폼 구현")
		return nil
	},
}

var routerDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "라우터 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runRouterDelete,
}

// --- Security Group 커맨드 ---

var secgroupCmd = &cobra.Command{
	Use:   "secgroup",
	Short: "보안그룹 관리",
}

var secgroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "보안그룹 목록 조회",
	RunE:  runSecgroupList,
}

var secgroupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "보안그룹 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 보안그룹 생성 폼 구현")
		return nil
	},
}

var secgroupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "보안그룹 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecgroupDelete,
}

// newNetworkClient 는 설정 파일을 로드하여 NetworkClient를 생성한다
func newNetworkClient() (sdk.NetworkClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewNetworkClient(client), nil
}

func runNetworkList(_ *cobra.Command, _ []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := nc.ListNetworks(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "네트워크 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "공유"}
	var rows [][]string
	for _, n := range resp.Items {
		shared := "N"
		if n.Shared {
			shared = "Y"
		}
		rows = append(rows, []string{n.ID, n.Name, n.Status, shared})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runNetworkDelete(_ *cobra.Command, args []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := nc.DeleteNetwork(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "네트워크 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("네트워크 삭제 완료: %s\n", args[0])
	return nil
}

func runSubnetList(_ *cobra.Command, _ []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := nc.ListSubnets(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "서브넷 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "CIDR", "게이트웨이"}
	var rows [][]string
	for _, s := range resp.Items {
		rows = append(rows, []string{s.ID, s.Name, s.CIDR, s.GatewayIP})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runSubnetDelete(_ *cobra.Command, args []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := nc.DeleteSubnet(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "서브넷 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("서브넷 삭제 완료: %s\n", args[0])
	return nil
}

func runRouterList(_ *cobra.Command, _ []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := nc.ListRouters(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "라우터 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태"}
	var rows [][]string
	for _, r := range resp.Items {
		rows = append(rows, []string{r.ID, r.Name, r.Status})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runRouterDelete(_ *cobra.Command, args []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := nc.DeleteRouter(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "라우터 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("라우터 삭제 완료: %s\n", args[0])
	return nil
}

func runSecgroupList(_ *cobra.Command, _ []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := nc.ListSecurityGroups(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "보안그룹 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "규칙 수"}
	var rows [][]string
	for _, sg := range resp.Items {
		rows = append(rows, []string{sg.ID, sg.Name, fmt.Sprintf("%d", len(sg.Rules))})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runSecgroupDelete(_ *cobra.Command, args []string) error {
	nc, err := newNetworkClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := nc.DeleteSecurityGroup(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "보안그룹 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("보안그룹 삭제 완료: %s\n", args[0])
	return nil
}

func init() {
	networkCmd.AddCommand(networkListCmd, networkCreateCmd, networkDeleteCmd)
	subnetCmd.AddCommand(subnetListCmd, subnetCreateCmd, subnetDeleteCmd)
	routerCmd.AddCommand(routerListCmd, routerCreateCmd, routerDeleteCmd)
	secgroupCmd.AddCommand(secgroupListCmd, secgroupCreateCmd, secgroupDeleteCmd)

	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(subnetCmd)
	rootCmd.AddCommand(routerCmd)
	rootCmd.AddCommand(secgroupCmd)
}
