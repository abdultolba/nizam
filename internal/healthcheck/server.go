package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Server provides HTTP endpoints for health check information
type Server struct {
	engine   *Engine
	server   *http.Server
	address  string
}

// NewServer creates a new health check HTTP server
func NewServer(engine *Engine, address string) *Server {
	if address == "" {
		address = ":8080"
	}

	return &Server{
		engine:  engine,
		address: address,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/health", s.handleHealthSummary)
	mux.HandleFunc("/api/health/", s.handleServiceHealth) // matches /api/health/{service}
	mux.HandleFunc("/api/services", s.handleAllServicesHealth)
	mux.HandleFunc("/api/check/", s.handleCheckServiceNow) // matches /api/check/{service}
	
	// Web UI endpoints
	mux.HandleFunc("/", s.handleWebDashboard)
	mux.HandleFunc("/service/", s.handleServiceDetails) // matches /service/{service}
	
	// Static assets
	mux.HandleFunc("/assets/", s.handleStaticAssets)
	
	s.server = &http.Server{
		Addr:    s.address,
		Handler: s.corsMiddleware(s.loggingMiddleware(mux)),
	}

	log.Info().Str("address", s.address).Msg("Health check HTTP server starting")

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down health check HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(shutdownCtx)
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health check server failed: %w", err)
	}

	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleHealthSummary returns overall health summary
func (s *Server) handleHealthSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := s.engine.GetHealthSummary()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Error().Err(err).Msg("Failed to encode health summary")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleServiceHealth returns health information for a specific service
func (s *Server) handleServiceHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/api/health/")
	if serviceName == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	healthInfo, exists := s.engine.GetServiceHealth(serviceName)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthInfo); err != nil {
		log.Error().Err(err).Str("service", serviceName).Msg("Failed to encode service health")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleAllServicesHealth returns health information for all services
func (s *Server) handleAllServicesHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	allHealth := s.engine.GetAllServicesHealth()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allHealth); err != nil {
		log.Error().Err(err).Msg("Failed to encode all services health")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleCheckServiceNow performs an immediate health check on a service
func (s *Server) handleCheckServiceNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/api/check/")
	if serviceName == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.engine.CheckServiceNow(ctx, serviceName)
	if err != nil {
		log.Error().Err(err).Str("service", serviceName).Msg("Failed to check service health")
		http.Error(w, fmt.Sprintf("Failed to check service: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error().Err(err).Str("service", serviceName).Msg("Failed to encode health check result")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleWebDashboard serves the web dashboard
func (s *Server) handleWebDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If path is not root, serve 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	summary := s.engine.GetHealthSummary()
	allHealth := s.engine.GetAllServicesHealth()

	data := struct {
		Summary    map[string]interface{}
		Services   map[string]*ServiceHealthInfo
		Timestamp  string
	}{
		Summary:   summary,
		Services:  allHealth,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl := template.Must(template.New("dashboard").Funcs(template.FuncMap{
		"GetHealthStatusColor": GetHealthStatusColor,
		"FormatDuration":       FormatDuration,
	}).Parse(dashboardTemplate))
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("Failed to render dashboard template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleServiceDetails serves detailed service health information
func (s *Server) handleServiceDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := strings.TrimPrefix(r.URL.Path, "/service/")
	if serviceName == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	healthInfo, exists := s.engine.GetServiceHealth(serviceName)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	data := struct {
		Service   *ServiceHealthInfo
		Timestamp string
	}{
		Service:   healthInfo,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	tmpl := template.Must(template.New("service").Funcs(template.FuncMap{
		"GetHealthStatusColor": GetHealthStatusColor,
		"FormatDuration":       FormatDuration,
	}).Parse(serviceDetailsTemplate))
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Error().Err(err).Str("service", serviceName).Msg("Failed to render service details template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleStaticAssets serves static CSS/JS assets
func (s *Server) handleStaticAssets(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/assets/")
	
	switch path {
	case "style.css":
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(cssStyles))
	case "script.js":
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(jsScripts))
	default:
		http.NotFound(w, r)
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request")
	})
}

// GetHealthStatusColor returns a color class for health status
func GetHealthStatusColor(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusStarting:
		return "starting"
	case HealthStatusNotRunning:
		return "not-running"
	default:
		return "unknown"
	}
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	
	return fmt.Sprintf("%.1fh", d.Hours())
}

// dashboardTemplate is the HTML template for the web dashboard
const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nizam Health Dashboard</title>
    <link rel="stylesheet" href="/assets/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>🏥 Nizam Health Dashboard</h1>
            <p class="subtitle">Real-time service health monitoring</p>
            <div class="timestamp">Last updated: {{.Timestamp}}</div>
        </header>

        <div class="summary-cards">
            <div class="card healthy">
                <div class="card-icon">✅</div>
                <div class="card-content">
                    <div class="card-value">{{index .Summary "healthy"}}</div>
                    <div class="card-label">Healthy</div>
                </div>
            </div>
            
            <div class="card unhealthy">
                <div class="card-icon">❌</div>
                <div class="card-content">
                    <div class="card-value">{{index .Summary "unhealthy"}}</div>
                    <div class="card-label">Unhealthy</div>
                </div>
            </div>
            
            <div class="card starting">
                <div class="card-icon">🔄</div>
                <div class="card-content">
                    <div class="card-value">{{index .Summary "starting"}}</div>
                    <div class="card-label">Starting</div>
                </div>
            </div>
            
            <div class="card not-running">
                <div class="card-icon">🛑</div>
                <div class="card-content">
                    <div class="card-value">{{index .Summary "not_running"}}</div>
                    <div class="card-label">Not Running</div>
                </div>
            </div>

            <div class="card unknown">
                <div class="card-icon">❓</div>
                <div class="card-content">
                    <div class="card-value">{{index .Summary "unknown"}}</div>
                    <div class="card-label">Unknown</div>
                </div>
            </div>
        </div>

        <div class="services-grid">
            {{range $name, $service := .Services}}
            <div class="service-card {{GetHealthStatusColor $service.Status}}">
                <div class="service-header">
                    <h3>{{$name}}</h3>
                    <span class="status-badge">{{$service.Status}}</span>
                </div>
                
                <div class="service-info">
                    {{if $service.ContainerName}}
                    <div class="info-row">
                        <span class="label">Container:</span>
                        <span class="value">{{$service.ContainerName}}</span>
                    </div>
                    {{end}}
                    
                    {{if $service.Image}}
                    <div class="info-row">
                        <span class="label">Image:</span>
                        <span class="value">{{$service.Image}}</span>
                    </div>
                    {{end}}
                    
                    <div class="info-row">
                        <span class="label">Last Check:</span>
                        <span class="value">{{$service.LastCheck.Format "15:04:05"}}</span>
                    </div>
                    
                    <div class="info-row">
                        <span class="label">Running:</span>
                        <span class="value">{{if $service.IsRunning}}Yes{{else}}No{{end}}</span>
                    </div>
                </div>
                
                <div class="service-actions">
                    <a href="/service/{{$name}}" class="btn btn-secondary">Details</a>
                    <button onclick="checkServiceNow('{{$name}}')" class="btn btn-primary">Check Now</button>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    
    <script src="/assets/script.js"></script>
</body>
</html>
`

// serviceDetailsTemplate is the HTML template for service details
const serviceDetailsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Service.ServiceName}} - Health Details</title>
    <link rel="stylesheet" href="/assets/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>🔍 {{.Service.ServiceName}} Health Details</h1>
            <div class="nav">
                <a href="/" class="btn btn-secondary">← Back to Dashboard</a>
                <button onclick="checkServiceNow('{{.Service.ServiceName}}')" class="btn btn-primary">Check Now</button>
            </div>
            <div class="timestamp">Last updated: {{.Timestamp}}</div>
        </header>

        <div class="service-overview">
            <div class="status-section {{GetHealthStatusColor .Service.Status}}">
                <h2>Current Status: {{.Service.Status}}</h2>
                <p>Last Check: {{.Service.LastCheck.Format "2006-01-02 15:04:05"}}</p>
            </div>
            
            <div class="info-grid">
                {{if .Service.ContainerName}}
                <div class="info-item">
                    <strong>Container Name:</strong> {{.Service.ContainerName}}
                </div>
                {{end}}
                
                {{if .Service.ContainerID}}
                <div class="info-item">
                    <strong>Container ID:</strong> {{.Service.ContainerID}}
                </div>
                {{end}}
                
                {{if .Service.Image}}
                <div class="info-item">
                    <strong>Image:</strong> {{.Service.Image}}
                </div>
                {{end}}
                
                <div class="info-item">
                    <strong>Running:</strong> {{if .Service.IsRunning}}Yes{{else}}No{{end}}
                </div>
            </div>
        </div>

        {{if .Service.Configuration}}
        <div class="config-section">
            <h3>Health Check Configuration</h3>
            <div class="config-details">
                {{if .Service.Configuration.Test}}
                <div class="config-item">
                    <strong>Test Command:</strong>
                    <pre>{{range .Service.Configuration.Test}}{{.}} {{end}}</pre>
                </div>
                {{end}}
                
                {{if .Service.Configuration.Interval}}
                <div class="config-item">
                    <strong>Interval:</strong> {{.Service.Configuration.Interval}}
                </div>
                {{end}}
                
                {{if .Service.Configuration.Timeout}}
                <div class="config-item">
                    <strong>Timeout:</strong> {{.Service.Configuration.Timeout}}
                </div>
                {{end}}
                
                {{if .Service.Configuration.Retries}}
                <div class="config-item">
                    <strong>Retries:</strong> {{.Service.Configuration.Retries}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        <div class="history-section">
            <h3>Check History</h3>
            <div class="history-list">
                {{range .Service.CheckHistory}}
                <div class="history-item {{GetHealthStatusColor .Status}}">
                    <div class="history-header">
                        <span class="status">{{.Status}}</span>
                        <span class="timestamp">{{.Timestamp.Format "15:04:05"}}</span>
                        <span class="duration">({{FormatDuration .Duration}})</span>
                    </div>
                    <div class="history-message">{{.Message}}</div>
                    {{if .Details}}
                    <details class="history-details">
                        <summary>Details</summary>
                        <pre>{{printf "%+v" .Details}}</pre>
                    </details>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <script src="/assets/script.js"></script>
</body>
</html>
`

// cssStyles contains the CSS for the web dashboard
const cssStyles = `
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: linear-gradient(135deg, #0a0f1c 0%, #1a2332 100%);
    color: #e4e4e7;
    min-height: 100vh;
    line-height: 1.6;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    text-align: center;
    margin-bottom: 40px;
    padding: 30px 0;
}

header h1 {
    font-size: 2.5rem;
    font-weight: 700;
    color: #00d2ff;
    margin-bottom: 10px;
    text-shadow: 0 0 20px rgba(0, 210, 255, 0.3);
}

.subtitle {
    font-size: 1.2rem;
    color: #94a3b8;
    margin-bottom: 15px;
}

.timestamp {
    font-size: 0.9rem;
    color: #64748b;
    font-family: 'Monaco', monospace;
}

.nav {
    display: flex;
    gap: 15px;
    justify-content: center;
    margin: 20px 0;
}

.summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    margin-bottom: 40px;
}

.card {
    background: rgba(30, 41, 59, 0.8);
    border-radius: 12px;
    padding: 25px;
    display: flex;
    align-items: center;
    gap: 20px;
    border: 1px solid rgba(148, 163, 184, 0.2);
    transition: all 0.3s ease;
}

.card:hover {
    transform: translateY(-5px);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.card-icon {
    font-size: 2.5rem;
}

.card-content {
    flex: 1;
}

.card-value {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: 5px;
}

.card-label {
    font-size: 0.9rem;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.card.healthy { border-left: 4px solid #10b981; }
.card.unhealthy { border-left: 4px solid #ef4444; }
.card.starting { border-left: 4px solid #f59e0b; }
.card.not-running { border-left: 4px solid #6b7280; }
.card.unknown { border-left: 4px solid #8b5cf6; }

.services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 25px;
}

.service-card {
    background: rgba(30, 41, 59, 0.8);
    border-radius: 12px;
    padding: 25px;
    border: 1px solid rgba(148, 163, 184, 0.2);
    transition: all 0.3s ease;
}

.service-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.service-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
}

.service-header h3 {
    font-size: 1.3rem;
    color: #00d2ff;
}

.status-badge {
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.service-card.healthy .status-badge {
    background: rgba(16, 185, 129, 0.2);
    color: #10b981;
    border: 1px solid #10b981;
}

.service-card.unhealthy .status-badge {
    background: rgba(239, 68, 68, 0.2);
    color: #ef4444;
    border: 1px solid #ef4444;
}

.service-card.starting .status-badge {
    background: rgba(245, 158, 11, 0.2);
    color: #f59e0b;
    border: 1px solid #f59e0b;
}

.service-card.not-running .status-badge {
    background: rgba(107, 114, 128, 0.2);
    color: #6b7280;
    border: 1px solid #6b7280;
}

.service-card.unknown .status-badge {
    background: rgba(139, 92, 246, 0.2);
    color: #8b5cf6;
    border: 1px solid #8b5cf6;
}

.service-info {
    margin-bottom: 20px;
}

.info-row {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 0.9rem;
}

.info-row .label {
    color: #94a3b8;
    font-weight: 500;
}

.info-row .value {
    color: #e4e4e7;
    font-family: 'Monaco', monospace;
}

.service-actions {
    display: flex;
    gap: 10px;
}

.btn {
    padding: 10px 16px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    text-decoration: none;
    display: inline-block;
    text-align: center;
    transition: all 0.2s ease;
}

.btn-primary {
    background: #00d2ff;
    color: #0a0f1c;
}

.btn-primary:hover {
    background: #00a8cc;
    transform: translateY(-2px);
}

.btn-secondary {
    background: rgba(148, 163, 184, 0.2);
    color: #e4e4e7;
    border: 1px solid rgba(148, 163, 184, 0.3);
}

.btn-secondary:hover {
    background: rgba(148, 163, 184, 0.3);
    transform: translateY(-2px);
}

/* Service Details Page */
.service-overview {
    margin-bottom: 40px;
}

.status-section {
    background: rgba(30, 41, 59, 0.8);
    border-radius: 12px;
    padding: 25px;
    margin-bottom: 25px;
    border-left: 4px solid #6b7280;
}

.status-section.healthy { border-left-color: #10b981; }
.status-section.unhealthy { border-left-color: #ef4444; }
.status-section.starting { border-left-color: #f59e0b; }
.status-section.not-running { border-left-color: #6b7280; }
.status-section.unknown { border-left-color: #8b5cf6; }

.info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 15px;
}

.info-item {
    background: rgba(30, 41, 59, 0.8);
    padding: 15px;
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.2);
}

.config-section,
.history-section {
    background: rgba(30, 41, 59, 0.8);
    border-radius: 12px;
    padding: 25px;
    margin-bottom: 25px;
    border: 1px solid rgba(148, 163, 184, 0.2);
}

.config-details {
    display: grid;
    gap: 15px;
}

.config-item {
    padding: 15px;
    background: rgba(15, 23, 42, 0.6);
    border-radius: 8px;
    border: 1px solid rgba(148, 163, 184, 0.1);
}

.config-item pre {
    margin-top: 10px;
    padding: 10px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 4px;
    font-family: 'Monaco', monospace;
    font-size: 0.85rem;
    overflow-x: auto;
}

.history-list {
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.history-item {
    padding: 15px;
    background: rgba(15, 23, 42, 0.6);
    border-radius: 8px;
    border-left: 4px solid #6b7280;
}

.history-item.healthy { border-left-color: #10b981; }
.history-item.unhealthy { border-left-color: #ef4444; }
.history-item.starting { border-left-color: #f59e0b; }
.history-item.not-running { border-left-color: #6b7280; }
.history-item.unknown { border-left-color: #8b5cf6; }

.history-header {
    display: flex;
    gap: 15px;
    align-items: center;
    margin-bottom: 8px;
    font-size: 0.9rem;
}

.history-header .status {
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.history-header .timestamp {
    color: #94a3b8;
    font-family: 'Monaco', monospace;
}

.history-header .duration {
    color: #64748b;
    font-size: 0.8rem;
}

.history-message {
    color: #e4e4e7;
    margin-bottom: 10px;
}

.history-details {
    margin-top: 10px;
}

.history-details summary {
    cursor: pointer;
    color: #00d2ff;
    font-size: 0.9rem;
}

.history-details pre {
    margin-top: 10px;
    padding: 10px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 4px;
    font-family: 'Monaco', monospace;
    font-size: 0.8rem;
    overflow-x: auto;
    max-height: 200px;
    overflow-y: auto;
}

/* Responsive */
@media (max-width: 768px) {
    .container {
        padding: 15px;
    }
    
    header h1 {
        font-size: 2rem;
    }
    
    .summary-cards {
        grid-template-columns: 1fr;
    }
    
    .services-grid {
        grid-template-columns: 1fr;
    }
    
    .service-header {
        flex-direction: column;
        gap: 10px;
        align-items: flex-start;
    }
    
    .service-actions {
        flex-direction: column;
    }
    
    .nav {
        flex-direction: column;
    }
}
`

// jsScripts contains the JavaScript for the web dashboard
const jsScripts = `
// Auto-refresh the page every 30 seconds
setInterval(() => {
    if (document.visibilityState === 'visible') {
        location.reload();
    }
}, 30000);

// Check service now function
async function checkServiceNow(serviceName) {
    const button = event.target;
    const originalText = button.textContent;
    
    button.textContent = 'Checking...';
    button.disabled = true;
    
    try {
        const response = await fetch('/api/check/' + serviceName, {
            method: 'POST'
        });
        
        if (response.ok) {
            const result = await response.json();
            // Reload the page to show updated results
            location.reload();
        } else {
            alert('Failed to check service: ' + response.statusText);
        }
    } catch (error) {
        alert('Error checking service: ' + error.message);
    } finally {
        button.textContent = originalText;
        button.disabled = false;
    }
}

// Handle visibility change to pause auto-refresh when tab is not visible
document.addEventListener('visibilitychange', () => {
    // Auto-refresh will only work when the page is visible
});

// Add keyboard shortcuts
document.addEventListener('keydown', (event) => {
    if (event.key === 'r' || event.key === 'R') {
        if (!event.ctrlKey && !event.metaKey) {
            event.preventDefault();
            location.reload();
        }
    }
});

// Add loading animation to buttons
document.addEventListener('DOMContentLoaded', () => {
    const buttons = document.querySelectorAll('.btn');
    buttons.forEach(button => {
        if (button.textContent.includes('Check Now')) {
            button.addEventListener('click', (event) => {
                // Animation will be handled by checkServiceNow function
            });
        }
    });
});
`
