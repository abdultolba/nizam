package resolve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abdultolba/nizam/internal/config"
)

// ServiceInfo holds resolved service connection information
type ServiceInfo struct {
	Name      string // Service name
	Engine    string // Engine type: postgres, mysql, redis, mongo
	Host      string // Usually "localhost"
	Port      int    // Service port
	User      string // Username
	Password  string // Password
	Database  string // Database name (for SQL databases)
	Container string // Container name
	Image     string // Docker image
}

// GetServiceInfo resolves service information from config
func GetServiceInfo(cfg *config.Config, serviceName string) (ServiceInfo, error) {
	service, exists := cfg.GetService(serviceName)
	if !exists {
		return ServiceInfo{}, fmt.Errorf("service '%s' not found in config", serviceName)
	}

	info := ServiceInfo{
		Name:      serviceName,
		Container: fmt.Sprintf("nizam_%s", serviceName),
		Image:     service.Image,
		Host:      "localhost",
	}

	// Determine engine from image or service name
	info.Engine = DetermineEngine(service.Image, serviceName)

	// Parse ports to get the host port
	if len(service.Ports) > 0 {
		portMapping := service.Ports[0] // Use first port mapping
		parts := strings.Split(portMapping, ":")
		if len(parts) == 2 {
			hostPort, err := strconv.Atoi(parts[0])
			if err != nil {
				return ServiceInfo{}, fmt.Errorf("invalid port mapping '%s': %w", portMapping, err)
			}
			info.Port = hostPort
		}
	}

	// Extract credentials and database from environment variables
	for key, value := range service.Environment {
		switch strings.ToLower(key) {
		case "postgres_user", "mysql_user", "mongo_initdb_root_username", "user":
			info.User = value
		case "postgres_password", "mysql_password", "mongo_initdb_root_password", "password":
			info.Password = value
		case "postgres_db", "mysql_database", "mongo_initdb_database", "database", "db":
			info.Database = value
		}
	}

	// Set defaults based on engine
	setDefaults(&info)

	return info, nil
}

// DetermineEngine determines the database engine from image name or service name
func DetermineEngine(image, serviceName string) string {
	image = strings.ToLower(image)
	serviceName = strings.ToLower(serviceName)

	// Check image name first
	if strings.Contains(image, "postgres") {
		return "postgres"
	}
	if strings.Contains(image, "mysql") || strings.Contains(image, "mariadb") {
		return "mysql"
	}
	if strings.Contains(image, "redis") {
		return "redis"
	}
	if strings.Contains(image, "mongo") {
		return "mongo"
	}

	// Fallback to service name
	if strings.Contains(serviceName, "postgres") || strings.Contains(serviceName, "pg") {
		return "postgres"
	}
	if strings.Contains(serviceName, "mysql") {
		return "mysql"
	}
	if strings.Contains(serviceName, "redis") {
		return "redis"
	}
	if strings.Contains(serviceName, "mongo") {
		return "mongo"
	}

	// Default fallback
	return "postgres"
}

// setDefaults sets default values based on the engine type
func setDefaults(info *ServiceInfo) {
	switch info.Engine {
	case "postgres":
		if info.Port == 0 {
			info.Port = 5432
		}
		if info.User == "" {
			info.User = "postgres"
		}
		if info.Database == "" {
			info.Database = "postgres"
		}

	case "mysql":
		if info.Port == 0 {
			info.Port = 3306
		}
		if info.User == "" {
			info.User = "root"
		}
		if info.Database == "" {
			info.Database = "mysql"
		}

	case "redis":
		if info.Port == 0 {
			info.Port = 6379
		}
		// Redis doesn't have users/databases by default

	case "mongo":
		if info.Port == 0 {
			info.Port = 27017
		}
		if info.User == "" {
			info.User = "root"
		}
		if info.Database == "" {
			info.Database = "admin"
		}
	}

	// Set default password if not specified
	if info.Password == "" && info.Engine != "redis" {
		info.Password = "password"
	}
}

// GetConnectionString builds a connection string for the service
func (si ServiceInfo) GetConnectionString() string {
	switch si.Engine {
	case "postgres":
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			si.User, si.Password, si.Host, si.Port, si.Database)

	case "mysql":
		return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
			si.User, si.Password, si.Host, si.Port, si.Database)

	case "redis":
		if si.Password != "" {
			return fmt.Sprintf("redis://:%s@%s:%d", si.Password, si.Host, si.Port)
		}
		return fmt.Sprintf("redis://%s:%d", si.Host, si.Port)

	case "mongo":
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			si.User, si.Password, si.Host, si.Port, si.Database)

	default:
		return ""
	}
}

// GetClientArgs returns the appropriate client command arguments
func (si ServiceInfo) GetClientArgs() []string {
	switch si.Engine {
	case "postgres":
		return []string{"psql", si.GetConnectionString()}

	case "mysql":
		args := []string{"mysql"}
		if si.User != "" {
			args = append(args, "-u", si.User)
		}
		if si.Password != "" {
			args = append(args, fmt.Sprintf("-p%s", si.Password))
		}
		if si.Host != "" {
			args = append(args, "-h", si.Host)
		}
		if si.Port != 0 {
			args = append(args, "-P", strconv.Itoa(si.Port))
		}
		if si.Database != "" {
			args = append(args, si.Database)
		}
		return args

	case "redis":
		args := []string{"redis-cli"}
		if si.Host != "" {
			args = append(args, "-h", si.Host)
		}
		if si.Port != 0 {
			args = append(args, "-p", strconv.Itoa(si.Port))
		}
		if si.Password != "" {
			args = append(args, "-a", si.Password)
		}
		return args

	case "mongo":
		return []string{"mongosh", si.GetConnectionString()}

	default:
		return []string{}
	}
}

// HasHostClient checks if the client binary is available on the host
func HasHostClient(engine string) bool {
	// TODO: Implement actual binary checking
	// For now, assume clients are not available and fallback to docker exec
	return false
}
