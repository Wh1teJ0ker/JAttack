import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { FolderSearch, Play, StopCircle, FolderOpen } from "lucide-react";
import { toast } from "sonner";
import * as InfoService from "../../../wailsjs/go/infogather/InfoService";
import { infogather } from "../../../wailsjs/go/models";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";

export function DirScanPage() {
    const [target, setTarget] = useState("");
    const [extensions, setExtensions] = useState("php,jsp,asp,txt");
    const [concurrency, setConcurrency] = useState(50);
    const [timeout, setTimeout] = useState(5);
    const [customDict, setCustomDict] = useState("");
    const [isScanning, setIsScanning] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [results, setResults] = useState<infogather.DirScanResult[]>([]);

    useEffect(() => {
        const logHandler = (message: string) => {
            setLogs(prev => [...prev, message]);
        };
        const resultHandler = (result: infogather.DirScanResult) => {
            setResults(prev => [...prev, result]);
        };
        const completeHandler = () => {
            setIsScanning(false);
            toast.success("目录扫描任务完成");
        };

        EventsOn("scan:log", logHandler);
        EventsOn("dirScanResult", resultHandler);
        EventsOn("dirScanComplete", completeHandler);

        return () => {
            EventsOff("scan:log");
            EventsOff("dirScanResult");
            EventsOff("dirScanComplete");
        };
    }, []);

    const handleSelectDict = async () => {
        try {
            const path = await InfoService.SelectDictionaryFile();
            if (path) {
                setCustomDict(path);
            }
        } catch (err) {
            toast.error("选择字典失败: " + err);
        }
    };

    const handleStartScan = async () => {
        if (!target) {
            toast.error("请输入目标 URL");
            return;
        }

        setIsScanning(true);
        setLogs([]);
        setResults([]);
        
        try {
            const config = new infogather.DirScanConfig({
                target: target,
                extensions: extensions.split(","),
                threads: concurrency,
                timeout: timeout,
                exclude_404: true,
                redirects: true,
                custom_dict: customDict,
                recursion_depth: 0,
                enable_fingerprint: false
            });

            await InfoService.StartDirScan(config);
            toast.success("目录扫描任务已下发");
        } catch (err) {
            toast.error("任务启动失败: " + err);
            setIsScanning(false);
        }
    };

    const handleStopScan = async () => {
        // Note: Check if backend supports stopping dir scan specifically or global StopScan
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
                    <FolderSearch className="h-6 w-6 text-primary" />
                    目录扫描
                </h1>
                <p className="text-muted-foreground">
                    高效枚举 Web 目录与敏感文件，发现隐藏的后台及配置泄漏。
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1">
                <div className="lg:col-span-1 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>扫描配置</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="target">目标 URL</Label>
                                <Input 
                                    id="target" 
                                    placeholder="例如: http://example.com" 
                                    value={target}
                                    onChange={(e) => setTarget(e.target.value)}
                                />
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="exts">文件后缀 (逗号分隔)</Label>
                                    <Input 
                                        id="exts" 
                                        value={extensions}
                                        onChange={(e) => setExtensions(e.target.value)}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>自定义字典 (可选)</Label>
                                    <div className="flex space-x-2">
                                        <Input 
                                            placeholder="默认使用内置字典" 
                                            value={customDict}
                                            onChange={(e) => setCustomDict(e.target.value)}
                                        />
                                        <Button variant="outline" size="icon" onClick={handleSelectDict}>
                                            <FolderOpen className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="concurrency">并发数</Label>
                                    <Input 
                                        id="concurrency" 
                                        type="number"
                                        min={1}
                                        max={500}
                                        value={concurrency}
                                        onChange={(e) => setConcurrency(parseInt(e.target.value) || 50)}
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
                                        <div key={i} className="flex items-center gap-3 border-b border-border pb-1 last:border-0">
                                            <span className={`px-2 py-0.5 rounded text-xs font-bold ${
                                                res.status >= 200 && res.status < 300 ? "bg-green-100 text-green-700" :
                                                res.status >= 300 && res.status < 400 ? "bg-blue-100 text-blue-700" :
                                                res.status >= 400 && res.status < 500 ? "bg-yellow-100 text-yellow-700" :
                                                "bg-red-100 text-red-700"
                                            }`}>
                                                {res.status}
                                            </span>
                                            <span className="flex-1 truncate" title={res.url}>{res.url}</span>
                                            <span className="text-muted-foreground text-xs whitespace-nowrap">{res.size} B</span>
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
