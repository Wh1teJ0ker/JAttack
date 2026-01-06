import { useEffect, useState } from "react";
import { Button } from "../../components/ui/Button";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import { Input } from "../../components/ui/Input";
import { Label } from "../../components/ui/Label";
import { Badge } from "../../components/ui/Badge";
import { 
    Dialog, 
    DialogContent, 
    DialogHeader, 
    DialogTitle, 
    DialogFooter,
    DialogDescription
} from "../../components/ui/Dialog";
import { 
    Select, 
    SelectContent, 
    SelectItem, 
    SelectTrigger, 
    SelectValue 
} from "../../components/ui/Select";
import { Textarea } from "../../components/ui/Textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/ui/Tabs";
import { Plus, Search, Trash2, Edit, AlertTriangle, Bug, Eye } from "lucide-react";
import { toast } from "sonner";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

import { 
    SearchVulns, 
    AddVuln, 
    UpdateVuln, 
    DeleteVuln
} from "../../../wailsjs/go/vuln/VulnService";
import { vuln } from "../../../wailsjs/go/models";

// 默认空漏洞对象
const emptyVuln: vuln.Vulnerability = {
    id: 0,
    name: "",
    product: "",
    vuln_type: "",
    severity: "Medium",
    description: "",
    details: "",
    status: "Open",
    poc_type: "nuclei",
    poc_content: "",
    reference: "",
    created_at: "",
    // @ts-ignore - 类方法
    convertValues: () => {}
};

