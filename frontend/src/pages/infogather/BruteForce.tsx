import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { Shield, Play, StopCircle, FolderOpen } from "lucide-react";
import { toast } from "sonner";
import * as BruteForceService from "../../../wailsjs/go/infogather/BruteForceService";
import { infogather } from "../../../wailsjs/go/models";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";

export function BruteForcePage() {
    const [target, setTarget] = useState("");
    const [service, setService] = useState("ssh");
    const [userDict, setUserDict] = useState("");
    const [passDict, setPassDict] = useState("");
    const [threads, setThreads] = useState(10);
    const [timeout, setTimeout] = useState(5);
    const [isRunning, setIsRunning] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);

    useEffect(() => {
        const logHandler = (message: string) => {
            setLogs(prev => [...prev, message]);
        };
        const finishedHandler = () => {
            setIsRunning(false);
            toast.success("爆破任务完成");
        };

        EventsOn("bruteforce:log", logHandler);
        EventsOn("bruteforce:finished", finishedHandler);

        return () => {
            EventsOff("bruteforce:log");
            EventsOff("bruteforce:finished");
        };
    }, []);

    const handleSelectUserDict = async () => {
        try {
            const path = await BruteForceService.SelectDictionaryFile();
            if (path) {
                setUserDict(path);
            }
        } catch (err) {
            toast.error("选择字典失败: " + err);
        }
    };

    const handleSelectPassDict = async () => {
        try {
            const path = await BruteForceService.SelectDictionaryFile();
            if (path) {
                setPassDict(path);
            }
        } catch (err) {
            toast.error("选择字典失败: " + err);
        }
    };

    const handleStart = async () => {
        if (!target) {
            toast.error("请输入目标地址 (IP:Port)");
            return;
        }

        setIsRunning(true);
        setLogs([]);
        
        try {
            // Parse target IP and Port
            const parts = target.split(":");
            if (parts.length !== 2) {
                toast.error("格式错误，请使用 IP:Port 格式");
                setIsRunning(false);
                return;
            }

            const ip = parts[0];
            const port = parseInt(parts[1]);

            const config = new infogather.BruteForceConfig({
                user_dict: userDict,
                pass_dict: passDict,
                threads: threads,
                timeout: timeout * 1000, // Convert to ms
                protocols: [service],
                targets: [
                    new infogather.BruteForceTarget({
                        ip: ip,
                        port: port,
                        protocol: service,
                        service: service
                    })
                ]
            });

            await BruteForceService.StartAttack(config);
            toast.success("爆破任务已启动");
        } catch (err) {
            toast.error("启动失败: " + err);
            setIsRunning(false);
        }
    };

    const handleStop = async () => {
        try {
            await BruteForceService.StopAttack();
            setIsRunning(false);
            toast.info("已停止任务");
        } catch (err) {
            toast.error("停止失败: " + err);
        }
    };

    return (
        <div className="h-full p-6 space-y-6 flex flex-col">
            <div className="flex flex-col gap-2">
                <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                    <Shield className="h-6 w-6 text-primary" />
                    弱口令爆破
                </h1>
                <p className="text-muted-foreground">
                    支持 SSH, FTP, MySQL, Redis 等常见服务的弱口令检测。
                </p>
            </div>

            <div className="flex flex-col lg:flex-row gap-6 flex-1 min-h-0">
                <div className="w-full lg:w-[400px] flex-shrink-0 flex flex-col gap-6">
                    <Card className="flex-1 overflow-y-auto">
                        <CardHeader>
                            <CardTitle>任务配置</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Target Section */}
                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <Label htmlFor="target">目标服务 (IP:Port)</Label>
                                    <Input 
                                        id="target" 
                                        placeholder="例如: 192.168.1.10:22" 
                                        value={target}
                                        onChange={(e) => setTarget(e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="service">服务类型</Label>
                                    <select 
                                        id="service" 
                                        className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                        value={service}
                                        onChange={(e) => setService(e.target.value)}
                                    >
                                        <option value="ssh">SSH</option>
                                        <option value="ftp">FTP</option>
                                        <option value="mysql">MySQL</option>
                                        <option value="redis">Redis</option>
                                        <option value="mssql">MSSQL</option>
                                        <option value="postgres">PostgreSQL</option>
                                        <option value="rdp">RDP</option>
                                        <option value="smb">SMB</option>
                                        <option value="telnet">Telnet</option>
                                    </select>
                                </div>
                            </div>

                            <div className="h-px bg-border/50" />

                            {/* Dictionary Section */}
                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <Label>用户名字典 (可选)</Label>
                                    <div className="flex space-x-2">
                                        <Input 
                                            placeholder="默认内置字典" 
                                            value={userDict}
                                            onChange={(e) => setUserDict(e.target.value)}
                                        />
                                        <Button variant="outline" size="icon" onClick={handleSelectUserDict}>
                                            <FolderOpen className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <Label>密码字典 (可选)</Label>
                                    <div className="flex space-x-2">
                                        <Input 
                                            placeholder="默认内置字典" 
                                            value={passDict}
                                            onChange={(e) => setPassDict(e.target.value)}
                                        />
                                        <Button variant="outline" size="icon" onClick={handleSelectPassDict}>
                                            <FolderOpen className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            </div>

                            <div className="h-px bg-border/50" />

                            {/* Performance Section */}
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="threads">并发线程</Label>
                                    <Input 
                                        id="threads" 
                                        type="number"
                                        min={1}
                                        max={100}
                                        value={threads}
                                        onChange={(e) => setThreads(parseInt(e.target.value) || 10)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="timeout">超时 (秒)</Label>
                                    <Input 
                                        id="timeout" 
                                        type="number"
                                        min={1}
                                        value={timeout}
                                        onChange={(e) => setTimeout(parseInt(e.target.value) || 5)}
                                    />
                                </div>
                            </div>
                            
                            <div className="pt-2">
                                {isRunning ? (
                                    <Button variant="destructive" onClick={handleStop} className="w-full">
                                        <StopCircle className="mr-2 h-4 w-4" /> 停止任务
                                    </Button>
                                ) : (
                                    <Button onClick={handleStart} className="w-full">
                                        <Play className="mr-2 h-4 w-4" /> 开始爆破
                                    </Button>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <div className="flex-1 flex flex-col min-w-0">
                    <Card className="flex-1 flex flex-col min-h-0">
                        <CardHeader>
                            <CardTitle>爆破日志</CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-auto bg-muted/30 p-4 mx-6 mb-6 rounded text-xs font-mono">
                            {logs.length === 0 ? (
                                <div className="text-center text-muted-foreground py-10">
                                    {isRunning ? "正在运行中..." : "暂无日志"}
                                </div>
                            ) : (
                                logs.map((log, i) => (
                                    <div key={i} className={`whitespace-pre-wrap break-all ${log.includes("[SUCCESS]") ? "text-green-600 font-bold" : "text-muted-foreground"}`}>
                                        {log}
                                    </div>
                                ))
                            )}
                            <div id="log-end" />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
