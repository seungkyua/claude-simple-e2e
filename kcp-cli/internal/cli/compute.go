package cli

import (
	"fmt"
	"os"
	"strings"

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
	Use:   "create <name>",
	Short: "서버 생성",
	Long: `서버를 생성한다. OpenStack CLI와 동일한 인자를 지원한다.
예: kcp server create --flavor m1.tiny --image <image-id> --network private --key-name mykey --security-group <sg-id> my-server`,
	Args: cobra.ExactArgs(1),
	RunE: runVMCreate,
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

// serverDetailFields 는 Server 객체를 OpenStack CLI server show 형식의 Field/Value 배열로 변환한다
func serverDetailFields(s *sdk.Server) [][]string {
	flavorName := s.Flavor.Name
	if flavorName == "" {
		flavorName = s.Flavor.ID
	}
	// flavor 표시: "m1.tiny (1)" 형식
	flavorDisplay := flavorName
	if s.Flavor.ID != "" && s.Flavor.ID != flavorName {
		flavorDisplay = fmt.Sprintf("%s (%s)", flavorName, s.Flavor.ID)
	}

	imageName := s.Image.Name
	if imageName == "" {
		imageName = s.Image.ID
	}
	imageDisplay := ""
	if s.Image.ID != "" {
		imageDisplay = fmt.Sprintf("%s (%s)", imageName, s.Image.ID)
	}

	networks := s.Networks
	if networks == "" {
		networks = s.FormatNetworks()
	}

	var sgNames []string
	for _, sg := range s.SecurityGroups {
		sgNames = append(sgNames, fmt.Sprintf("name='%s'", sg.Name))
	}

	var volIDs []string
	for _, v := range s.VolumesAttached {
		volIDs = append(volIDs, v.ID)
	}

	powerState := "NOSTATE"
	switch s.PowerState {
	case 1:
		powerState = "Running"
	case 3:
		powerState = "Paused"
	case 4:
		powerState = "Shutdown"
	case 6:
		powerState = "Crashed"
	case 7:
		powerState = "Suspended"
	}

	fields := [][]string{
		{"OS-DCF:diskConfig", s.DiskConfig},
		{"OS-EXT-AZ:availability_zone", s.AvailabilityZone},
		{"OS-EXT-SRV-ATTR:host", noneIfEmpty(s.Host)},
		{"OS-EXT-SRV-ATTR:hypervisor_hostname", noneIfEmpty(s.HypervisorHostname)},
		{"OS-EXT-SRV-ATTR:instance_name", s.InstanceName},
		{"OS-EXT-STS:power_state", powerState},
		{"OS-EXT-STS:task_state", noneIfEmpty(s.TaskState)},
		{"OS-EXT-STS:vm_state", s.VMState},
		{"OS-SRV-USG:launched_at", noneIfEmpty(s.LaunchedAt)},
		{"OS-SRV-USG:terminated_at", noneIfEmpty(s.TerminatedAt)},
		{"accessIPv4", s.AccessIPv4},
		{"accessIPv6", s.AccessIPv6},
		{"addresses", networks},
	}

	// adminPass는 서버 생성 직후에만 표시
	if s.AdminPass != "" {
		fields = append(fields, []string{"adminPass", s.AdminPass})
	}

	fields = append(fields, [][]string{
		{"config_drive", s.ConfigDrive},
		{"created", s.Created.Format("2006-01-02T15:04:05Z")},
		{"description", noneIfEmpty(s.Description)},
		{"flavor", flavorDisplay},
		{"hostId", s.HostID},
		{"id", s.ID},
		{"image", imageDisplay},
		{"key_name", s.KeyName},
		{"locked", fmt.Sprintf("%v", s.Locked)},
		{"name", s.Name},
		{"progress", fmt.Sprintf("%d", s.Progress)},
		{"project_id", s.ProjectID},
		{"security_groups", strings.Join(sgNames, ", ")},
		{"status", s.Status},
		{"updated", s.Updated.Format("2006-01-02T15:04:05Z")},
		{"user_id", s.UserID},
		{"volumes_attached", strings.Join(volIDs, ", ")},
	}...)

	return fields
}

// openstack server show 동일 출력 (Field/Value 세로 테이블)
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

	formatDetailOutput(outputFormat, serverDetailFields(s), s)
	return nil
}

var (
	vmCreateFlavor         string
	vmCreateImage          string
	vmCreateNetwork        string
	vmCreateKeyName        string
	vmCreateSecurityGroup  string
)

// kcp server create --flavor m1.tiny --image <id> --network private --key-name mykey --security-group <sg-id> <name>
func runVMCreate(_ *cobra.Command, args []string) error {
	cc, err := newComputeClient()
	if err != nil {
		return err
	}

	req := &sdk.CreateServerRequest{
		Name:     args[0],
		FlavorID: vmCreateFlavor,
		ImageID:  vmCreateImage,
	}
	if vmCreateNetwork != "" {
		req.NetworkIDs = []string{vmCreateNetwork}
	}
	if vmCreateKeyName != "" {
		req.KeyName = vmCreateKeyName
	}
	if vmCreateSecurityGroup != "" {
		req.SecurityGroupIDs = []string{vmCreateSecurityGroup}
	}

	s, err := cc.CreateServer(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "서버 생성 실패: %v\n", err)
		return err
	}

	// 생성 후 상세 정보를 Field/Value 형식으로 출력 (openstack server create 동일)
	// Gateway가 생성 직후 상세 조회 + enrichment + adminPass를 반환한다
	formatDetailOutput(outputFormat, serverDetailFields(s), s)
	return nil
}

// noneIfEmpty 는 빈 문자열이면 "None"을 반환한다
func noneIfEmpty(s string) string {
	if s == "" {
		return "None"
	}
	return s
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
	vmCreateCmd.Flags().StringVar(&vmCreateFlavor, "flavor", "", "Flavor 이름 또는 ID (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateImage, "image", "", "이미지 이름 또는 ID (필수)")
	vmCreateCmd.Flags().StringVar(&vmCreateNetwork, "network", "", "네트워크 이름 또는 ID")
	vmCreateCmd.Flags().StringVar(&vmCreateKeyName, "key-name", "", "SSH 키 이름")
	vmCreateCmd.Flags().StringVar(&vmCreateSecurityGroup, "security-group", "", "보안그룹 이름 또는 ID")

	serverCmd.AddCommand(vmListCmd, vmShowCmd, vmCreateCmd, vmDeleteCmd, vmStartCmd, vmStopCmd, vmRebootCmd)
	flavorCmd.AddCommand(flavorListCmd, flavorCreateCmd, flavorDeleteCmd)

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(flavorCmd)
}
