import { useState } from 'react';
import { Shield, Settings, Globe, Github, FileText, ChevronDown, ChevronRight, Activity, Network, FolderSearch, FileCode } from 'lucide-react';
import { NavLink, useLocation } from 'react-router-dom';
import { cn } from '../../lib/utils';
import { Button } from '../ui/Button';

interface NavItem {
    icon: any;
    label: string;
    path: string;
    children?: { label: string; path: string }[];
}

const navItems: NavItem[] = [
    {
        icon: Globe,
        label: '信息搜集',
        path: '/info-gathering',
        children: [
            { label: '存活探测', path: '/info-gathering/alive' },
            { label: '端口扫描', path: '/info-gathering/portscan' },
            { label: '目录扫描', path: '/info-gathering/dirscan' },
            { label: 'JS 敏感提取', path: '/info-gathering/jsfinder' },
            { label: '弱口令爆破', path: '/info-gathering/bruteforce' },
        ]
    },
    { icon: Shield, label: '漏洞管理', path: '/vuln-manager' },
    { icon: FileText, label: '日志信息', path: '/logs' },
    { icon: Settings, label: '设置', path: '/settings' },
];

export function Sidebar() {
    const location = useLocation();
    const [expanded, setExpanded] = useState<Record<string, boolean>>({
        '/info-gathering': true
    });

    const toggleExpand = (path: string) => {
        setExpanded(prev => ({ ...prev, [path]: !prev[path] }));
    };

    return (
        <aside className="w-64 h-screen bg-card text-card-foreground fixed left-0 top-0 flex flex-col border-r border-border">
            <div className="p-6 border-b border-border">
                <h1 className="text-xl font-bold flex items-center gap-2">
                    <span className="text-primary">Joker</span>Attack
                </h1>
            </div>
            
            <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
                {navItems.map((item) => {
                    if (item.children) {
                        const isExpanded = expanded[item.path];
                        const isActive = location.pathname.startsWith(item.path);
                        
                        return (
                            <div key={item.path} className="space-y-1">
                                <button
                                    onClick={() => toggleExpand(item.path)}
                                    className={cn(
                                        "w-full flex items-center justify-between px-4 py-3 rounded-md transition-colors duration-200 text-sm font-medium",
                                        "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                        isActive && "text-foreground"
                                    )}
                                >
                                    <div className="flex items-center gap-3">
                                        <item.icon size={18} />
                                        <span>{item.label}</span>
                                    </div>
                                    {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                                </button>
                                
                                {isExpanded && (
                                    <div className="pl-11 space-y-1">
                                        {item.children.map(child => (
                                            <NavLink
                                                key={child.path}
                                                to={child.path}
                                                className={({ isActive }) =>
                                                    cn(
                                                        "block px-4 py-2 rounded-md transition-colors duration-200 text-sm",
                                                        "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                                        isActive && "bg-secondary text-secondary-foreground"
                                                    )
                                                }
                                            >
                                                {child.label}
                                            </NavLink>
                                        ))}
                                    </div>
                                )}
                            </div>
                        );
                    }

                    return (
                        <NavLink
                            key={item.path}
                            to={item.path}
                            className={({ isActive }) =>
                                cn(
                                    "flex items-center gap-3 px-4 py-3 rounded-md transition-colors duration-200 text-sm font-medium",
                                    "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                    isActive && "bg-secondary text-secondary-foreground"
                                )
                            }
                        >
                            <item.icon size={18} />
                            <span>{item.label}</span>
                        </NavLink>
                    );
                })}
            </nav>

            <div className="p-4 border-t border-border space-y-4">
                <Button variant="outline" className="w-full justify-start gap-2" size="sm">
                    <Github size={16} />
                    <span>GitHub</span>
                </Button>
                <div className="bg-muted/50 rounded-md p-3">
                    <p className="text-xs text-muted-foreground font-mono">v1.0.0</p>
                </div>
            </div>
        </aside>
    );
}
