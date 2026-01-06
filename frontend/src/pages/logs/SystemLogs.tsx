import { useState, useEffect, useMemo } from "react";
// @ts-ignore
import { ListLogFiles, ReadLogFile } from "../../../wailsjs/go/logs/LogService";
import { PageContainer } from "../../components/layout/PageContainer";
import { PageHeader } from "../../components/layout/PageHeader";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/Card";
import { FileText, Download, Trash2, RefreshCw, Filter } from "lucide-react";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/ui/Select";

interface LogFile {
  name: string;
  path: string;
}

interface LogEntry {
  时间: string;
  级别: string;
  信息: string;
  模块?: string;
  [key: string]: any;
}

export default function LogPage() {
  const [logs, setLogs] = useState<LogFile[]>([]);
  const [selectedLog, setSelectedLog] = useState<string | null>(null);
  const [logContent, setLogContent] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  
  // Filters
  const [levelFilter, setLevelFilter] = useState<string>("ALL");
  const [searchFilter, setSearchFilter] = useState<string>("");

  useEffect(() => {
    loadLogs();
  }, []);

  useEffect(() => {
    if (selectedLog) {
      readLog(selectedLog);
    }
  }, [selectedLog]);

  const filteredLogs = useMemo(() => {
    return logContent.filter(entry => {
      // Level filter
      if (levelFilter !== "ALL" && entry.级别?.toUpperCase() !== levelFilter) {
        return false;
      }
      
      // Search filter (Module/Message)
      if (searchFilter) {
        const searchLower = searchFilter.toLowerCase();
        const matchMessage = entry.信息?.toLowerCase().includes(searchLower);
        const matchModule = entry.模块?.toLowerCase().includes(searchLower);
        const matchOther = Object.values(entry).some(v => 
          typeof v === 'string' && v.toLowerCase().includes(searchLower)
        );
        
        if (!matchMessage && !matchModule && !matchOther) {
          return false;
        }
      }
      
      return true;
    });
  }, [logContent, levelFilter, searchFilter]);

  const loadLogs = async () => {
    try {
      const files = await ListLogFiles();
      setLogs(files || []);
      if (files && files.length > 0 && !selectedLog) {
        setSelectedLog(files[0].name);
      }
    } catch (err) {
      console.error("Failed to load log files:", err);
    }
  };

  const readLog = async (filename: string) => {
    setLoading(true);
    try {
      const lines = await ReadLogFile(filename);
      
      const parsed: LogEntry[] = [];
      lines?.forEach((line: string) => {
        try {
          if (line.trim()) {
            parsed.push(JSON.parse(line));
          }
        } catch (e) {
          // If not JSON, treat as raw text with minimal structure
          parsed.push({
            时间: "-",
            级别: "RAW",
            信息: line,
          });
        }
      });
      // Reverse to show newest first
      setLogContent(parsed.reverse());
    } catch (err) {
      console.error("Failed to read log file:", err);
    } finally {
      setLoading(false);
    }
  };

  const getLevelColor = (level: string) => {
    switch (level?.toUpperCase()) {
      case "INFO":
        return "text-blue-400";
      case "ERROR":
        return "text-red-400";
      case "WARN":
        return "text-yellow-400";
      case "DEBUG":
        return "text-gray-400";
      default:
        return "text-slate-300";
    }
  };

  return (
    <PageContainer>
      <PageHeader
        title="日志管理"
        description="查看系统运行日志"
        action={
          <Button onClick={() => {
            loadLogs();
            if (selectedLog) readLog(selectedLog);
          }}>
            <RefreshCw className="mr-2 h-4 w-4" /> 刷新
          </Button>
        }
      />

      <div className="flex gap-6 h-[calc(100vh-12rem)]">
        {/* Log File List */}
        <Card className="w-64 flex flex-col">
          <div className="p-4 border-b border-border">
            <h3 className="font-semibold text-card-foreground">日志文件</h3>
          </div>
          <CardContent className="flex-1 overflow-y-auto p-2 space-y-1">
            {logs.map((log) => (
              <button
                key={log.name}
                onClick={() => setSelectedLog(log.name)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  selectedLog === log.name
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                }`}
              >
                {log.name}
              </button>
            ))}
            {logs.length === 0 && (
              <div className="text-center text-muted-foreground py-4 text-sm">
                无日志文件
              </div>
            )}
          </CardContent>
        </Card>

        {/* Log Content */}
        <Card className="flex-1 flex flex-col min-w-0">
          <div className="p-4 border-b border-border space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="font-semibold text-card-foreground">
                {selectedLog ? selectedLog : "请选择日志文件"}
              </h3>
              <span className="text-xs text-muted-foreground">
                {filteredLogs.length} 条记录 (总共 {logContent.length})
              </span>
            </div>
            
            <div className="flex gap-4">
               <div className="w-32">
                 <Select value={levelFilter} onValueChange={setLevelFilter}>
                    <SelectTrigger className="h-8">
                      <SelectValue placeholder="日志等级" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="ALL">所有等级</SelectItem>
                      <SelectItem value="INFO">INFO</SelectItem>
                      <SelectItem value="ERROR">ERROR</SelectItem>
                      <SelectItem value="WARN">WARN</SelectItem>
                      <SelectItem value="DEBUG">DEBUG</SelectItem>
                    </SelectContent>
                 </Select>
               </div>
               <div className="flex-1">
                 <Input 
                   placeholder="搜索内容或模块..." 
                   value={searchFilter}
                   onChange={(e) => setSearchFilter(e.target.value)}
                   className="h-8"
                 />
               </div>
            </div>
          </div>
          
          <CardContent className="flex-1 overflow-y-auto p-4 font-mono text-xs">
            {loading ? (
              <div className="flex justify-center items-center h-full text-muted-foreground">
                加载中...
              </div>
            ) : (
              <div className="space-y-1">
                {filteredLogs.map((entry, idx) => (
                  <div key={idx} className="flex gap-4 hover:bg-muted/50 p-1 rounded border-b border-border/50 last:border-0">
                    <span className="text-muted-foreground w-32 shrink-0">
                      {entry.时间}
                    </span>
                    <span className={`w-12 shrink-0 font-bold ${getLevelColor(entry.级别)}`}>
                      [{entry.级别}]
                    </span>
                    <span className="text-card-foreground break-all">
                      {entry.信息}
                      {Object.keys(entry).map((key) => {
                        if (["时间", "级别", "信息"].includes(key)) return null;
                        return (
                          <span key={key} className="ml-2 text-muted-foreground">
                            {key}={JSON.stringify(entry[key])}
                          </span>
                        );
                      })}
                    </span>
                  </div>
                ))}
                {filteredLogs.length === 0 && (
                  <div className="text-center text-muted-foreground py-10">
                    {logContent.length > 0 ? "没有匹配的日志记录" : "无日志内容"}
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}
