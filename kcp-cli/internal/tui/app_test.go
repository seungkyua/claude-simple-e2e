package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewApp 은 새로운 App 생성 시 초기 상태를 검증한다
func TestNewApp(t *testing.T) {
	app := NewApp()

	// 초기 뷰가 "dashboard"인지 확인한다
	if app.currentView != "dashboard" {
		t.Errorf("초기 currentView가 'dashboard'이어야 하지만 '%s'이다", app.currentView)
	}

	// 초기 크기가 0인지 확인한다
	if app.width != 0 {
		t.Errorf("초기 width가 0이어야 하지만 %d이다", app.width)
	}
	if app.height != 0 {
		t.Errorf("초기 height가 0이어야 하지만 %d이다", app.height)
	}

	// Init()이 nil을 반환해야 한다
	if cmd := app.Init(); cmd != nil {
		t.Error("Init()이 nil을 반환해야 한다")
	}
}

// TestAppQuit 은 'q' 키 입력 시 tea.Quit 커맨드가 반환되는지 검증한다
func TestAppQuit(t *testing.T) {
	app := NewApp()

	// 'q' 키 메시지를 전달한다
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("'q' 입력 시 tea.Quit 커맨드가 반환되어야 한다")
	}

	// tea.Quit 은 실행 시 QuitMsg를 반환한다
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("반환된 커맨드가 QuitMsg를 생성해야 하지만 %T이다", msg)
	}
}

// TestWindowResize 는 WindowSizeMsg 수신 시 width/height가 업데이트되는지 검증한다
func TestWindowResize(t *testing.T) {
	app := NewApp()

	// 윈도우 크기 변경 메시지를 전달한다
	model, cmd := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// 커맨드가 nil이어야 한다 (종료 아님)
	if cmd != nil {
		t.Error("WindowSizeMsg 처리 후 커맨드가 nil이어야 한다")
	}

	// 업데이트된 모델의 크기를 확인한다
	updated, ok := model.(App)
	if !ok {
		t.Fatal("반환된 모델이 App 타입이어야 한다")
	}

	if updated.width != 120 {
		t.Errorf("width가 120이어야 하지만 %d이다", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height가 40이어야 하지만 %d이다", updated.height)
	}
}

// TestAppView 는 View()가 비어있지 않은 문자열을 반환하는지 검증한다
func TestAppView(t *testing.T) {
	app := NewApp()
	view := app.View()

	if view == "" {
		t.Error("View()가 비어있지 않은 문자열을 반환해야 한다")
	}
}
