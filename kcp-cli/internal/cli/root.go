package cli

import (
	"github.com/spf13/cobra"
)

var (
	// 글로벌 플래그
	cfgFile      string
	outputFormat string
)

// rootCmd 는 KCP CLI의 루트 커맨드이다
var rootCmd = &cobra.Command{
	Use:   "kcp",
	Short: "KCP CLI — OpenStack 인프라 통합 관리 도구",
	Long:  "KCP CLI는 Gateway를 통해 OpenStack 인프라를 관리하는 CLI 도구입니다.",
}

// Execute 는 루트 커맨드를 실행한다
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정 파일 경로 (기본: ~/.kcp/kcp-config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table|json|yaml)")
}
