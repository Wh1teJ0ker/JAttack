import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { MainLayout } from './components/layout/MainLayout';

// Modules
import { AliveScanPage } from './pages/infogather/AliveScan';
import { PortScanPage } from './pages/infogather/PortScan';
import { DirScanPage } from './pages/infogather/DirScan';
import { JSFinderPage } from './pages/infogather/JSFinder';
import { BruteForcePage } from './pages/infogather/BruteForce';

import VulnerabilityManagerPage from './pages/vuln/VulnManager';
import { SettingsPage } from './pages/settings/Settings';
import LogPage from './pages/logs/SystemLogs';

function App() {
    return (
        <HashRouter basename="/">
            <Routes>
                <Route path="/" element={<MainLayout />}>
                    <Route index element={<Navigate to="/info-gathering/alive" replace />} />
                    
                    {/* Info Gathering Routes */}
                    <Route path="info-gathering">
                        <Route index element={<Navigate to="alive" replace />} />
                        <Route path="alive" element={<AliveScanPage />} />
                        <Route path="portscan" element={<PortScanPage />} />
                        <Route path="dirscan" element={<DirScanPage />} />
                        <Route path="jsfinder" element={<JSFinderPage />} />
                        <Route path="bruteforce" element={<BruteForcePage />} />
                    </Route>
                    
                    <Route path="vuln-manager" element={<VulnerabilityManagerPage />} />
                    <Route path="logs" element={<LogPage />} />
                    <Route path="settings" element={<SettingsPage />} />
                    
                    {/* Fallback */}
                    <Route path="*" element={<Navigate to="/info-gathering/alive" replace />} />
                </Route>
            </Routes>
        </HashRouter>
    );
}

export default App;
