package main

import (
	"context"
	"time"

	"github.com/BenedictKing/ccx/desktop/internal/backend"
	"github.com/pkg/browser"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DesktopService struct {
	manager    *backend.Manager
	app        *application.App
	mainWindow application.Window
}

func NewDesktopService(manager *backend.Manager) *DesktopService {
	return &DesktopService{manager: manager}
}

func (s *DesktopService) setApp(app *application.App) {
	s.app = app
}

func (s *DesktopService) setMainWindow(window application.Window) {
	s.mainWindow = window
}

func (s *DesktopService) GetStatus() backend.Status {
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()
	return s.manager.Status(ctx)
}

func (s *DesktopService) StartService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return s.manager.Start(ctx)
}

func (s *DesktopService) StopService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	return s.manager.Stop(ctx)
}

func (s *DesktopService) RestartService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.manager.Restart(ctx)
}

func (s *DesktopService) GetLogs() []string {
	return s.manager.Logs()
}

func (s *DesktopService) ShowStatusTab() error {
	s.showWindow()
	if s.app != nil {
		s.app.Event.Emit("desktop:show-tab", "status")
	}
	return nil
}

func (s *DesktopService) ShowWebUITab() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := s.manager.Start(ctx); err != nil {
		return err
	}
	if err := s.manager.WaitHealthy(ctx, 15*time.Second); err != nil {
		return err
	}
	s.showWindow()
	if s.app != nil {
		s.app.Event.Emit("desktop:show-tab", "web")
	}
	return nil
}

func (s *DesktopService) OpenWebUIInBrowser() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := s.manager.Start(ctx); err != nil {
		return err
	}
	if err := s.manager.WaitHealthy(ctx, 15*time.Second); err != nil {
		return err
	}
	return browser.OpenURL(s.manager.WebURL())
}

func (s *DesktopService) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = s.manager.Stop(ctx)
}

func (s *DesktopService) showWindow() {
	if s.mainWindow == nil {
		return
	}
	if s.mainWindow.IsMinimised() {
		s.mainWindow.UnMinimise()
	}
	s.mainWindow.Show()
	s.mainWindow.Focus()
}
