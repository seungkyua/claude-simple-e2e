package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// --- Volume 커맨드 ---

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "볼륨 관리",
}

var volumeListCmd = &cobra.Command{
	Use:   "list",
	Short: "볼륨 목록 조회",
	RunE:  runVolumeList,
}

var volumeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "볼륨 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 볼륨 생성 폼 구현")
		return nil
	},
}

var volumeDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "볼륨 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeDelete,
}

// volumeAttachServerID 는 볼륨 연결 시 대상 서버 ID이다
var volumeAttachServerID string

var volumeAttachCmd = &cobra.Command{
	Use:   "attach <volume-id>",
	Short: "볼륨을 서버에 연결",
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeAttach,
}

var volumeDetachCmd = &cobra.Command{
	Use:   "detach <volume-id>",
	Short: "볼륨을 서버에서 분리",
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeDetach,
}

// --- Snapshot 커맨드 ---

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "스냅샷 관리",
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "스냅샷 목록 조회",
	RunE:  runSnapshotList,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "스냅샷 생성",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("TODO: 스냅샷 생성 폼 구현")
		return nil
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "스냅샷 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotDelete,
}

// newStorageClient 는 설정 파일을 로드하여 StorageClient를 생성한다
func newStorageClient() (sdk.StorageClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewStorageClient(client), nil
}

func runVolumeList(_ *cobra.Command, _ []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := sc.ListVolumes(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "볼륨 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "크기(GB)", "타입"}
	var rows [][]string
	for _, v := range resp.Items {
		rows = append(rows, []string{v.ID, v.Name, v.Status, fmt.Sprintf("%d", v.Size), v.VolumeType})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runVolumeDelete(_ *cobra.Command, args []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := sc.DeleteVolume(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "볼륨 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("볼륨 삭제 완료: %s\n", args[0])
	return nil
}

func runVolumeAttach(_ *cobra.Command, args []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := sc.AttachVolume(args[0], volumeAttachServerID); err != nil {
		fmt.Fprintf(os.Stderr, "볼륨 연결 실패: %v\n", err)
		return err
	}
	fmt.Printf("볼륨 연결 완료: %s -> 서버 %s\n", args[0], volumeAttachServerID)
	return nil
}

func runVolumeDetach(_ *cobra.Command, args []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := sc.DetachVolume(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "볼륨 분리 실패: %v\n", err)
		return err
	}
	fmt.Printf("볼륨 분리 완료: %s\n", args[0])
	return nil
}

func runSnapshotList(_ *cobra.Command, _ []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := sc.ListSnapshots(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "스냅샷 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "볼륨ID", "크기(GB)"}
	var rows [][]string
	for _, s := range resp.Items {
		rows = append(rows, []string{s.ID, s.Name, s.Status, s.VolumeID, fmt.Sprintf("%d", s.Size)})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runSnapshotDelete(_ *cobra.Command, args []string) error {
	sc, err := newStorageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := sc.DeleteSnapshot(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "스냅샷 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("스냅샷 삭제 완료: %s\n", args[0])
	return nil
}

func init() {
	// 볼륨 attach 에 --server 플래그 등록
	volumeAttachCmd.Flags().StringVar(&volumeAttachServerID, "server", "", "연결 대상 서버 ID (필수)")

	volumeCmd.AddCommand(volumeListCmd, volumeCreateCmd, volumeDeleteCmd, volumeAttachCmd, volumeDetachCmd)
	snapshotCmd.AddCommand(snapshotListCmd, snapshotCreateCmd, snapshotDeleteCmd)

	rootCmd.AddCommand(volumeCmd)
	rootCmd.AddCommand(snapshotCmd)
}
