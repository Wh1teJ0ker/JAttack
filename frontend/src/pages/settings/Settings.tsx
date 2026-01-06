import { useState } from "react";
import { PageContainer } from "../../components/layout/PageContainer";
import { PageHeader } from "../../components/layout/PageHeader";
import { Button } from "../../components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../../components/ui/Card";
import { Input } from "../../components/ui/Input";
import { InitDatabase } from "../../../wailsjs/go/settings/SettingsService";
import { Database, FolderOpen, CheckCircle, AlertCircle } from "lucide-react";

export function SettingsPage() {
    const [dbPath, setDbPath] = useState("runtime/data/jattack.db");
    const [status, setStatus] = useState<{ type: 'success' | 'error' | '', msg: string }>({ type: '', msg: '' });
    const [loading, setLoading] = useState(false);

    const handleInit = async () => {
        if (!dbPath.trim()) {
            setStatus({ type: 'error', msg: '数据库路径不能为空' });
            return;
        }

        setLoading(true);
        setStatus({ type: '', msg: '' });

        try {
            await InitDatabase(dbPath);
            setStatus({ type: 'success', msg: '数据库初始化成功！' });
        } catch (e: any) {
            setStatus({ type: 'error', msg: `初始化失败: ${e}` });
        } finally {
            setLoading(false);
        }
    };

    return (
        <PageContainer>
            <PageHeader 
                title="基础设置" 
                description="管理系统核心配置，包括数据库连接和运行时参数。" 
            />
            
            <div className="grid gap-6">
                <Card className="border-l-4 border-l-blue-500">
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <Database className="w-5 h-5 text-blue-500" />
                            <CardTitle>数据库配置</CardTitle>
                        </div>
                        <CardDescription>
                            配置系统使用的 SQLite 数据库文件路径。默认路径为 <code className="bg-slate-800 px-1 py-0.5 rounded text-xs">runtime/data/jattack.db</code>。
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-slate-300">数据库文件路径</label>
                            <div className="flex gap-3">
                                <div className="relative flex-1">
                                    <FolderOpen className="absolute left-3 top-2.5 w-4 h-4 text-slate-500" />
                                    <Input
                                        value={dbPath}
                                        onChange={(e) => setDbPath(e.target.value)}
                                        placeholder="例如: runtime/data/jattack.db"
                                        className="pl-9 font-mono text-sm"
                                    />
                                </div>
                                <Button 
                                    onClick={handleInit} 
                                    disabled={loading}
                                    className="min-w-[100px]"
                                >
                                    {loading ? '正在初始化...' : '初始化数据库'}
                                </Button>
                            </div>
                        </div>

                        {/* Status Message */}
                        {status.msg && (
                            <div className={`flex items-center gap-2 p-3 rounded-md text-sm ${
                                status.type === 'success' 
                                    ? 'bg-green-500/10 text-green-400 border border-green-500/20' 
                                    : 'bg-red-500/10 text-red-400 border border-red-500/20'
                            }`}>
                                {status.type === 'success' ? (
                                    <CheckCircle className="w-4 h-4 shrink-0" />
                                ) : (
                                    <AlertCircle className="w-4 h-4 shrink-0" />
                                )}
                                <span>{status.msg}</span>
                            </div>
                        )}
                        
                        <div className="text-xs text-slate-500 mt-2">
                            <p>提示：如果指定路径的文件不存在，系统将自动创建新的数据库文件。</p>
                        </div>
                    </CardContent>
                </Card>

                {/* Future settings sections can go here */}
                {/* 
                <Card>
                    <CardHeader>
                        <CardTitle>其他设置</CardTitle>
                    </CardHeader>
                    <CardContent>
                        ...
                    </CardContent>
                </Card> 
                */}
            </div>
        </PageContainer>
    );
}
