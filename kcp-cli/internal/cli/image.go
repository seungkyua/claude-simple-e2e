package cli

import (
	"fmt"
	"os"

	"github.com/kcp-cli/kcp-cli/internal/config"
	"github.com/kcp-cli/kcp-cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// imageCmd 은 이미지 관련 상위 커맨드이다
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "이미지 관리",
}

var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "이미지 목록 조회",
	RunE:  runImageList,
}

var imageShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "이미지 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  runImageShow,
}

var imageDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "이미지 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  runImageDelete,
}

// newImageClient 는 설정 파일을 로드하여 ImageClient를 생성한다
func newImageClient() (sdk.ImageClient, error) {
	cfgPath := config.ResolvePath(cfgFile)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %w", err)
	}
	client := sdk.NewClient(cfg.ServerURL, sdk.WithToken(cfg.Token))
	return sdk.NewImageClient(client), nil
}

func runImageList(_ *cobra.Command, _ []string) error {
	ic, err := newImageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	resp, err := ic.ListImages(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "이미지 목록 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "디스크형식", "크기(MB)"}
	var rows [][]string
	for _, img := range resp.Items {
		sizeMB := img.Size / (1024 * 1024)
		rows = append(rows, []string{img.ID, img.Name, img.Status, img.DiskFormat, fmt.Sprintf("%d", sizeMB)})
	}
	formatOutput(outputFormat, headers, rows, resp.Items)
	return nil
}

func runImageShow(_ *cobra.Command, args []string) error {
	ic, err := newImageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	img, err := ic.GetImage(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "이미지 조회 실패: %v\n", err)
		return err
	}
	headers := []string{"ID", "이름", "상태", "디스크형식", "가시성", "최소디스크(GB)", "최소RAM(MB)"}
	rows := [][]string{{
		img.ID, img.Name, img.Status, img.DiskFormat, img.Visibility,
		fmt.Sprintf("%d", img.MinDisk), fmt.Sprintf("%d", img.MinRAM),
	}}
	formatOutput(outputFormat, headers, rows, img)
	return nil
}

func runImageDelete(_ *cobra.Command, args []string) error {
	ic, err := newImageClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		return err
	}
	if err := ic.DeleteImage(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "이미지 삭제 실패: %v\n", err)
		return err
	}
	fmt.Printf("이미지 삭제 완료: %s\n", args[0])
	return nil
}

func init() {
	imageCmd.AddCommand(imageListCmd, imageShowCmd, imageDeleteCmd)
	rootCmd.AddCommand(imageCmd)
}
