import React, { useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { Terminal as TerminalIcon, ShieldCheck } from 'lucide-react';

export const WebTerminalPage: React.FC = () => {
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!terminalRef.current) return;

    const term = new XTerm({
      cursorBlink: true,
      fontFamily: 'JetBrains Mono, monospace',
      fontSize: 13,
      theme: {
        background: '#09090b',
        foreground: '#fafafa',
        cursor: '#3b82f6',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();

    term.writeln('\x1b[1;34m==============================================================\x1b[0m');
    term.writeln('\x1b[1;32m  SparkPanelHost v0.1.0 Web Terminal (PTY Session Active)\x1b[0m');
    term.writeln('  UID/GID dropped to tenant context | cgroups v2 isolated');
    term.writeln('\x1b[1;34m==============================================================\x1b[0m\r\n');

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/terminal/ws`;
    const ws = new WebSocket(wsUrl);

    ws.binaryType = 'arraybuffer';

    ws.onmessage = (event) => {
      term.write(new Uint8Array(event.data));
    };

    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });

    const handleResize = () => fitAddon.fit();
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      ws.close();
      term.dispose();
    };
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-text-primary flex items-center gap-2">
            <TerminalIcon className="h-5 w-5 text-accent-primary" />
            <span>Web SSH Terminal</span>
          </h1>
          <p className="text-xs text-text-secondary mt-0.5">
            Real-time interactive PTY shell locked to tenant UID/GID and cgroups v2 slice
          </p>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-status-success font-semibold">
          <ShieldCheck className="h-4 w-4" />
          <span>Privilege Dropped</span>
        </div>
      </div>

      <div className="rounded-xl border border-border-subtle bg-[#09090b] p-4 shadow-xl overflow-hidden h-[540px]">
        <div ref={terminalRef} className="h-full w-full" />
      </div>
    </div>
  );
};
