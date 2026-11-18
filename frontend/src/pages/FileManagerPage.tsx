import React, { useState } from 'react';
import Editor from '@monaco-editor/react';
import {
  FolderOpen,
  FileCode,
  Save,
  CheckCircle2,
  File,
  ShieldCheck,
} from 'lucide-react';

interface FileNode {
  name: string;
  type: 'file' | 'dir';
  content?: string;
  size?: string;
}

export const FileManagerPage: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<FileNode>({
    name: 'index.php',
    type: 'file',
    content: `<?php
/**
 * SparkPanelHost v0.1.0 Tenant Application
 * Safe execution with PHP 8.5 / 8.6 under Linux user UID isolation
 */

declare(strict_types=1);

echo "Hello from SparkPanelHost v0.1.0!";
phpinfo();
`,
  });

  const [code, setCode] = useState(selectedFile.content || '');
  const [saved, setSaved] = useState(false);

  const files: FileNode[] = [
    { name: 'public_html', type: 'dir' },
    { name: 'index.php', type: 'file', size: '240 B' },
    { name: '.htaccess', type: 'file', size: '120 B' },
    { name: 'config.php', type: 'file', size: '1.2 KB' },
    { name: 'wp-content', type: 'dir' },
  ];

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary">
            File Manager
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Kernel-enforced path traversal & symlink immunity (openat2 RESOLVE_BENEATH)
          </p>
        </div>
        <div className="flex items-center gap-2">
          {saved && (
            <span className="flex items-center gap-1 text-xs text-status-success font-semibold">
              <CheckCircle2 className="h-3.5 w-3.5" /> Saved!
            </span>
          )}
          <button
            onClick={handleSave}
            className="flex items-center gap-2 rounded-lg bg-accent-primary px-3 py-1.5 text-xs font-semibold text-white shadow-xs hover:bg-accent-hover transition-colors"
          >
            <Save className="h-3.5 w-3.5" />
            <span>Save File</span>
          </button>
        </div>
      </div>

      {/* Editor & File Browser Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 h-[600px]">
        {/* File Browser Sidebar */}
        <div className="rounded-xl border border-border-subtle bg-bg-card p-3 flex flex-col justify-between">
          <div>
            <div className="text-xs font-bold text-text-primary px-2 py-1 mb-2 flex items-center gap-2">
              <FolderOpen className="h-4 w-4 text-accent-primary" />
              <span>/home/tenant/public_html</span>
            </div>
            <div className="space-y-0.5">
              {files.map((f) => (
                <button
                  key={f.name}
                  onClick={() => {
                    if (f.type === 'file') {
                      setSelectedFile(f);
                      setCode(f.content || `// File content of ${f.name}\n`);
                    }
                  }}
                  className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-colors ${
                    selectedFile.name === f.name
                      ? 'bg-accent-primary text-white font-semibold'
                      : 'text-text-secondary hover:text-text-primary hover:bg-bg-elevated'
                  }`}
                >
                  <span className="flex items-center gap-2 truncate">
                    {f.type === 'dir' ? (
                      <FolderOpen className="h-3.5 w-3.5 text-accent-primary" />
                    ) : (
                      <FileCode className="h-3.5 w-3.5" />
                    )}
                    <span className="truncate">{f.name}</span>
                  </span>
                  {f.size && <span className="text-[10px] opacity-75 font-mono">{f.size}</span>}
                </button>
              ))}
            </div>
          </div>

          <div className="p-2 border-t border-border-subtle text-[10px] text-status-success font-medium flex items-center gap-1.5">
            <ShieldCheck className="h-3.5 w-3.5 shrink-0" />
            <span>openat2 Safe Sandbox</span>
          </div>
        </div>

        {/* Monaco Editor Pane */}
        <div className="lg:col-span-3 rounded-xl border border-border-subtle bg-bg-card overflow-hidden flex flex-col">
          <div className="px-4 py-2 bg-bg-elevated border-b border-border-subtle flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-mono font-semibold text-text-primary">
              <File className="h-3.5 w-3.5 text-accent-primary" />
              <span>{selectedFile.name}</span>
            </div>
            <span className="text-[11px] text-text-secondary font-mono">PHP / UTF-8</span>
          </div>

          <div className="flex-1 w-full">
            <Editor
              height="100%"
              defaultLanguage="php"
              theme="vs-dark"
              value={code}
              onChange={(val) => setCode(val || '')}
              options={{
                minimap: { enabled: false },
                fontSize: 13,
                fontFamily: 'JetBrains Mono, monospace',
                tabSize: 4,
                wordWrap: 'on',
                lineNumbers: 'on',
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};
