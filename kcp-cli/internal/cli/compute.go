package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// --- Server 커맨드 ---

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"vm"},
	Short:   "서버(인스턴스) 관리",
}

var vmListCmd = &cobra.Command{
	Use:   "list",
	Short: "서버 목록 조회",
	RunE:  runVMList,
}

var vmShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "서버 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMShow,
}

var vmCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "서버 생성",
	RunE:  runVMCreate,
}

var vmDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "서버 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMDelete,
}

var vmStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "서버 시작",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("start"),
}

var vmStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "서버 정지",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("stop"),
}

var vmRebootCmd = &cobra.Command{
	Use:   "reboot <id>",
	Short: "서버 재부팅",
	Args:  cobra.ExactArgs(1),
	RunE:  runVMAction("reboot"),
}

// --- Flavor 커맨드 ---

var flavorCmd = &cobra.Command{
	Use:   "flavor",
	Short: "Flavor(VM 사양) 관리",
}

var flavorListCmd = &cobra.Command{
	Use:   "list",
	Short: "Flavor 목록 조회",
	RunE:  runFlavorList,
}

var flavorCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Flavor 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: Flavor 생성 폼 구현")
		return nil
	},
}

var flavorDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Flavor 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runFlavorDelete,
}

func newComputeClient() (sdk.ComputeClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewComputeClient(client), nil
}

// openstack server list 동일 출력:
// ID | Name | Status | Networks | Image | Flavor
// Gateway에서 flavor_name, image_name, networks를 enrichment하여 전달
func runVMList(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
		return err
	}
	resp, err := cc.ListServers(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 목록 조회 실패: %v\n", err)
		return err
	}

	headers := []string{"ID", "Name", "Status", "Networks", "Image", "Flavor"}
	var rows [][]string
	for _, s := range resp.Items {
		// Gateway가 enrichment한 flavor.name, image.name, networks 사용
		flavorName := s.Flavor.Name
		if flavorName == "" {
			flavorName = s.Flavor.ID
		}
		imageName := s.Image.Name
		if imageName == "" {
			imageName = s.Image.ID
		}
		networks := s.Networks
		if networks == "" {
			networks = s.FormatNetworks()
		}
		rows = append(rows, []string{
			s.ID, s.Name, s.Status, networks, imageName, flavorName,
		})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

// openstack server show 동일 출력
func runVMShow(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
		return err
	}
	s, err := cc.GetServer(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 조회 실패: %v\n", err)
		return err
	}

	flavorName := s.Flavor.Name
	if flavorName == "" {
		flavorName = s.Flavor.ID
	}
	imageName := s.Image.Name
	if imageName == "" {
		imageName = s.Image.ID
	}
	networks := s.Networks
	if networks == "" {
		networks = s.FormatNetworks()
	}

	headers := []string{"ID", "Name", "Status", "Networks", "Image", "Flavor"}
	rows := [][]string{{s.ID, s.Name, s.Status, networks, imageName, flavorName}}
	formatOutput(outputFormat, headers, rows, s)
	return nil
}

var (
	vmCreateName   string
	vmCreateFlavor string
	vmCreateImage  string
)

func runVMCreate(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
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
		return err
	}
	if err := cc.DeleteServer(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "서버 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("서버 삭제 완료: %s\n", args[0])
	return nil
}

func runVMAction(action string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		cc, err := newComputeClient()
		if err != nil {
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

// openstack flavor list 동일 출력:
// ID | Name | RAM | Disk | Ephemeral | VCPUs | Is Public
func runFlavorList(_ *cobra.Command, _ []string) error {
	cc, err := newComputeClient()
	if err != nil {
		return err
	}
	flavors, err := cc.ListFlavors()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Flavor 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "Name", "RAM", "Disk", "VCPUs", "Is Public"}
	var rows [][]string
	for _, f := range flavors {
		isPublic := "True"
		if !f.IsPublic {
			isPublic = "False"
		}
		rows = append(rows, []string{
			f.ID, f.Name,
			fmt.Sprintf("%d", f.RAM),
			fmt.Sprintf("%d", f.Disk),
			fmt.Sprintf("%d", f.VCPUs),
			isPublic,
		})
	}
	formatOutput(outputFormat, headers, rows, flavors)
	return nil
}

func runFlavorDelete(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
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
	vmCreateCmd.Flags().StringVar(&vmCreateName, "name", "", "서버 이름 (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateFlavor, "flavor", "", "Flavor ID (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateImage, "image", "", "이미지 ID (필수)")

	serverCmd.AddCommand(vmListCmd, vmShowCmd, vmCreateCmd, vmDeleteCmd, vmStartCmd, vmStopCmd, vmRebootCmd)
	flavorCmd.AddCommand(flavorListCmd, flavorCreateCmd, flavorDeleteCmd)

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(flavorCmd)
}
