package db

import (
	"database/sql"
	"time"
)

// --- Assets ---

// UpsertAsset inserts or updates an asset. Returns the ID.
func (m *Manager) UpsertAsset(ip string, os string, alive bool) (int64, error) {
	var id int64
	err := m.ExecTask(func(db *sql.DB) error {
		// Check if exists
		err := db.QueryRow("SELECT id FROM assets WHERE ip = ?", ip).Scan(&id)
		if err == sql.ErrNoRows {
			// Insert
			now := time.Now()
			var res sql.Result
			res, err = db.Exec("INSERT INTO assets (ip, os, alive, last_scan_time, created_at) VALUES (?, ?, ?, ?, ?)", ip, os, alive, now, now)
			if err != nil {
				return err
			}
			id, err = res.LastInsertId()
			return err
		} else if err != nil {
			return err
		}

		// Update
		_, err = db.Exec("UPDATE assets SET os = ?, alive = ?, last_scan_time = ? WHERE id = ?", os, alive, time.Now(), id)
		return err
	})
	return id, err
}

// GetAssetIDByIP retrieves the asset ID for a given IP.
func (m *Manager) GetAssetIDByIP(ip string) (int64, error) {
	db := m.GetDB()
	var id int64
	err := db.QueryRow("SELECT id FROM assets WHERE ip = ?", ip).Scan(&id)
	return id, err
}

// GetAsset retrieves an asset by ID.
func (m *Manager) GetAsset(id int64) (*Asset, error) {
	db := m.GetDB()
	var a Asset
	err := db.QueryRow("SELECT id, ip, os, alive, last_scan_time, created_at FROM assets WHERE id = ?", id).Scan(
		&a.ID, &a.IP, &a.OS, &a.Alive, &a.LastScanTime, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// --- Ports ---

// UpsertAssetPort inserts or updates a port for an asset.
func (m *Manager) UpsertAssetPort(assetID int64, port int, protocol, service, product, version, banner, state string) (int64, error) {
	var id int64
	err := m.ExecTask(func(db *sql.DB) error {
		err := db.QueryRow("SELECT id FROM asset_ports WHERE asset_id = ? AND port = ? AND protocol = ?", assetID, port, protocol).Scan(&id)
		if err == sql.ErrNoRows {
			var res sql.Result
			res, err = db.Exec("INSERT INTO asset_ports (asset_id, port, protocol, service, product, version, banner, state) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				assetID, port, protocol, service, product, version, banner, state)
			if err != nil {
				return err
			}
			id, err = res.LastInsertId()
			return err
		} else if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE asset_ports SET service = ?, product = ?, version = ?, banner = ?, state = ?, updated_at = ? WHERE id = ?",
			service, product, version, banner, state, time.Now(), id)
		return err
	})
	return id, err
}

// GetAssetPorts retrieves ports for an asset.
func (m *Manager) GetAssetPorts(assetID int64) ([]AssetPort, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, asset_id, port, protocol, service, product, version, banner, state, updated_at FROM asset_ports WHERE asset_id = ? ORDER BY port ASC", assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []AssetPort
	for rows.Next() {
		var p AssetPort
		if err := rows.Scan(&p.ID, &p.AssetID, &p.Port, &p.Protocol, &p.Service, &p.Product, &p.Version, &p.Banner, &p.State, &p.UpdatedAt); err != nil {
			continue
		}
		ports = append(ports, p)
	}
	return ports, nil
}

// --- Web Services ---

// UpsertWebService inserts or updates a web service.
func (m *Manager) UpsertWebService(assetID, portID int64, url, title, server, fingerprints string) (int64, error) {
	var id int64
	err := m.ExecTask(func(db *sql.DB) error {
		err := db.QueryRow("SELECT id FROM web_services WHERE port_id = ? AND url = ?", portID, url).Scan(&id)
		if err == sql.ErrNoRows {
			var res sql.Result
			res, err = db.Exec("INSERT INTO web_services (asset_id, port_id, url, title, server, fingerprints) VALUES (?, ?, ?, ?, ?, ?)",
				assetID, portID, url, title, server, fingerprints)
			if err != nil {
				return err
			}
			id, err = res.LastInsertId()
			return err
		} else if err != nil {
			return err
		}

		_, err = db.Exec("UPDATE web_services SET title = ?, server = ?, fingerprints = ?, updated_at = ? WHERE id = ?",
			title, server, fingerprints, time.Now(), id)
		return err
	})
	return id, err
}

// GetWebServiceIDByURL retrieves the web service ID for a given URL.
func (m *Manager) GetWebServiceIDByURL(url string) (int64, error) {
	db := m.GetDB()
	var id int64
	// Simple match, might need normalization in real world
	err := db.QueryRow("SELECT id FROM web_services WHERE url = ?", url).Scan(&id)
	return id, err
}

// GetWebServices retrieves web services for an asset.
func (m *Manager) GetWebServices(assetID int64) ([]WebService, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, asset_id, port_id, url, title, server, fingerprints, updated_at FROM web_services WHERE asset_id = ?", assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []WebService
	for rows.Next() {
		var s WebService
		if err := rows.Scan(&s.ID, &s.AssetID, &s.PortID, &s.URL, &s.Title, &s.Server, &s.Fingerprints, &s.UpdatedAt); err != nil {
			continue
		}
		services = append(services, s)
	}
	return services, nil
}

// --- Directories ---

// AddWebDirectory adds a discovered directory.
func (m *Manager) AddWebDirectory(webServiceID int64, path string, status, length int, title, contentType, redirect string) error {
	return m.ExecTask(func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO web_directories (web_service_id, path, status_code, content_length, title, content_type, redirect_url) VALUES (?, ?, ?, ?, ?, ?, ?)",
			webServiceID, path, status, length, title, contentType, redirect)
		return err
	})
}

