import React, { useState, useEffect, useRef } from 'react';
import { TerminalSquare, Cpu, Loader2, Wifi, WifiOff } from 'lucide-react';
import { api } from '../services/api';

// Detect file type from user input text heuristically
function detectFileType(text) {
  const lower = text.toLowerCase();
  if (lower.includes('package.json') || lower.includes('npm') || lower.includes('node_modules')) return 'package.json';
  if (lower.includes('dockerfile') || lower.includes('docker build')) return 'dockerfile';
  if (lower.includes('go.mod') || lower.includes('go test') || lower.includes('golang')) return 'go.mod';
  if (lower.includes('import react') || lower.includes("from 'react'") || lower.includes('jsx')) return 'jsx';
  if (lower.includes('def ') || lower.includes('import ') && lower.includes('python')) return 'python';
  return 'go'; // default
}

// Transform the backend schema.Pipeline response into the UI's job format
function transformPipelineResponse(backendPipeline) {
  if (!backendPipeline) return null;

  const jobs = [];
  if (backendPipeline.jobs) {
    Object.entries(backendPipeline.jobs).forEach(([id, job]) => {
      jobs.push({
        id,
        runsOn: job.runs_on || 'ubuntu-latest',
        steps: (job.steps || []).map((step, idx) => ({
          id: idx + 1,
          name: step.name || `Step ${idx + 1}`,
          uses: step.uses || undefined,
          run: step.run || undefined,
        })),
      });
    });
  }

  return {
    name: backendPipeline.name || 'Generated Pipeline',
    on: backendPipeline.on || { push: { branches: ['main'] } },
    jobs,
  };
}

export default function AIGenerator({ onPipelineGenerated }) {
  const [input, setInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [backendOnline, setBackendOnline] = useState(false);
  const [error, setError] = useState('');
  const [statusMsg, setStatusMsg] = useState('');
  const checkInterval = useRef(null);

  // Poll health endpoint every 5 seconds
  useEffect(() => {
    const check = async () => {
      const ok = await api.checkHealth();
      setBackendOnline(ok);
    };
    check();
    checkInterval.current = setInterval(check, 5000);
    return () => clearInterval(checkInterval.current);
  }, []);

  const handleGenerate = async () => {
    if (!input.trim()) return;
    setIsGenerating(true);
    setError('');
    setStatusMsg('');

    if (!backendOnline) {
      // Graceful fallback: run local simulation if backend is offline
      setStatusMsg('⚡ Backend offline — using local simulation...');
      await runLocalSimulation();
      setIsGenerating(false);
      return;
    }

    try {
      setStatusMsg('🔍 Analyzing stack...');
      const fileType = detectFileType(input);
      const result = await api.generatePipeline(input, fileType);

      setStatusMsg('✅ Pipeline generated from backend!');
      const uiPipeline = transformPipelineResponse(result);
      if (uiPipeline) {
        onPipelineGenerated(uiPipeline);
      }
    } catch (err) {
      setError(`Backend error: ${err.message}. Falling back to local simulation.`);
      await runLocalSimulation();
    } finally {
      setIsGenerating(false);
    }
  };

  // Local simulation (kept as fallback when backend is offline)
  const runLocalSimulation = () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        const lowerText = input.toLowerCase();
        const pipeline = {
          name: 'Generated Pipeline',
          on: { push: { branches: ['main'] } },
          jobs: [],
        };

        if (lowerText.includes('go') || lowerText.includes('golang')) {
          pipeline.jobs.push({
            id: 'build-go', runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Setup Go', uses: 'actions/setup-go@v4' },
              { id: 3, name: 'Run Tests', run: 'go test ./...' },
              { id: 4, name: 'Build Binary', run: 'go build -o bin/app main.go' },
            ],
          });
        }

        if (lowerText.includes('node') || lowerText.includes('react')) {
          pipeline.jobs.push({
            id: 'build-node', runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Setup Node', uses: 'actions/setup-node@v3' },
              { id: 3, name: 'Install Deps', run: 'npm ci' },
              { id: 4, name: 'Run Linter', run: 'npm run lint' },
              { id: 5, name: 'Build Project', run: 'npm run build' },
            ],
          });
        }

        if (lowerText.includes('docker')) {
          pipeline.jobs.push({
            id: 'docker-build', runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Build Image', run: 'docker build -t myapp:latest .' },
            ],
          });
        }

        if (pipeline.jobs.length === 0) {
          pipeline.jobs.push({
            id: 'default-job', runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Say Hello', run: 'echo "Hello from YamlAnchor!"' },
            ],
          });
        }

        onPipelineGenerated(pipeline);
        resolve();
      }, 1200);
    });
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <TerminalSquare size={16} />
          Input &amp; AI Generator
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.7rem' }}>
          {backendOnline ? (
            <><Wifi size={12} style={{ color: 'var(--accent-green)' }} /><span style={{ color: 'var(--accent-green)' }}>BACKEND LIVE</span></>
          ) : (
            <><WifiOff size={12} style={{ color: 'var(--text-secondary)' }} /><span style={{ color: 'var(--text-secondary)' }}>LOCAL MODE</span></>
          )}
        </div>
      </div>
      <div className="panel-content">
        <div className="ai-input-wrapper">
          <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
            Describe your project stack or paste your code. The AI will analyze it and construct the optimal YamlAnchor pipeline.
            <br /><br />
            <em>Hint: Try mentioning "React and Docker" or "Go with unit tests".</em>
            {!backendOnline && (
              <><br /><br />
              <span style={{ color: 'var(--accent-yellow, #f59e0b)', fontSize: '0.8rem' }}>
                ⚡ Run <code>anchor server</code> to connect the backend for real pipeline generation.
              </span></>
            )}
          </p>

          {error && (
            <div style={{ padding: '0.6rem 0.8rem', background: 'rgba(239,68,68,0.1)', border: '1px solid var(--danger)', borderRadius: '6px', color: 'var(--danger)', fontSize: '0.8rem', marginBottom: '0.75rem' }}>
              {error}
            </div>
          )}

          {statusMsg && !error && (
            <div style={{ padding: '0.6rem 0.8rem', background: 'rgba(99,102,241,0.1)', border: '1px solid var(--accent)', borderRadius: '6px', color: 'var(--accent)', fontSize: '0.8rem', marginBottom: '0.75rem' }}>
              {statusMsg}
            </div>
          )}

          <textarea
            className="ai-textarea"
            placeholder="e.g., I have a Go backend that needs testing and building, then packaging into a Docker container..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
          />

          <button
            className="btn btn-ai"
            onClick={handleGenerate}
            disabled={isGenerating || !input.trim()}
          >
            {isGenerating ? (
              <><Loader2 size={18} className="animate-spin" />Processing...</>
            ) : (
              <><Cpu size={18} />Generate Pipeline</>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
