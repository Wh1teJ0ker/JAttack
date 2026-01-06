import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/Card";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { FileCode, Play, StopCircle } from "lucide-react";
import { toast } from "sonner";
import * as JSFinderService from "../../../wailsjs/go/infogather/JSFinderService";
import { infogather } from "../../../wailsjs/go/models";

export function JSFinderPage() {
    const [url, setUrl] = useState("");
    const [deepScan, setDeepScan] = useState(false);
    const [concurrency, setConcurrency] = useState(10);
    const [timeout, setTimeout] = useState(10);
    const [isScanning, setIsScanning] = useState(false);
    const [results, setResults] = useState<infogather.JSFindResult | null>(null);

    const handleStartScan = async () => {
        if (!url) {
            toast.error("请输入目标 URL");
            return;
        }

        setIsScanning(true);
        setResults(null);
        
        try {
            const options = new infogather.JSFinderOptions({
                deep_scan: deepScan,
                active_scan: false,
                danger_filter: false,
                concurrency: concurrency,
                timeout: timeout * 1000 // Convert to ms
            });

            // Note: JSFinderService.Run might return results directly or be async
            // Assuming it returns results based on models
            const res = await JSFinderService.FindJS(url, options);
            setResults(res);
            toast.success("JS 分析完成");
        } catch (err) {
            toast.error("分析失败: " + err);
        } finally {
            setIsScanning(false);
        }
    };

    return (
        <div className="h-full p-6 space-y-6 flex flex-col">
            <div className="flex flex-col gap-2">
                <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                    <FileCode className="h-6 w-6 text-primary" />
                    JS 敏感提取
                </h1>
                <p className="text-muted-foreground">
                    爬取并分析 JavaScript 文件，提取 API 接口、子域名及敏感密钥信息。
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>任务配置</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="url">目标 URL</Label>
                        <Input 
                            id="url" 
                            placeholder="例如: http://example.com" 
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                        />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="concurrency">并发数</Label>
                            <Input 
                                id="concurrency" 
                                type="number"
                                min={1}
                                max={100}
                                value={concurrency}
                                onChange={(e) => setConcurrency(parseInt(e.target.value) || 10)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="timeout">超时 (秒)</Label>
                            <Input 
                                id="timeout" 
                                type="number"
                                min={1}
                                value={timeout}
                                onChange={(e) => setTimeout(parseInt(e.target.value) || 10)}
                            />
                        </div>
                    </div>

                    <div className="flex items-center space-x-2">
                        <input 
                            type="checkbox"
                            id="deep" 
                            checked={deepScan} 
                            onChange={(e) => setDeepScan(e.target.checked)} 
                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <Label htmlFor="deep">启用深度扫描 (递归分析)</Label>
                    </div>
                    
                    <div className="pt-2">
                        <Button onClick={handleStartScan} disabled={isScanning}>
                            {isScanning ? <span className="animate-spin mr-2">⏳</span> : <Play className="mr-2 h-4 w-4" />}
                            开始分析
                        </Button>
                    </div>
                </CardContent>
            </Card>

            <Card className="flex-1">
                <CardHeader>
                    <CardTitle>分析结果</CardTitle>
                </CardHeader>
                <CardContent>
                    {!results && !isScanning && (
                        <div className="text-center text-muted-foreground py-10">
                            暂无结果，请开始分析
                        </div>
                    )}
                    
                    {results && (
                        <div className="space-y-6">
                            <div>
                                <h3 className="text-lg font-medium mb-2">Endpoints ({results.endpoints?.length || 0})</h3>
                                <div className="bg-muted p-4 rounded-md max-h-40 overflow-y-auto font-mono text-sm">
                                    {results.endpoints?.map((e, i) => (
                                        <div key={i}>{e}</div>
                                    ))}
                                </div>
                            </div>
                            
                            <div>
                                <h3 className="text-lg font-medium mb-2">Sensitive Info ({results.sensitive_info?.length || 0})</h3>
                                <div className="bg-muted p-4 rounded-md max-h-40 overflow-y-auto font-mono text-sm text-red-400">
                                    {results.sensitive_info?.map((e, i) => (
                                        <div key={i} className="break-all">{e}</div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
