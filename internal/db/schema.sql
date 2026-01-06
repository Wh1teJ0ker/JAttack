CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE IF NOT EXISTS info_gathering (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target TEXT NOT NULL,
    info_type TEXT,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vulnerabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    product TEXT,
    vuln_type TEXT,
    severity TEXT,
    description TEXT,
    details TEXT,
    status TEXT,
    poc_type TEXT,
    poc_content TEXT,
    reference TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    port INTEGER,
    protocol TEXT,
    probe TEXT,
    match_rule TEXT,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS weak_passwords (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target TEXT NOT NULL,
    port INTEGER NOT NULL,
    service TEXT NOT NULL,
    username TEXT,
    password TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dir_scan_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target TEXT NOT NULL,
    url TEXT NOT NULL,
    status INTEGER,
    size INTEGER,
    location TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- New Asset-Centric Schema --

-- 1. Assets (IPs)
CREATE TABLE IF NOT EXISTS assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip TEXT NOT NULL UNIQUE,
    os TEXT,
    alive BOOLEAN DEFAULT FALSE,
    last_scan_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Ports (Belong to Assets)
CREATE TABLE IF NOT EXISTS asset_ports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id INTEGER NOT NULL,
    port INTEGER NOT NULL,
    protocol TEXT DEFAULT 'tcp',
    service TEXT,
    product TEXT,
    version TEXT,
    banner TEXT,
    state TEXT DEFAULT 'open',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    UNIQUE(asset_id, port, protocol)
);

-- 3. Web Services (Belong to Ports)
CREATE TABLE IF NOT EXISTS web_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id INTEGER NOT NULL,
    port_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    title TEXT,
    server TEXT,
    fingerprints TEXT, -- JSON array of detected technologies
    screenshot_path TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY(port_id) REFERENCES asset_ports(id) ON DELETE CASCADE,
    UNIQUE(port_id, url)
);

-- 4. Web Directories (Belong to Web Services)
CREATE TABLE IF NOT EXISTS web_directories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    web_service_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER,
    content_length INTEGER,
    title TEXT,
    content_type TEXT,
    redirect_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(web_service_id) REFERENCES web_services(id) ON DELETE CASCADE
);

-- 5. JS Files (Belong to Web Services)
CREATE TABLE IF NOT EXISTS web_js_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    web_service_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    full_url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(web_service_id) REFERENCES web_services(id) ON DELETE CASCADE
);

-- 6. Sensitive Info (Belong to Web Service or JS File)
CREATE TABLE IF NOT EXISTS sensitive_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    web_service_id INTEGER NOT NULL,
    source_file TEXT, -- URL where it was found
    info_type TEXT NOT NULL, -- 'AWS', 'API Key', 'Email'
    content TEXT,
    context TEXT, -- Surrounding text
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(web_service_id) REFERENCES web_services(id) ON DELETE CASCADE
);

-- 7. Auth Results (Weak Passwords)
CREATE TABLE IF NOT EXISTS auth_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id INTEGER NOT NULL,
    port_id INTEGER NOT NULL,
    service_type TEXT NOT NULL, -- 'ssh', 'ftp', 'mysql'
    username TEXT,
    password TEXT,
    success BOOLEAN,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY(port_id) REFERENCES asset_ports(id) ON DELETE CASCADE
);
