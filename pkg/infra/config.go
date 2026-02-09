// Package infra provides infrastructure configuration and API clients
// for managing the Host UK production environment.
package infra

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the top-level infrastructure configuration parsed from infra.yaml.
type Config struct {
	Hosts        map[string]*Host      `yaml:"hosts"`
	LoadBalancer LoadBalancer          `yaml:"load_balancer"`
	Network      Network               `yaml:"network"`
	DNS          DNS                   `yaml:"dns"`
	SSL          SSL                   `yaml:"ssl"`
	Database     Database              `yaml:"database"`
	Cache        Cache                 `yaml:"cache"`
	Containers   map[string]*Container `yaml:"containers"`
	S3           S3Config              `yaml:"s3"`
	CDN          CDN                   `yaml:"cdn"`
	CICD         CICD                  `yaml:"cicd"`
	Monitoring   Monitoring            `yaml:"monitoring"`
	Backups      Backups               `yaml:"backups"`
}

// Host represents a server in the infrastructure.
type Host struct {
	FQDN      string   `yaml:"fqdn"`
	IP        string   `yaml:"ip"`
	PrivateIP string   `yaml:"private_ip,omitempty"`
	Type      string   `yaml:"type"` // hcloud, hrobot
	Role      string   `yaml:"role"` // bastion, app, builder
	SSH       SSHConf  `yaml:"ssh"`
	Services  []string `yaml:"services"`
}

// SSHConf holds SSH connection details for a host.
type SSHConf struct {
	User string `yaml:"user"`
	Key  string `yaml:"key"`
	Port int    `yaml:"port"`
}

// LoadBalancer represents a Hetzner managed load balancer.
type LoadBalancer struct {
	Name      string      `yaml:"name"`
	FQDN      string      `yaml:"fqdn"`
	Provider  string      `yaml:"provider"`
	Type      string      `yaml:"type"`
	Location  string      `yaml:"location"`
	Algorithm string      `yaml:"algorithm"`
	Backends  []Backend   `yaml:"backends"`
	Health    HealthCheck `yaml:"health_check"`
	Listeners []Listener  `yaml:"listeners"`
	SSL       LBCert      `yaml:"ssl"`
}

// Backend is a load balancer backend target.
type Backend struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// HealthCheck configures load balancer health checking.
type HealthCheck struct {
	Protocol string `yaml:"protocol"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

// Listener maps a frontend port to a backend port.
type Listener struct {
	Frontend      int    `yaml:"frontend"`
	Backend       int    `yaml:"backend"`
	Protocol      string `yaml:"protocol"`
	ProxyProtocol bool   `yaml:"proxy_protocol"`
}

// LBCert holds the SSL certificate configuration for the load balancer.
type LBCert struct {
	Certificate string   `yaml:"certificate"`
	SAN         []string `yaml:"san"`
}

// Network describes the private network.
type Network struct {
	CIDR string `yaml:"cidr"`
	Name string `yaml:"name"`
}

// DNS holds DNS provider configuration and zone records.
type DNS struct {
	Provider    string           `yaml:"provider"`
	Nameservers []string         `yaml:"nameservers"`
	Zones       map[string]*Zone `yaml:"zones"`
}

// Zone is a DNS zone with its records.
type Zone struct {
	Records []DNSRecord `yaml:"records"`
}

// DNSRecord is a single DNS record.
type DNSRecord struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
	TTL   int    `yaml:"ttl"`
}

// SSL holds SSL certificate configuration.
type SSL struct {
	Wildcard WildcardCert `yaml:"wildcard"`
}

// WildcardCert describes a wildcard SSL certificate.
type WildcardCert struct {
	Domains     []string `yaml:"domains"`
	Method      string   `yaml:"method"`
	DNSProvider string   `yaml:"dns_provider"`
	Termination string   `yaml:"termination"`
}

// Database describes the database cluster.
type Database struct {
	Engine    string       `yaml:"engine"`
	Version   string       `yaml:"version"`
	Cluster   string       `yaml:"cluster"`
	Nodes     []DBNode     `yaml:"nodes"`
	SSTMethod string       `yaml:"sst_method"`
	Backup    BackupConfig `yaml:"backup"`
}

// DBNode is a database cluster node.
type DBNode struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// BackupConfig describes automated backup settings.
type BackupConfig struct {
	Schedule    string `yaml:"schedule"`
	Destination string `yaml:"destination"`
	Bucket      string `yaml:"bucket"`
	Prefix      string `yaml:"prefix"`
}

// Cache describes the cache/session cluster.
type Cache struct {
	Engine   string      `yaml:"engine"`
	Version  string      `yaml:"version"`
	Sentinel bool        `yaml:"sentinel"`
	Nodes    []CacheNode `yaml:"nodes"`
}

// CacheNode is a cache cluster node.
type CacheNode struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Container describes a container deployment.
type Container struct {
	Image     string   `yaml:"image"`
	Port      int      `yaml:"port,omitempty"`
	Runtime   string   `yaml:"runtime,omitempty"`
	Command   string   `yaml:"command,omitempty"`
	Replicas  int      `yaml:"replicas,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
}

