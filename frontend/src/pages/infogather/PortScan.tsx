import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { Network, Play, StopCircle } from "lucide-react";
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

export function PortScanPage() {
    const [target, setTarget] = useState("");
    const [portMode, setPortMode] = useState("all");
    const [customPorts, setCustomPorts] = useState("");
    const [concurrency, setConcurrency] = useState(500);
    const [timeout, setTimeout] = useState(800);
    const [isScanning, setIsScanning] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [results, setResults] = useState<ScanResult[]>([]);

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
            toast.error("请输入目标 IP");
            return;
        }

        let ports = "";
        if (portMode === "all") ports = "1-65535";
        else ports = customPorts;

        setIsScanning(true);
        setLogs([]);
        setResults([]);
        
        try {
            const config = new infogather.ScanConfig({
                target: target,
                ports: ports,
                concurrency: concurrency,
                timeout: timeout,
                skip_alive_check: false,
                enable_icmp: true,
                enable_ping: true,
                enable_fingerprint: false, // Fingerprint is now a separate step
                enable_udp: false
            });

            await InfoService.StartScan(config);
            toast.success("端口扫描任务已下发");
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
                    <Network className="h-6 w-6 text-primary" />
                    端口扫描
                </h1>
                <p className="text-muted-foreground">
                    全面扫描目标开放端口，识别服务版本及潜在风险。
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1">
                <div className="lg:col-span-1 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>扫描配置</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="space-y-2">
                                <Label htmlFor="target">目标主机 (IP / Domain)</Label>
                                <Input 
                                    id="target" 
                                    placeholder="例如: 192.168.1.10, example.com" 
                                    value={target}
                                    onChange={(e) => setTarget(e.target.value)}
                                />
                            </div>

                            <div className="space-y-3">
                                <Label>端口范围</Label>
                                <div className="flex flex-col space-y-2">
                                    <div className="flex items-center space-x-2">
                                        <input 
                                            type="radio" 
                                            id="all" 
                                            name="portMode"
                                            value="all"
                                            checked={portMode === "all"}
                                            onChange={(e) => setPortMode(e.target.value)}
                                            className="h-4 w-4 border-gray-300 text-primary focus:ring-primary"
                                        />
                                        <Label htmlFor="all">全端口 (1-65535)</Label>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        <input 
                                            type="radio" 
                                            id="custom" 
                                            name="portMode"
                                            value="custom"
                                            checked={portMode === "custom"}
                                            onChange={(e) => setPortMode(e.target.value)}
                                            className="h-4 w-4 border-gray-300 text-primary focus:ring-primary"
                                        />
                                        <Label htmlFor="custom">自定义端口</Label>
                                    </div>
                                </div>
                                {portMode === "custom" && (
                                    <Input 
                                        placeholder="例如: 80,443,8000-9000" 
                                        value={customPorts}
                                        onChange={(e) => setCustomPorts(e.target.value)}
                                        className="mt-2"
                                    />
                                )}
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="concurrency">并发数</Label>
                                    <Input 
                                        id="concurrency" 
                                        type="number"
                                        min={1}
                                        max={5000}
                                        value={concurrency}
                                        onChange={(e) => setConcurrency(parseInt(e.target.value) || 500)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="timeout">超时 (毫秒)</Label>
                                    <Input 
                                        id="timeout" 
                                        type="number"
                                        min={1}
                                        value={timeout}
                                        onChange={(e) => setTimeout(parseInt(e.target.value) || 800)}
                                    />
                                </div>
                            </div>

                            <div className="pt-2">
                                {isScanning ? (
                                    <Button variant="destructive" onClick={handleStopScan} className="w-full">
                                        <StopCircle className="mr-2 h-4 w-4" /> 停止扫描
                                    </Button>
                                ) : (
                                    <Button onClick={handleStartScan} className="w-full">
                                        <Play className="mr-2 h-4 w-4" /> 开始扫描
                                    </Button>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <div className="lg:col-span-2 flex flex-col space-y-6 h-full">
                    <Card className="flex-1 flex flex-col">
                        <CardHeader>
                            <CardTitle>扫描结果</CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-auto">
                            {results.length === 0 ? (
                                <div className="text-center text-muted-foreground py-10">
                                    {isScanning ? (
                                        <div className="space-y-2">
                                            <p>正在扫描中...</p>
                                            <div className="text-sm text-muted-foreground/80 font-mono">
                                                {logs.length > 0 && logs[logs.length - 1]}
                                            </div>
                                        </div>
                                    ) : "暂无结果，请开始扫描"}
                                </div>
                            ) : (
                                <div className="space-y-2 font-mono text-sm">
                                    {results.map((res, i) => (
                                        <div key={i} className="flex gap-2 border-b border-border pb-1 last:border-0">
                                            <span className="text-muted-foreground">[{res.time}]</span>
                                            <span className="font-semibold text-primary">{res.target}</span>
                                            <span className="text-blue-500">{res.info_type}</span>
                                            <span>{res.content}</span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card className="h-1/3 flex flex-col">
                        <CardHeader className="py-3">
                            <CardTitle className="text-base">运行日志</CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-auto bg-muted/30 p-2 mx-6 mb-6 rounded text-xs font-mono">
                            {logs.map((log, i) => (
                                <div key={i}>{log}</div>
                            ))}
                            <div id="log-end" />
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
