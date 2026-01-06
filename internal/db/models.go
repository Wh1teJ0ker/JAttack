package db

import "time"

// Asset represents a host/IP
type Asset struct {
	ID           int64     `json:"id"`
	IP           string    `json:"ip"`
	OS           string    `json:"os"`
	Alive        bool      `json:"alive"`
	LastScanTime time.Time `json:"last_scan_time"`
	CreatedAt    time.Time `json:"created_at"`
}

// AssetPort represents a port on an asset
type AssetPort struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Service   string    `json:"service"`
	Product   string    `json:"product"`
	Version   string    `json:"version"`
	Banner    string    `json:"banner"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebService represents a web service running on a port
type WebService struct {
	ID             int64     `json:"id"`
	AssetID        int64     `json:"asset_id"`
	PortID         int64     `json:"port_id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Server         string    `json:"server"`
	Fingerprints   string    `json:"fingerprints"` // JSON string
	ScreenshotPath string    `json:"screenshot_path"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WebDirectory represents a discovered directory
type WebDirectory struct {
	ID            int64     `json:"id"`
	WebServiceID  int64     `json:"web_service_id"`
	Path          string    `json:"path"`
	StatusCode    int       `json:"status_code"`
	ContentLength int       `json:"content_length"`
	Title         string    `json:"title"`
	ContentType   string    `json:"content_type"`
	RedirectURL   string    `json:"redirect_url"`
	CreatedAt     time.Time `json:"created_at"`
}

// WebJSFile represents a JS file found on a web service
type WebJSFile struct {
	ID           int64     `json:"id"`
	WebServiceID int64     `json:"web_service_id"`
	Path         string    `json:"path"`
	FullURL      string    `json:"full_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// SensitiveResult represents sensitive info found
type SensitiveResult struct {
	ID           int64     `json:"id"`
	WebServiceID int64     `json:"web_service_id"`
	SourceFile   string    `json:"source_file"`
	InfoType     string    `json:"info_type"`
	Content      string    `json:"content"`
	Context      string    `json:"context"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthResult represents brute force results
type AuthResult struct {
	ID          int64     `json:"id"`
	AssetID     int64     `json:"asset_id"`
	PortID      int64     `json:"port_id"`
	ServiceType string    `json:"service_type"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Success     bool      `json:"success"`
	CreatedAt   time.Time `json:"created_at"`
}