// S3Config describes object storage.
type S3Config struct {
	Endpoint string               `yaml:"endpoint"`
	Buckets  map[string]*S3Bucket `yaml:"buckets"`
}

// S3Bucket is an S3 bucket configuration.
type S3Bucket struct {
	Purpose string   `yaml:"purpose"`
	Paths   []string `yaml:"paths"`
}

// CDN describes CDN configuration.
type CDN struct {
	Provider string   `yaml:"provider"`
	Origin   string   `yaml:"origin"`
	Zones    []string `yaml:"zones"`
}

// CICD describes CI/CD configuration.
type CICD struct {
	Provider   string `yaml:"provider"`
	URL        string `yaml:"url"`
	Runner     string `yaml:"runner"`
	Registry   string `yaml:"registry"`
	DeployHook string `yaml:"deploy_hook"`
}

// Monitoring describes monitoring configuration.
type Monitoring struct {
	HealthEndpoints []HealthEndpoint `yaml:"health_endpoints"`
	Alerts          map[string]int   `yaml:"alerts"`
}

// HealthEndpoint is a URL to monitor.
type HealthEndpoint struct {
	URL      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

// Backups describes backup schedules.
type Backups struct {
	Daily  []BackupJob `yaml:"daily"`
	Weekly []BackupJob `yaml:"weekly"`
}

// BackupJob is a scheduled backup task.
type BackupJob struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Destination string   `yaml:"destination,omitempty"`
	Hosts       []string `yaml:"hosts,omitempty"`
}

// Load reads and parses an infra.yaml file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read infra config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse infra config: %w", err)
	}

	// Expand SSH key paths
	for _, h := range cfg.Hosts {
		if h.SSH.Key != "" {
			h.SSH.Key = expandPath(h.SSH.Key)
		}
		if h.SSH.Port == 0 {
			h.SSH.Port = 22
		}
	}

	return &cfg, nil
}

// Discover searches for infra.yaml in the given directory and parent directories.
func Discover(startDir string) (*Config, string, error) {
	dir := startDir
	for {
		path := filepath.Join(dir, "infra.yaml")
		if _, err := os.Stat(path); err == nil {
			cfg, err := Load(path)
			return cfg, path, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, "", fmt.Errorf("infra.yaml not found (searched from %s)", startDir)
}

// HostsByRole returns all hosts matching the given role.
func (c *Config) HostsByRole(role string) map[string]*Host {
	result := make(map[string]*Host)
	for name, h := range c.Hosts {
		if h.Role == role {
			result[name] = h
		}
	}
	return result
}

// AppServers returns hosts with role "app".
func (c *Config) AppServers() map[string]*Host {
	return c.HostsByRole("app")
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
