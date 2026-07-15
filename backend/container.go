package backend

import (
	"fmt"
	"sync"
)

type ServiceContainer struct {
	mu       sync.RWMutex
	services map[string]interface{}
}

func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		services: make(map[string]interface{}),
	}
}

func (c *ServiceContainer) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

func (c *ServiceContainer) Get(name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.services[name]
}

func (c *ServiceContainer) GetAppService() *AppService {
	s := c.Get("AppService")
	if s == nil {
		return nil
	}
	return s.(*AppService)
}

func (c *ServiceContainer) GetMINIMTSService() *MINIMTSService {
	s := c.Get("MINIMTSService")
	if s == nil {
		return nil
	}
	return s.(*MINIMTSService)
}

func (c *ServiceContainer) GetHIKCameraService() *HIKCameraService {
	s := c.Get("HIKCameraService")
	if s == nil {
		return nil
	}
	return s.(*HIKCameraService)
}

func (c *ServiceContainer) GetProjectService() *ProjectService {
	s := c.Get("ProjectService")
	if s == nil {
		return nil
	}
	return s.(*ProjectService)
}

func (c *ServiceContainer) GetSystemService() *SystemService {
	s := c.Get("SystemService")
	if s == nil {
		return nil
	}
	return s.(*SystemService)
}

func (c *ServiceContainer) GetLoginService() *LoginService {
	s := c.Get("LoginService")
	if s == nil {
		return nil
	}
	return s.(*LoginService)
}

func (c *ServiceContainer) GetUpdateService() *UpdateService {
	s := c.Get("UpdateService")
	if s == nil {
		return nil
	}
	return s.(*UpdateService)
}

func (c *ServiceContainer) GetUser() *User {
	s := c.Get("User")
	if s == nil {
		return nil
	}
	return s.(*User)
}

func (c *ServiceContainer) SetUser(user *User) {
	c.Register("User", user)
}

func (c *ServiceContainer) InitAll() error {
	for name, service := range c.services {
		if initable, ok := service.(interface{ Init(*ServiceContainer) error }); ok {
			if err := initable.Init(c); err != nil {
				return fmt.Errorf("failed to init %s: %w", name, err)
			}
		}
	}
	return nil
}
