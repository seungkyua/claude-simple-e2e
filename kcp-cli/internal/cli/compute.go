package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// --- VM 커맨드 ---

// vmCmd 은 VM 관련 상위 커맨드이다
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "VM(서버) 관리",
}

// vmListCmd 은 서버 목록을 조회하는 커맨드이다
var vmListCmd = &cobra.Command{
	Use:   "list",
	Short: "서버 목록 조회",
	RunE:  runVMList,
}

// vmShowCmd 은 서버 상세 정보를 조회하는 커맨드이다
var vmShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "서버 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMShow,
}

// vmCreateCmd 은 서버를 생성하는 커맨드이다
var vmCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "서버 생성",
	RunE:  runVMCreate,
}

// vmDeleteCmd 은 서버를 삭제하는 커맨드이다
var vmDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "서버 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMDelete,
}

// vmStartCmd 은 서버를 시작하는 커맨드이다
var vmStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "서버 시작",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("start"),
}

// vmStopCmd 은 서버를 정지하는 커맨드이다
var vmStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "서버 정지",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("stop"),
}

// vmRebootCmd 은 서버를 재부팅하는 커맨드이다
var vmRebootCmd = &cobra.Command{
	Use:   "reboot <id>",
	Short: "서버 재부팅",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("reboot"),
}

// --- Flavor 커맨드 ---

// flavorCmd 은 Flavor 관련 상위 커맨드이다
var flavorCmd = &cobra.Command{
	Use:   "flavor",
	Short: "Flavor(VM 사양) 관리",
}

// flavorListCmd 은 Flavor 목록을 조회하는 커맨드이다
var flavorListCmd = &cobra.Command{
	Use:   "list",
	Short: "Flavor 목록 조회",
	RunE:  runFlavorList,
}

// flavorCreateCmd 은 Flavor를 생성하는 커맨드이다
var flavorCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Flavor 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: Flavor 생성 폼 구현")
		return nil
	},
}

// flavorDeleteCmd 은 Flavor를 삭제하는 커맨드이다
var flavorDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Flavor 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runFlavorDelete,
}

// newComputeClient 는 설정 파일을 로드하여 ComputeClient를 생성한다
func newComputeClient() (sdk.ComputeClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewComputeClient(client), nil
}

func runVMList(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := cc.ListServers(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "생성일"}
	var rows [][]string
	for _, s := range resp.Items {
		rows = append(rows, []string{s.ID, s.Name, s.Status, s.Created.Format("2006-01-02 15:04")})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runVMShow(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	s, err := cc.GetServer(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "Flavor", "생성일"}
	rows := [][]string{{s.ID, s.Name, s.Status, s.Flavor.Name, s.Created.Format("2006-01-02 15:04")}}
	formatOutput(outputFormat, headers, rows, s)
	return nil
}

// vmCreateName, vmCreateFlavor, vmCreateImage 는 vm create 플래그이다
var (
	vmCreateName   string
	vmCreateFlavor string
	vmCreateImage  string
)

func runVMCreate(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	req := &sdk.CreateServerRequest{
		Name:     vmCreateName,
		FlavorID: vmCreateFlavor,
		ImageID:  vmCreateImage,
	}
	s, err := cc.CreateServer(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 생성 실패: %v\n", err)
		return err
	}
	fmt.Printf("서버 생성 완료: %s (%s)\n", s.Name, s.ID)
	return nil
}

func runVMDelete(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := cc.DeleteServer(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "서버 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("서버 삭제 완료: %s\n", args[0])
	return nil
}

// runVMAction 은 서버 액션(시작/정지/재부팅) 커맨드 핸들러를 반환한다
func runVMAction(action string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		cc, err := newComputeClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			return err
		}
		if err := cc.ServerAction(args[0], action); err != nil {
			fmt.Fprintf(os.Stderr, "서버 %s 실패: %v\n", action, err)
			return err
		}
		fmt.Printf("서버 %s 완료: %s\n", action, args[0])
		return nil
	}
}

func runFlavorList(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	flavors, err := cc.ListFlavors()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Flavor 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "vCPU", "RAM(MB)", "Disk(GB)"}
	var rows [][]string
	for _, f := range flavors {
		rows = append(rows, []string{f.ID, f.Name, fmt.Sprintf("%d", f.VCPUs), fmt.Sprintf("%d", f.RAM), fmt.Sprintf("%d", f.Disk)})
	}
	formatOutput(outputFormat, headers, rows, flavors)
	return nil
}

func runFlavorDelete(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := cc.DeleteFlavor(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Flavor 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("Flavor 삭제 완료: %s\n", args[0])
	return nil
}

func init() {
	// VM 서브커맨드 등록
	vmCreateCmd.Flags().StringVar(&vmCreateName, "name", "", "서버 이름 (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateFlavor, "flavor", "", "Flavor ID (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateImage, "image", "", "이미지 ID (필수)")

	vmCmd.AddCommand(vmListCmd, vmShowCmd, vmCreateCmd, vmDeleteCmd, vmStartCmd, vmStopCmd, vmRebootCmd)
	flavorCmd.AddCommand(flavorListCmd, flavorCreateCmd, flavorDeleteCmd)

	rootCmd.AddCommand(vmCmd)
	rootCmd.AddCommand(flavorCmd)
}