export default function VulnList() {
    const [vulns, setVulns] = useState<vuln.Vulnerability[]>([]);
    const [keyword, setKeyword] = useState("");
    
    // Dialog States
    const [dialogOpen, setDialogOpen] = useState(false);
    const [currentVuln, setCurrentVuln] = useState<vuln.Vulnerability>({...emptyVuln});
    const [isEditing, setIsEditing] = useState(false);

    // View Dialog States
    const [viewDialogOpen, setViewDialogOpen] = useState(false);

    // Delete Dialog States
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deleteId, setDeleteId] = useState<number | null>(null);

    const loadVulns = async () => {
        try {
            const data = await SearchVulns(keyword);
            setVulns(data || []);
        } catch (err) {
            toast.error("加载漏洞列表失败");
        }
    };

    useEffect(() => {
        loadVulns();
    }, []);

    const handleSearch = () => {
        loadVulns();
    };

    const handleOpenAdd = () => {
        setCurrentVuln({...emptyVuln});
        setIsEditing(false);
        setDialogOpen(true);
    };

    const handleOpenEdit = (v: vuln.Vulnerability) => {
        setCurrentVuln({...v});
        setIsEditing(true);
        setDialogOpen(true);
    };

    const handleOpenView = (v: vuln.Vulnerability) => {
        setCurrentVuln({...v});
        setViewDialogOpen(true);
    };

    const handleSave = async () => {
        if (!currentVuln.name) {
            toast.error("请输入漏洞名称");
            return;
        }

        try {
            if (isEditing) {
                await UpdateVuln(currentVuln);
                toast.success("漏洞更新成功");
            } else {
                await AddVuln(currentVuln);
                toast.success("漏洞添加成功");
            }
            setDialogOpen(false);
            loadVulns();
        } catch (err) {
            toast.error("保存失败: " + err);
        }
    };

    const handleDelete = (id: number) => {
        setDeleteId(id);
        setDeleteDialogOpen(true);
    };

    const confirmDelete = async () => {
        if (deleteId === null) return;
        
        try {
            await DeleteVuln(deleteId);
            toast.success("删除成功");
            loadVulns();
            setDeleteDialogOpen(false);
        } catch (err) {
            toast.error("删除失败");
        }
    };

    const getSeverityColor = (severity: string): "default" | "secondary" | "destructive" | "outline" => {
        switch (severity.toLowerCase()) {
            case "critical": return "destructive";
            case "high": return "destructive";
            case "medium": return "default";
            case "low": return "secondary";
            default: return "outline";
        }
    };

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <h2 className="text-xl font-bold tracking-tight">漏洞列表</h2>
                <div className="flex items-center space-x-2">
                    <Button onClick={handleOpenAdd}>
                        <Plus className="mr-2 h-4 w-4" /> 新增漏洞
                    </Button>
                </div>
            </div>

            <Card className="flex-1 flex flex-col overflow-hidden">
                <CardHeader className="py-3 shrink-0">
                    <div className="flex items-center gap-2">
                        <Input 
                            placeholder="搜索漏洞名称、组件、类型、描述..." 
                            value={keyword}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setKeyword(e.target.value)}
                            onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => e.key === "Enter" && handleSearch()}
                            className="flex-1"
                        />
                        <Button variant="secondary" onClick={handleSearch}>
                            <Search className="mr-2 h-4 w-4" /> 搜索
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="p-0 flex-1 overflow-auto">
                    <table className="w-full text-sm text-left">
                        <thead className="bg-muted/50 text-muted-foreground font-medium sticky top-0 z-10 backdrop-blur-sm">
                            <tr>
                                <th className="p-4 w-[50px]">序号</th>
                                <th className="p-4 w-[200px]">漏洞名称</th>
                                <th className="p-4 w-[150px]">归属组件</th>
                                <th className="p-4 w-[120px]">类型</th>
                                <th className="p-4 w-[100px]">等级</th>
                                <th className="p-4">描述</th>
                                <th className="p-4 w-[100px]">POC类型</th>
                                <th className="p-4 w-[180px] text-right">操作</th>
                            </tr>
                        </thead>
                        <tbody>
                            {vulns.map((vuln, index) => (
                                <tr key={vuln.id} className="border-t border-border hover:bg-muted/50 transition-colors">
                                    <td className="p-4 text-muted-foreground">{index + 1}</td>
                                    <td className="p-4 font-medium">
                                        <div className="flex items-center cursor-pointer hover:underline" onClick={() => handleOpenView(vuln)}>
                                            <Bug className="mr-2 h-4 w-4 text-muted-foreground" />
                                            {vuln.name}
                                        </div>
                                    </td>
                                    <td className="p-4 text-muted-foreground">{vuln.product || "-"}</td>
                                    <td className="p-4">
                                        {vuln.vuln_type && (
                                            <Badge variant="outline" className="font-normal text-xs">
                                                {vuln.vuln_type}
                                            </Badge>
                                        )}
                                    </td>
                                    <td className="p-4">
                                        <Badge variant={getSeverityColor(vuln.severity)}>
                                            {vuln.severity}
                                        </Badge>
                                    </td>
                                    <td className="p-4 truncate max-w-xs text-muted-foreground">
                                        {vuln.description}
                                    </td>
                                    <td className="p-4">
                                        <Badge variant="secondary" className="font-mono text-xs">
                                            {vuln.poc_type}
                                        </Badge>
                                    </td>
                                    <td className="p-4 text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenView(vuln)} title="查看详情">
                                                <Eye className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenEdit(vuln)} title="编辑">
                                                <Edit className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => handleDelete(vuln.id)} title="删除">
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {vulns.length === 0 && (
                                <tr>
                                    <td colSpan={8} className="p-8 text-center text-muted-foreground">
                                        未找到匹配的漏洞记录。
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </CardContent>
            </Card>

            {/* 新增/编辑 对话框 */}
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>{isEditing ? "编辑漏洞" : "新增漏洞"}</DialogTitle>
                        <DialogDescription>
                            填写漏洞详细信息及 POC 配置。
                        </DialogDescription>
                    </DialogHeader>
                    
                    <Tabs defaultValue="basic" className="w-full">
                        <TabsList className="grid w-full grid-cols-3">
                            <TabsTrigger value="basic">基本信息</TabsTrigger>
                            <TabsTrigger value="details">详细详情</TabsTrigger>
                            <TabsTrigger value="poc">POC 配置</TabsTrigger>
                        </TabsList>
                        
                        <TabsContent value="basic" className="space-y-4 py-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>漏洞名称</Label>
                                    <Input 
                                        value={currentVuln.name} 
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCurrentVuln({...currentVuln, name: e.target.value})}
                                        placeholder="例如: Apache Log4j RCE"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>归属组件/产品</Label>
                                    <Input 
                                        value={currentVuln.product} 
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCurrentVuln({...currentVuln, product: e.target.value})}
                                        placeholder="例如: Apache Log4j2"
                                    />
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>漏洞类型</Label>
                                    <Input 
                                        value={currentVuln.vuln_type} 
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCurrentVuln({...currentVuln, vuln_type: e.target.value})}
                                        placeholder="例如: RCE, SQLi"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>严重等级</Label>
                                    <Select 
                                        value={currentVuln.severity} 
                                        onValueChange={(val: string) => setCurrentVuln({...currentVuln, severity: val})}
                                    >
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="Critical">Critical (严重)</SelectItem>
                                            <SelectItem value="High">High (高危)</SelectItem>
                                            <SelectItem value="Medium">Medium (中危)</SelectItem>
                                            <SelectItem value="Low">Low (低危)</SelectItem>
                                            <SelectItem value="Info">Info (信息)</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label>简要描述</Label>
                                <Input 
                                    value={currentVuln.description} 
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCurrentVuln({...currentVuln, description: e.target.value})}
                                    placeholder="一句话描述漏洞影响"
                                />
                            </div>

                            <div className="space-y-2">
                                <Label>参考链接</Label>
                                <Input 
                                    value={currentVuln.reference} 
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCurrentVuln({...currentVuln, reference: e.target.value})}
                                    placeholder="https://..."
                                />
                            </div>
                        </TabsContent>

                        <TabsContent value="details" className="space-y-4 py-4">
                            <div className="space-y-2 h-[400px]">
                                <Label>漏洞详情</Label>
                                <Textarea 
                                    value={currentVuln.details} 
                                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setCurrentVuln({...currentVuln, details: e.target.value})}
                                    placeholder="漏洞原理、复现步骤等详细信息"
                                    className="h-full resize-none font-mono text-sm"
                                />
                            </div>
                        </TabsContent>

                        <TabsContent value="poc" className="space-y-4 py-4">
                            <div className="space-y-2">
                                <Label>POC 类型</Label>
                                <Select 
                                    value={currentVuln.poc_type} 
                                    onValueChange={(val: string) => setCurrentVuln({...currentVuln, poc_type: val})}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="nuclei">Nuclei 模板</SelectItem>
                                        <SelectItem value="python">Python 脚本</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2 h-[350px]">
                                <Label>POC 内容 / 脚本代码</Label>
                                <Textarea 
                                    value={currentVuln.poc_content} 
                                    onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setCurrentVuln({...currentVuln, poc_content: e.target.value})}
                                    placeholder={currentVuln.poc_type === 'nuclei' ? "在此粘贴 Nuclei YAML 模板内容..." : "在此粘贴 Python 脚本代码..."}
                                    className="h-full resize-none font-mono text-sm"
                                />
                            </div>
                        </TabsContent>
                    </Tabs>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDialogOpen(false)}>取消</Button>
                        <Button onClick={handleSave}>保存</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* 查看详情 对话框 */}
            <Dialog open={viewDialogOpen} onOpenChange={setViewDialogOpen}>
                <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>漏洞详情: {currentVuln.name}</DialogTitle>
                    </DialogHeader>
                    
                    <Tabs defaultValue="info" className="w-full">
                        <TabsList className="grid w-full grid-cols-3">
                            <TabsTrigger value="info">基本信息</TabsTrigger>
                            <TabsTrigger value="details">详细详情</TabsTrigger>
                            <TabsTrigger value="poc">POC</TabsTrigger>
                        </TabsList>
                        
                        <TabsContent value="info" className="py-4 space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <Label className="text-muted-foreground">组件/产品</Label>
                                    <div className="mt-1 font-medium">{currentVuln.product || "-"}</div>
                                </div>
                                <div>
                                    <Label className="text-muted-foreground">类型</Label>
                                    <div className="mt-1">{currentVuln.vuln_type || "-"}</div>
                                </div>
                                <div>
                                    <Label className="text-muted-foreground">等级</Label>
                                    <div className="mt-1">
                                        <Badge variant={getSeverityColor(currentVuln.severity)}>{currentVuln.severity}</Badge>
                                    </div>
                                </div>
                                <div>
                                    <Label className="text-muted-foreground">参考链接</Label>
                                    <div className="mt-1 text-blue-500 truncate">{currentVuln.reference || "-"}</div>
                                </div>
                            </div>
                            <div>
                                <Label className="text-muted-foreground">简要描述</Label>
                                <div className="mt-1 p-2 bg-muted rounded-md text-sm">{currentVuln.description || "无"}</div>
                            </div>
                        </TabsContent>
                        
                        <TabsContent value="details" className="py-4">
                            <div className="bg-muted p-4 rounded-md font-mono text-sm min-h-[300px] overflow-auto prose prose-sm dark:prose-invert max-w-none">
                                {currentVuln.details ? (
                                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                        {currentVuln.details}
                                    </ReactMarkdown>
                                ) : (
                                    <span className="text-muted-foreground">暂无详细信息</span>
                                )}
                            </div>
                        </TabsContent>
                        
                        <TabsContent value="poc" className="py-4 space-y-4">
                            <div>
                                <Badge variant="secondary">{currentVuln.poc_type}</Badge>
                            </div>
                            <div className="bg-black text-green-400 p-4 rounded-md overflow-auto whitespace-pre font-mono text-xs max-h-[500px]">
                                {currentVuln.poc_content || "# 无 POC 内容"}
                            </div>
                        </TabsContent>
                    </Tabs>
                </DialogContent>
            </Dialog>

            {/* 删除确认 对话框 */}
            <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle className="flex items-center text-destructive">
                            <AlertTriangle className="mr-2 h-5 w-5" /> 
                            确认删除
                        </DialogTitle>
                        <DialogDescription>
                            此操作无法撤销。这将永久删除该漏洞记录。
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>取消</Button>
                        <Button variant="destructive" onClick={confirmDelete}>确认删除</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}