// GetWebDirectories retrieves directories for a web service.
func (m *Manager) GetWebDirectories(webServiceID int64) ([]WebDirectory, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, web_service_id, path, status_code, content_length, title, content_type, redirect_url, created_at FROM web_directories WHERE web_service_id = ? ORDER BY path ASC", webServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dirs []WebDirectory
	for rows.Next() {
		var d WebDirectory
		if err := rows.Scan(&d.ID, &d.WebServiceID, &d.Path, &d.StatusCode, &d.ContentLength, &d.Title, &d.ContentType, &d.RedirectURL, &d.CreatedAt); err != nil {
			continue
		}
		dirs = append(dirs, d)
	}
	return dirs, nil
}

// GetAssetWebDirectories retrieves all directories for an asset (across all web services).
func (m *Manager) GetAssetWebDirectories(assetID int64) ([]WebDirectory, error) {
	db := m.GetDB()
	query := `
		SELECT wd.id, wd.web_service_id, wd.path, wd.status_code, wd.content_length, wd.title, wd.content_type, wd.redirect_url, wd.created_at 
		FROM web_directories wd
		JOIN web_services ws ON wd.web_service_id = ws.id
		WHERE ws.asset_id = ?
		ORDER BY wd.path ASC
	`
	rows, err := db.Query(query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dirs []WebDirectory
	for rows.Next() {
		var d WebDirectory
		if err := rows.Scan(&d.ID, &d.WebServiceID, &d.Path, &d.StatusCode, &d.ContentLength, &d.Title, &d.ContentType, &d.RedirectURL, &d.CreatedAt); err != nil {
			continue
		}
		dirs = append(dirs, d)
	}
	return dirs, nil
}

// --- JS Files ---

// AddWebJSFile adds a discovered JS file.
func (m *Manager) AddWebJSFile(webServiceID int64, path, fullURL string) error {
	return m.ExecTask(func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO web_js_files (web_service_id, path, full_url) VALUES (?, ?, ?)",
			webServiceID, path, fullURL)
		return err
	})
}

