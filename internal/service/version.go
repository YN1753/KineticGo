package service

// Version 由 CI 通过 -ldflags 注入，本地 dev 时保持 "dev"
var Version = "dev"

const (
	GitHubOwner = "YN1753"
	GitHubRepo  = "KineticGo"
)
