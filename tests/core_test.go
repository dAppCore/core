package core_test

import (
	. "forge.lthn.ai/core/go/pkg/core"
	"context"
	"embed"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockApp is a simple mock for testing app injection
type mockApp struct{}

func TestCore_New_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// Mock service for testing
type MockService struct {
	Name string
}

func (m *MockService) GetName() string {
	return m.Name
}

func TestCore_WithService_Good(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return &MockService{Name: "test"}, nil
	}
	c, err := New(WithService(factory))
	assert.NoError(t, err)
	svc := c.Service().Get("core")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_WithService_Bad(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return nil, assert.AnError
	}
	_, err := New(WithService(factory))
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

type MockConfigService struct{}

func (m *MockConfigService) Get(key string, out any) error { return nil }
func (m *MockConfigService) Set(key string, v any) error   { return nil }

func TestCore_Services_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	err = c.RegisterService("config", &MockConfigService{})
	assert.NoError(t, err)

	svc := c.Service("config")
	assert.NotNil(t, svc)

	// Cfg() returns Cfg (always available, not a service)
	cfg := c.Config()
	assert.NotNil(t, cfg)
}

func TestCore_App_Good(t *testing.T) {
	app := &mockApp{}
	c, err := New(WithApp(app))
	assert.NoError(t, err)

	// To test the global CoreGUI() function, we need to set the global instance.
	originalInstance := GetInstance()
	SetInstance(c)
	defer SetInstance(originalInstance)

	assert.Equal(t, app, CoreGUI())
}

func TestCore_App_Ugly(t *testing.T) {
	// This test ensures that calling CoreGUI() before the core is initialized panics.
	originalInstance := GetInstance()
	ClearInstance()
	defer SetInstance(originalInstance)
	assert.Panics(t, func() {
		CoreGUI()
	})
}

func TestCore_Core_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.Equal(t, c, c.Core())
}

func TestEtc_Features_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	c.Config().Enable("feature1")
	c.Config().Enable("feature2")

	assert.True(t, c.Config().Enabled("feature1"))
	assert.True(t, c.Config().Enabled("feature2"))
	assert.False(t, c.Config().Enabled("feature3"))
	assert.False(t, c.Config().Enabled(""))
}

func TestEtc_Settings_Good(t *testing.T) {
	c, _ := New()
	c.Config().Set("api_url", "https://api.lthn.sh")
	c.Config().Set("max_agents", 5)

	assert.Equal(t, "https://api.lthn.sh", c.Config().GetString("api_url"))
	assert.Equal(t, 5, c.Config().GetInt("max_agents"))
	assert.Equal(t, "", c.Config().GetString("missing"))
}

func TestEtc_Features_Edge(t *testing.T) {
	c, _ := New()
	c.Config().Enable("foo")
	assert.True(t, c.Config().Enabled("foo"))
	assert.False(t, c.Config().Enabled("FOO")) // Case sensitive

	c.Config().Disable("foo")
	assert.False(t, c.Config().Enabled("foo"))
}

func TestCore_ServiceLifecycle_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	var messageReceived Message
	handler := func(c *Core, msg Message) error {
		messageReceived = msg
		return nil
	}
	c.RegisterAction(handler)

	// Test Startup
	_ = c.ServiceStartup(context.TODO(), nil)
	_, ok := messageReceived.(ActionServiceStartup)
	assert.True(t, ok, "expected ActionServiceStartup message")

	// Test Shutdown
	_ = c.ServiceShutdown(context.TODO())
	_, ok = messageReceived.(ActionServiceShutdown)
	assert.True(t, ok, "expected ActionServiceShutdown message")
}

func TestCore_WithApp_Good(t *testing.T) {
	app := &mockApp{}
	c, err := New(WithApp(app))
	assert.NoError(t, err)
	assert.Equal(t, app, c.App().Runtime)
}

//go:embed testdata
var testFS embed.FS

func TestCore_WithAssets_Good(t *testing.T) {
	c, err := New(WithAssets(testFS))
	assert.NoError(t, err)
	file, err := c.Embed().Open("testdata/test.txt")
	assert.NoError(t, err)
	defer func() { _ = file.Close() }()
	content, err := io.ReadAll(file)
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(content))
}

func TestCore_WithServiceLock_Good(t *testing.T) {
	c, err := New(WithServiceLock())
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.Error(t, err)
}

func TestCore_RegisterService_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc := c.Service("test")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_RegisterService_Bad(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.Error(t, err)
	err = c.RegisterService("", &MockService{})
	assert.Error(t, err)
}

func TestCore_ServiceFor_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc, err := ServiceFor[*MockService](c, "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", svc.GetName())
}

func TestCore_ServiceFor_Bad(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	_, err = ServiceFor[*MockService](c, "nonexistent")
	assert.Error(t, err)
	err = c.RegisterService("test", "not a service")
	assert.NoError(t, err)
	_, err = ServiceFor[*MockService](c, "test")
	assert.Error(t, err)
}

func TestCore_MustServiceFor_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc := MustServiceFor[*MockService](c, "test")
	assert.Equal(t, "test", svc.GetName())
}

func TestCore_MustServiceFor_Ugly(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// MustServiceFor panics on missing service
	assert.Panics(t, func() {
		MustServiceFor[*MockService](c, "nonexistent")
	})

	err = c.RegisterService("test", "not a service")
	assert.NoError(t, err)

	// MustServiceFor panics on type mismatch
	assert.Panics(t, func() {
		MustServiceFor[*MockService](c, "test")
	})
}

type MockAction struct {
	handled bool
}

func (a *MockAction) Handle(c *Core, msg Message) error {
	a.handled = true
	return nil
}

func TestCore_ACTION_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	action := &MockAction{}
	c.RegisterAction(action.Handle)
	err = c.ACTION(nil)
	assert.NoError(t, err)
	assert.True(t, action.handled)
}

func TestCore_RegisterActions_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	action1 := &MockAction{}
	action2 := &MockAction{}
	c.RegisterActions(action1.Handle, action2.Handle)
	err = c.ACTION(nil)
	assert.NoError(t, err)
	assert.True(t, action1.handled)
	assert.True(t, action2.handled)
}

func TestCore_WithName_Good(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return &MockService{Name: "test"}, nil
	}
	c, err := New(WithName("my-service", factory))
	assert.NoError(t, err)
	svc := c.Service("my-service")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_WithName_Bad(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return nil, assert.AnError
	}
	_, err := New(WithName("my-service", factory))
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestCore_GlobalInstance_ThreadSafety_Good(t *testing.T) {
	// Save original instance
	original := GetInstance()
	defer SetInstance(original)

	// Test SetInstance/GetInstance
	c1, _ := New()
	SetInstance(c1)
	assert.Equal(t, c1, GetInstance())

	// Test ClearInstance
	ClearInstance()
	assert.Nil(t, GetInstance())

	// Test concurrent access (race detector should catch issues)
	c2, _ := New(WithApp(&mockApp{}))
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			SetInstance(c2)
			_ = GetInstance()
			done <- true
		}()
		go func() {
			inst := GetInstance()
			if inst != nil {
				_ = inst.App
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