// GetWebJSFiles retrieves JS files for a web service.
func (m *Manager) GetWebJSFiles(webServiceID int64) ([]WebJSFile, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, web_service_id, path, full_url, created_at FROM web_js_files WHERE web_service_id = ?", webServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []WebJSFile
	for rows.Next() {
		var f WebJSFile
		if err := rows.Scan(&f.ID, &f.WebServiceID, &f.Path, &f.FullURL, &f.CreatedAt); err != nil {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

// GetAssetWebJSFiles retrieves all JS files for an asset.
func (m *Manager) GetAssetWebJSFiles(assetID int64) ([]WebJSFile, error) {
	db := m.GetDB()
	query := `
		SELECT jf.id, jf.web_service_id, jf.path, jf.full_url, jf.created_at 
		FROM web_js_files jf
		JOIN web_services ws ON jf.web_service_id = ws.id
		WHERE ws.asset_id = ?
	`
	rows, err := db.Query(query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []WebJSFile
	for rows.Next() {
		var f WebJSFile
		if err := rows.Scan(&f.ID, &f.WebServiceID, &f.Path, &f.FullURL, &f.CreatedAt); err != nil {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

// --- Sensitive Info ---

// AddSensitiveResult adds a sensitive info finding.
func (m *Manager) AddSensitiveResult(webServiceID int64, sourceFile, infoType, content, context string) error {
	return m.ExecTask(func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO sensitive_results (web_service_id, source_file, info_type, content, context) VALUES (?, ?, ?, ?, ?)",
			webServiceID, sourceFile, infoType, content, context)
		return err
	})
}

// GetAssetSensitiveResults retrieves sensitive info for an asset.
func (m *Manager) GetAssetSensitiveResults(assetID int64) ([]SensitiveResult, error) {
	db := m.GetDB()
	query := `
		SELECT sr.id, sr.web_service_id, sr.source_file, sr.info_type, sr.content, sr.context, sr.created_at
		FROM sensitive_results sr
		JOIN web_services ws ON sr.web_service_id = ws.id
		WHERE ws.asset_id = ?
	`
	rows, err := db.Query(query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SensitiveResult
	for rows.Next() {
		var r SensitiveResult
		if err := rows.Scan(&r.ID, &r.WebServiceID, &r.SourceFile, &r.InfoType, &r.Content, &r.Context, &r.CreatedAt); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// --- Auth Results ---

// AddAuthResult adds a brute force result.
func (m *Manager) AddAuthResult(assetID, portID int64, serviceType, username, password string, success bool) error {
	return m.ExecTask(func(db *sql.DB) error {
		_, err := db.Exec("INSERT INTO auth_results (asset_id, port_id, service_type, username, password, success) VALUES (?, ?, ?, ?, ?, ?)",
			assetID, portID, serviceType, username, password, success)
		return err
	})
}

// GetAssetAuthResults retrieves auth results for an asset.
func (m *Manager) GetAssetAuthResults(assetID int64) ([]AuthResult, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, asset_id, port_id, service_type, username, password, success, created_at FROM auth_results WHERE asset_id = ?", assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AuthResult
	for rows.Next() {
		var r AuthResult
		if err := rows.Scan(&r.ID, &r.AssetID, &r.PortID, &r.ServiceType, &r.Username, &r.Password, &r.Success, &r.CreatedAt); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// --- Queries for UI ---

func (m *Manager) GetAllAssets() ([]Asset, error) {
	db := m.GetDB()
	rows, err := db.Query("SELECT id, ip, os, alive, last_scan_time, created_at FROM assets ORDER BY last_scan_time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.IP, &a.OS, &a.Alive, &a.LastScanTime, &a.CreatedAt); err != nil {
			continue
		}
		assets = append(assets, a)
	}
	return assets, nil
}
