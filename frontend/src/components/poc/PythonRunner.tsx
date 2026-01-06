import { useState, useEffect } from 'react';
import { RunPython, GetPythonPath, SetPythonPath } from '../../../wailsjs/go/poc/PocService';
import { Button } from '../ui/Button';
import { CodeEditor } from '../ui/CodeEditor';
import { Input } from '../ui/Input';
import { Label } from '../ui/Label';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/Card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/Dialog";
import { Play, Settings } from 'lucide-react';
import { toast } from 'sonner';

export function PythonRunner() {
    const [code, setCode] = useState('print("Hello from JAttack Python Sandbox")\nimport sys\nprint(sys.version)');
    const [output, setOutput] = useState('');
    const [running, setRunning] = useState(false);
    const [pythonPath, setPythonPathState] = useState('');
    const [isSettingsOpen, setIsSettingsOpen] = useState(false);

    useEffect(() => {
        // Load current python path
        loadPythonPath();
    }, []);

    const loadPythonPath = async () => {
        try {
            const path = await GetPythonPath();
            setPythonPathState(path);
        } catch (e) {
            console.error("Failed to get python path", e);
            // Don't show error toast on load, just leave empty or show default logic in backend
        }
    };

    const handleRun = async () => {
        setRunning(true);
        try {
            const res = await RunPython(code);
            setOutput(res);
            toast.success("执行完成");
        } catch (e: any) {
            setOutput(`Error: ${e}`);
            toast.error("执行出错: " + e);
        } finally {
            setRunning(false);
        }
    };

    const handleSaveSettings = async () => {
        try {
            await SetPythonPath(pythonPath);
            toast.success("设置已保存");
            setIsSettingsOpen(false);
        } catch (e) {
            toast.error("保存失败");
        }
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 h-full">
            <Card className="flex flex-col h-full">
                <CardHeader className="py-3">
                    <CardTitle className="text-base flex justify-between items-center">
                        <div className="flex items-center gap-2">
                            <span>Python 编辑器</span>
                            <Dialog open={isSettingsOpen} onOpenChange={setIsSettingsOpen}>
                                <DialogTrigger asChild>
                                    <Button variant="ghost" size="icon" className="h-6 w-6">
                                        <Settings className="h-4 w-4" />
                                    </Button>
                                </DialogTrigger>
                                <DialogContent className="sm:max-w-[425px]">
                                    <DialogHeader>
                                        <DialogTitle>Python 环境设置</DialogTitle>
                                        <DialogDescription>
                                            设置 Python 解释器的路径。如果不设置，将自动尝试检测系统中的 python3 或 python。
                                        </DialogDescription>
                                    </DialogHeader>
                                    <div className="grid gap-4 py-4">
                                        <div className="grid grid-cols-4 items-center gap-4">
                                            <Label htmlFor="python-path" className="text-right">
                                                Python 路径
                                            </Label>
                                            <Input
                                                id="python-path"
                                                value={pythonPath}
                                                onChange={(e) => setPythonPathState(e.target.value)}
                                                className="col-span-3"
                                                placeholder="例如: /usr/bin/python3"
                                            />
                                        </div>
                                    </div>
                                    <DialogFooter>
                                        <Button onClick={handleSaveSettings}>保存</Button>
                                    </DialogFooter>
                                </DialogContent>
                            </Dialog>
                        </div>
                        <Button size="sm" onClick={handleRun} disabled={running}>
                            <Play className="mr-2 h-4 w-4" /> {running ? '运行中...' : '运行'}
                        </Button>
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden">
                    <CodeEditor 
                        value={code} 
                        onChange={(val) => setCode(val || "")}
                        language="python"
                    />
                </CardContent>
            </Card>
            
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
