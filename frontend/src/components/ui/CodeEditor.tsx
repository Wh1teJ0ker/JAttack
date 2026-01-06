import Editor, { OnMount } from '@monaco-editor/react';
import { useEffect, useState } from 'react';

// Python 关键字和内置函数列表
const pythonKeywords = [
    'False', 'None', 'True', 'and', 'as', 'assert', 'async', 'await', 'break', 'class', 'continue', 'def', 'del', 'elif', 'else', 'except', 'finally', 'for', 'from', 'global', 'if', 'import', 'in', 'is', 'lambda', 'nonlocal', 'not', 'or', 'pass', 'raise', 'return', 'try', 'while', 'with', 'yield'
];

const pythonBuiltins = [
    'abs', 'all', 'any', 'ascii', 'bin', 'bool', 'breakpoint', 'bytearray', 'bytes', 'callable', 'chr', 'classmethod', 'compile', 'complex', 'delattr', 'dict', 'dir', 'divmod', 'enumerate', 'eval', 'exec', 'filter', 'float', 'format', 'frozenset', 'getattr', 'globals', 'hasattr', 'hash', 'help', 'hex', 'id', 'input', 'int', 'isinstance', 'issubclass', 'iter', 'len', 'list', 'locals', 'map', 'max', 'memoryview', 'min', 'next', 'object', 'oct', 'open', 'ord', 'pow', 'print', 'property', 'range', 'repr', 'reversed', 'round', 'set', 'setattr', 'slice', 'sorted', 'staticmethod', 'str', 'sum', 'super', 'tuple', 'type', 'vars', 'zip', '__import__'
];

let isPythonCompletionRegistered = false;

interface CodeEditorProps {
    value: string;
    onChange: (value: string | undefined) => void;
    language?: string;
    height?: string | number;
    readOnly?: boolean;
}

export function CodeEditor({ 
    value, 
    onChange, 
    language = "python", 
    height = "100%", 
    readOnly = false 
}: CodeEditorProps) {
    const [mounted, setMounted] = useState(false);

    const handleEditorDidMount: OnMount = (editor, monaco) => {
        setMounted(true);
        // Add any custom configuration here if needed
        editor.updateOptions({
            fontFamily: "'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace",
            fontLigatures: true,
        });

        // 注册 Python 自动补全 (仅注册一次)
        if (language === 'python' && !isPythonCompletionRegistered) {
            isPythonCompletionRegistered = true;
            
            monaco.languages.registerCompletionItemProvider('python', {
                provideCompletionItems: (model, position) => {
                    const word = model.getWordUntilPosition(position);
                    const range = {
                        startLineNumber: position.lineNumber,
                        endLineNumber: position.lineNumber,
                        startColumn: word.startColumn,
                        endColumn: word.endColumn
                    };
                    
                    const suggestions = [
                        // Keywords
                        ...pythonKeywords.map(k => ({
                            label: k,
                            kind: monaco.languages.CompletionItemKind.Keyword,
                            insertText: k,
                            range: range
                        })),
                        // Builtins
                        ...pythonBuiltins.map(b => ({
                            label: b,
                            kind: monaco.languages.CompletionItemKind.Function,
                            insertText: b,
                            range: range
                        })),
                        // Snippets
                        {
                            label: 'def',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'def ${1:name}(${2:args}):\n\t${3:pass}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'Function definition',
                            range: range
                        },
                        {
                            label: 'class',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'class ${1:ClassName}:\n\tdef __init__(self, ${2:args}):\n\t\t${3:pass}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'Class definition',
                            range: range
                        },
                        {
                            label: 'if',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'if ${1:condition}:\n\t${2:pass}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'If statement',
                            range: range
                        },
                        {
                            label: 'ifmain',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'if __name__ == "__main__":\n\t${1:main()}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'Main check',
                            range: range
                        },
                        {
                            label: 'try',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'try:\n\t${1:pass}\nexcept ${2:Exception} as ${3:e}:\n\t${4:print(e)}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'Try/Except block',
                            range: range
                        },
                        {
                            label: 'for',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'for ${1:item} in ${2:iterable}:\n\t${3:pass}',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'For loop',
                            range: range
                        },
                         {
                            label: 'print',
                            kind: monaco.languages.CompletionItemKind.Snippet,
                            insertText: 'print(${1:object})',
                            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                            detail: 'Print function',
                            range: range
                        }
                    ];
                    
                    return { suggestions: suggestions };
                }
            });
        }

        // 针对 Python 语言的特定配置建议
        if (language === 'python') {
            // 这里可以做一些 Python 特有的初始化，比如注册快捷键等
        }
    };

    return (
        <div className="h-full w-full border rounded-md overflow-hidden bg-[#1e1e1e]">
            <Editor
                height={height}
                defaultLanguage={language}
                language={language}
                value={value}
                onChange={onChange}
                theme="vs-dark"
                onMount={handleEditorDidMount}
                loading={<div className="text-muted-foreground p-4">Loading editor...</div>}
                options={{
                    // 基础外观
                    minimap: { enabled: true, scale: 0.75 },
                    scrollBeyondLastLine: false,
                    fontSize: 14,
                    lineNumbers: "on",
                    renderWhitespace: "selection",
                    fontFamily: "'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace",
                    fontLigatures: true,
                    cursorBlinking: "smooth",
                    cursorSmoothCaretAnimation: "on",
                    smoothScrolling: true,
                    padding: { top: 16, bottom: 16 },

                    // 自动换行与缩进 (针对 Python 优化)
                    wordWrap: "on",           // 开启自动换行
                    wordWrapColumn: 80,       // 换行参考列
                    wrappingIndent: "same",   // 换行后的缩进保持一致
                    tabSize: 4,               // Python 默认 4 空格缩进
                    insertSpaces: true,       // 使用空格代替 Tab
                    detectIndentation: false, // 强制使用上述缩进配置

                    // 辅助功能
                    rulers: [80, 120],        // 垂直标尺，提示代码长度
                    bracketPairColorization: { enabled: true }, // 括号彩色配对
                    guides: {
                        indentation: true,    // 缩进辅助线
                        bracketPairs: true,   // 括号匹配辅助线
                    },
                    folding: true,            // 代码折叠
                    foldingHighlight: true,   // 折叠高亮
                    matchBrackets: "always",  // 总是匹配括号
                    autoClosingBrackets: "always", // 自动闭合括号
                    autoClosingQuotes: "always",   // 自动闭合引号
                    
                    // 智能提示与格式化
                    suggest: {
                        preview: true,
                        showWords: false,     // 减少纯文本干扰，优先显示语法提示
                    },
                    formatOnPaste: true,
                    formatOnType: true,
                    acceptSuggestionOnEnter: "smart", // 智能回车接受建议
                    
                    // 其他
                    contextmenu: true,        // 启用右键菜单
                    mouseWheelZoom: true,     // 允许 Ctrl + 滚轮缩放字体
                    automaticLayout: true,    // 自动适应容器大小
                }}
            />
        </div>
    );
}
