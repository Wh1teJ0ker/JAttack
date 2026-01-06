import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { Badge } from "../../components/ui/Badge";
import { Activity, Play, StopCircle, Laptop } from "lucide-react";
import { toast } from "sonner";
import * as InfoService from "../../../wailsjs/go/infogather/InfoService";
import { infogather } from "../../../wailsjs/go/models";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";

interface ScanResult {
    target: string;
    info_type: string;
    content: string;
    time: string;
}

export function AliveScanPage() {
    const [target, setTarget] = useState("");
    const [isScanning, setIsScanning] = useState(false);
    const [results, setResults] = useState<ScanResult[]>([]);
    const [logs, setLogs] = useState<string[]>([]);

    useEffect(() => {
        const logHandler = (message: string) => {
            setLogs(prev => [...prev, message]);
        };
        const resultHandler = (result: ScanResult) => {
            setResults(prev => [...prev, result]);
        };
        const completeHandler = () => {
            setIsScanning(false);
            toast.success("扫描任务完成");
        };

        EventsOn("scan:log", logHandler);
        EventsOn("scan:result", resultHandler);
        EventsOn("scan:complete", completeHandler);

        return () => {
            EventsOff("scan:log");
            EventsOff("scan:result");
            EventsOff("scan:complete");
        };
    }, []);

    const handleStartScan = async () => {
        if (!target) {
            toast.error("请输入目标 IP 或网段");
            return;
        }

        setIsScanning(true);
        setResults([]);
        setLogs([]);
        
        try {
            const config = new infogather.ScanConfig({
                target: target,
                ports: "", // No ports for alive scan
                concurrency: 100,
                timeout: 1000,
                skip_alive_check: false,
                enable_icmp: true,
                enable_ping: true,
                enable_fingerprint: false,
                enable_udp: false
            });

            await InfoService.StartScan(config);
            toast.success("存活探测任务已下发");
        } catch (err) {
            toast.error("任务启动失败: " + err);
            setIsScanning(false);
        }
    };

    const handleStopScan = async () => {
        try {
            await InfoService.StopScan();
            setIsScanning(false);
            toast.info("已停止扫描");
        } catch (err) {
            toast.error("停止失败: " + err);
        }
    };

    return (
        <div className="h-full p-6 space-y-6 flex flex-col">
            <div className="flex flex-col gap-2">
                <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                    <Activity className="h-6 w-6 text-primary" />
                    存活探测
                </h1>
                <p className="text-muted-foreground">
                    基于 ICMP 和 Ping 的主机存活发现，快速识别网段内的在线资产。
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>扫描配置</CardTitle>
                    <CardDescription>设置目标网段进行批量探测</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex gap-4 items-end">
                        <div className="flex-1 space-y-2">
                            <Label htmlFor="target">目标网段 (CIDR / IP Range)</Label>
                            <Input 
                                id="target" 
                                placeholder="例如: 192.168.1.0/24, 10.0.0.1-100" 
                                value={target}
                                onChange={(e) => setTarget(e.target.value)}
                            />
                        </div>
                        <div className="flex gap-2">
                            {isScanning ? (
                                <Button variant="destructive" onClick={handleStopScan}>
                                    <StopCircle className="mr-2 h-4 w-4" /> 停止
                                </Button>
                            ) : (
                                <Button onClick={handleStartScan}>
                                    <Play className="mr-2 h-4 w-4" /> 开始探测
                                </Button>
                            )}
                        </div>
                    </div>
                </CardContent>
            </Card>

            <div className="flex gap-6 flex-1 min-h-0">
                 {/* Left: Results */}
                <Card className="flex-[2] flex flex-col min-h-0">
                    <CardHeader>
                        <CardTitle className="flex items-center justify-between">
                            <span>探测结果</span>
                            <Badge variant="secondary">{results.length} 个存活</Badge>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 overflow-auto">
                        {results.length === 0 ? (
                            <div className="text-center text-muted-foreground py-10">
                                {isScanning ? "正在扫描中..." : "暂无结果，请开始扫描"}
                            </div>
                        ) : (
                            <div className="space-y-2">
                                {results.map((res, index) => (
                                    <div key={index} className="flex items-center justify-between p-3 border rounded-md bg-card hover:bg-accent/50 transition-colors">
                                        <div className="flex items-center gap-3">
                                            <Laptop className="h-5 w-5 text-green-500" />
                                            <div className="font-mono text-base">{res.target}</div>
                                        </div>
                                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                                            <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50">Alive</Badge>
                                            <span>{res.time}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* Right: Logs */}
                <Card className="flex-1 flex flex-col min-h-0">
                    <CardHeader>
                        <CardTitle>扫描日志</CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 overflow-auto font-mono text-xs p-0">
                         <div className="p-4 space-y-1">
                            {logs.map((log, i) => (
                                <div key={i} className="text-muted-foreground border-b border-border/50 last:border-0 pb-1 mb-1">
                                    <span className="text-primary mr-2">[{new Date().toLocaleTimeString()}]</span>
                                    {log}
                                </div>
                            ))}
                            {isScanning && (
                                <div className="animate-pulse text-primary">_</div>
                            )}
                         </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
