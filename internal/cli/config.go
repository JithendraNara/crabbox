package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Profile     string
	Provider    string
	Class       string
	ServerType  string
	Coordinator string
	CoordToken  string
	Location    string
	Image       string
	AWSRegion   string
	AWSAMI      string
	AWSSGID     string
	AWSSubnetID string
	AWSProfile  string
	AWSRootGB   int32
	SSHUser     string
	SSHKey      string
	SSHPort     string
	ProviderKey string
	WorkRoot    string
	TTL         time.Duration
}

func defaultConfig() Config {
	home, _ := os.UserHomeDir()
	sshKey := os.Getenv("CRABBOX_SSH_KEY")
	if sshKey == "" && home != "" {
		sshKey = filepath.Join(home, ".ssh", "id_ed25519")
	}

	class := getenv("CRABBOX_DEFAULT_CLASS", "beast")
	provider := getenv("CRABBOX_PROVIDER", "hetzner")
	return Config{
		Profile:     getenv("CRABBOX_PROFILE", "openclaw-check"),
		Provider:    provider,
		Class:       class,
		ServerType:  getenv("CRABBOX_SERVER_TYPE", serverTypeForProviderClass(provider, class)),
		Coordinator: os.Getenv("CRABBOX_COORDINATOR"),
		CoordToken:  os.Getenv("CRABBOX_COORDINATOR_TOKEN"),
		Location:    getenv("CRABBOX_HETZNER_LOCATION", "fsn1"),
		Image:       getenv("CRABBOX_HETZNER_IMAGE", "ubuntu-24.04"),
		AWSRegion:   getenv("CRABBOX_AWS_REGION", getenv("AWS_REGION", "eu-west-1")),
		AWSAMI:      os.Getenv("CRABBOX_AWS_AMI"),
		AWSSGID:     os.Getenv("CRABBOX_AWS_SECURITY_GROUP_ID"),
		AWSSubnetID: os.Getenv("CRABBOX_AWS_SUBNET_ID"),
		AWSProfile:  os.Getenv("CRABBOX_AWS_INSTANCE_PROFILE"),
		AWSRootGB:   int32(getenvInt("CRABBOX_AWS_ROOT_GB", 400)),
		SSHUser:     getenv("CRABBOX_SSH_USER", "crabbox"),
		SSHKey:      sshKey,
		SSHPort:     getenv("CRABBOX_SSH_PORT", "2222"),
		ProviderKey: getenv("CRABBOX_HETZNER_SSH_KEY", "crabbox-steipete"),
		WorkRoot:    getenv("CRABBOX_WORK_ROOT", "/work/crabbox"),
		TTL:         90 * time.Minute,
	}
}

func serverTypeForClass(class string) string {
	return serverTypeCandidatesForClass(class)[0]
}

func serverTypeForProviderClass(provider, class string) string {
	if provider == "aws" {
		return awsInstanceTypeCandidatesForClass(class)[0]
	}
	return serverTypeForClass(class)
}

func serverTypeCandidatesForClass(class string) []string {
	switch class {
	case "standard":
		return []string{"ccx33", "cpx62", "cx53"}
	case "fast":
		return []string{"ccx43", "cpx62", "cx53"}
	case "large":
		return []string{"ccx53", "ccx43", "cpx62", "cx53"}
	case "beast":
		return []string{"ccx63", "ccx53", "ccx43", "cpx62", "cx53"}
	default:
		return []string{class}
	}
}

func awsInstanceTypeCandidatesForClass(class string) []string {
	switch class {
	case "standard":
		return []string{"c7a.8xlarge", "c7a.4xlarge"}
	case "fast":
		return []string{"c7a.16xlarge", "c7a.12xlarge", "c7a.8xlarge"}
	case "large":
		return []string{"c7a.24xlarge", "c7a.16xlarge", "c7a.12xlarge"}
	case "beast":
		return []string{"c7a.48xlarge", "c7a.32xlarge", "c7a.24xlarge", "c7a.16xlarge"}
	default:
		return []string{class}
	}
}

func getenv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func getenvInt(name string, fallback int) int {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
