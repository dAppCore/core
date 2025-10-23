package core

import (
	"embed"
	"fmt"
	"sync"

	"core/config"
	"core/crypt"
	"core/display"
	"core/docs"
	"core/filesystem"
	"core/filesystem/local"
	"core/workspace"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service provides access to all core application services.
type Service struct {
	app              *application.App
	configService    *config.Service
	displayService   *display.Service
	docsService      *docs.Service
	cryptService     *crypt.Service
	workspaceService *workspace.Service
}

var (
	instance *Service
	once     sync.Once
	initErr  error
)

// New performs Phase 1 of initialization: Instantiation.
// It creates the raw service objects without wiring them together.
func New(assets embed.FS) *Service {
	once.Do(func() {
		// Instantiate services in the correct order of dependency.
		configService, err := config.NewService()
		if err != nil {
			initErr = fmt.Errorf("failed to initialize config service: %w", err)
			return
		}

		// Initialize the local filesystem medium
		filesystem.Local, err = local.New(configService.Get().RootDir)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize local filesystem: %w", err)
			return
		}

		displayService := display.NewService(display.ClientHub, assets)
		docsService := docs.NewService(assets)
		cryptService := crypt.NewService(configService.Get())
		workspaceService := workspace.NewService(configService.Get(), workspace.NewLocalMedium())

		instance = &Service{
			configService:    configService,
			displayService:   displayService,
			docsService:      docsService,
			cryptService:     cryptService,
			workspaceService: workspaceService,
		}
	})

	if initErr != nil {
		panic(initErr) // A failure in a core service is fatal.
	}

	return instance
}

// Setup performs Phase 2 of initialization: Wiring.
// It injects the required dependencies into each service.
func Setup(app *application.App) {
	if instance == nil {
		panic("core.Setup() called before core.New() was successfully initialized")
	}
	instance.app = app

	// Wire the services with their dependencies.
	instance.displayService.Setup(app, instance.configService, nil)
	instance.docsService.Setup(app, instance.displayService)
}

// App returns the global application instance.
func App() *application.App {
	if instance == nil || instance.app == nil {
		panic("core.App() called before core.Setup() was successfully initialized")
	}
	return instance.app
}

// Config returns the singleton instance of the ConfigService.
func Config() *config.Service {
	if instance == nil {
		panic("core.Config() called before core.New() was successfully initialized")
	}
	return instance.configService
}

// Display returns the singleton instance of the display.Service.
func Display() *display.Service {
	if instance == nil {
		panic("core.Display() called before core.New() was successfully initialized")
	}
	return instance.displayService
}

// Docs returns the singleton instance of the DocsService.
func Docs() *docs.Service {
	if instance == nil {
		panic("core.Docs() called before core.New() was successfully initialized")
	}
	return instance.docsService
}

// Crypt returns the singleton instance of the CryptService.
func Crypt() *crypt.Service {
	if instance == nil {
		panic("core.Crypt() called before core.New() was successfully initialized")
	}
	return instance.cryptService
}

// Filesystem returns the singleton instance of the FilesystemService.
func Filesystem() filesystem.Medium {
	if instance == nil {
		panic("core.Filesystem() called before core.New() was successfully initialized")
	}
	return filesystem.Local
}

// Workspace returns the singleton instance of the WorkspaceService.
func Workspace() *workspace.Service {
	if instance == nil {
		panic("core.Workspace() called before core.New() was successfully initialized")
	}
	return instance.workspaceService
}

// Runtime returns the singleton instance of the Service.
func Runtime() *Service {
	if instance == nil {
		panic("core.Runtime() called before core.New() was successfully initialized")
	}
	return instance
}
