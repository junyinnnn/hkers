package service

// Container holds all application services.
// Pass this to handlers instead of individual services.
type Container struct {
	Auth AuthServiceInterface
	User UserServiceInterface
	// Add more services as needed:
	// Email *EmailService
}

// NewContainer creates a new service container.
func NewContainer(authSvc AuthServiceInterface, userSvc UserServiceInterface) *Container {
	return &Container{
		Auth: authSvc,
		User: userSvc,
	}
}
