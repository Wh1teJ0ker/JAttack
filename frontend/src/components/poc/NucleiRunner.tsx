import { useState } from 'react';
import { RunNuclei } from '../../../wailsjs/go/poc/PocService';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { CodeEditor } from '../ui/CodeEditor';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/Card';
import { Play } from 'lucide-react';
import { toast } from 'sonner';

export function NucleiRunner() {
    const [target, setTarget] = useState('');
    const [template, setTemplate] = useState('id: basic-check\ninfo:\n  name: Basic Check\n  author: jattack\n  severity: info\n\nrequests:\n  - method: GET\n    path:\n      - "{{BaseURL}}"\n    matchers:\n      - type: status\n        status:\n          - 200');
    const [output, setOutput] = useState('');
    const [running, setRunning] = useState(false);

    const handleRun = async () => {
        if (!target) {
            toast.error("请输入目标 URL");
            return;
        }
        setRunning(true);
        try {
            const res = await RunNuclei(target, template);
            setOutput(res);
            toast.success("执行完成");
        } catch (e) {
            setOutput(`Error: ${e}`);
            toast.error("执行出错");
        } finally {
            setRunning(false);
        }
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 h-full">
            <div className="flex flex-col gap-4 h-full">
                <Card>
                    <CardHeader className="py-3">
                        <CardTitle className="text-base">目标配置</CardTitle>
                    </CardHeader>
                    <CardContent className="py-4">
                        <Input 
                            placeholder="目标 URL (e.g., http://example.com)" 
                            value={target}
                            onChange={(e) => setTarget(e.target.value)}
                        />
                    </CardContent>
                </Card>

                <Card className="flex-1 flex flex-col">
                    <CardHeader className="py-3">
                        <CardTitle className="text-base flex justify-between items-center">
                            Nuclei 模板
                            <Button size="sm" onClick={handleRun} disabled={running}>
                                <Play className="mr-2 h-4 w-4" /> {running ? '扫描中...' : '开始验证'}
                            </Button>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 p-0 overflow-hidden">
                        <CodeEditor 
                            value={template} 
                            onChange={(val) => setTemplate(val || "")}
                            language="yaml"
                        />
                    </CardContent>
                </Card>
            </div>
            
            <Card className="flex flex-col h-full">
                <CardHeader className="py-3">
                    <CardTitle className="text-base">执行结果</CardTitle>
                </CardHeader>
                <CardContent className="flex-1 bg-black/90 text-green-400 p-4 font-mono text-sm overflow-auto rounded-b-lg">
                    <pre className="whitespace-pre-wrap break-all">
                        {output || "等待执行..."}
                    </pre>
                </CardContent>
            </Card>
        </div>
    );
}
