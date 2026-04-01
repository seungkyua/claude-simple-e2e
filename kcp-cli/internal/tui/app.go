package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// App 은 TUI 애플리케이션의 메인 모델이다
type App struct {
	currentView string
	width       int
	height      int
}

// NewApp 은 새로운 TUI 앱을 생성한다
func NewApp() App {
	return App{
		currentView: "dashboard",
	}
}

// Init 은 초기 커맨드를 반환한다
func (a App) Init() tea.Cmd {
	return nil
}

// Update 는 메시지를 처리하고 모델을 업데이트한다
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}
	return a, nil
}

// View 는 현재 UI를 렌더링한다
func (a App) View() string {
	return "KCP TUI — OpenStack 인프라 관리\n\n'q' 를 눌러 종료합니다."
}
