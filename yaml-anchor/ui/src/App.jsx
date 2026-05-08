import { useMemo, useState, useEffect } from 'react';
import AIGenerator from './components/AIGenerator';
import VisualGraph from './components/VisualGraph';
import { Anchor, Download, Copy, Check, FileCode2, GitBranch, ShieldCheck, TerminalSquare } from 'lucide-react';
import hljs from 'highlight.js/lib/core';
import yamlLanguage from 'highlight.js/lib/languages/yaml';
import 'highlight.js/styles/atom-one-dark-reasonable.css'; // Fits the terminal theme better
import yaml from 'js-yaml';

hljs.registerLanguage('yaml', yamlLanguage);

function App() {
  const [pipelineState, setPipelineState] = useState(null);
  const [copied, setCopied] = useState(false);

  const pipelineStats = useMemo(() => {
    const jobs = Array.isArray(pipelineState?.jobs) ? pipelineState.jobs : [];
    const steps = jobs.reduce((total, job) => total + (job.steps?.length || 0), 0);
    const remoteActions = jobs.reduce(
      (total, job) => total + (job.steps || []).filter((step) => step.uses).length,
      0
    );
    const shellSteps = steps - remoteActions;

    return { jobs: jobs.length, steps, remoteActions, shellSteps };
  }, [pipelineState]);

  const { yamlContent, yamlError } = useMemo(() => {
    if (!pipelineState) {
      return { yamlContent: '', yamlError: '' };
    }
    try {
      const jobs = Array.isArray(pipelineState.jobs) ? pipelineState.jobs : [];
      const yamlObj = {
        name: pipelineState.name || 'Generated Pipeline',
        on: pipelineState.on || { push: { branches: ['main'] } },
        jobs: {}
      };

      jobs.forEach(job => {
        yamlObj.jobs[job.id] = {
          'runs-on': job.runsOn || 'ubuntu-latest',
          steps: (job.steps || []).map(step => {
            const s = { name: step.name || 'Step' };
            if (step.uses) s.uses = step.uses;
            if (step.run) s.run = step.run;
            return s;
          })
        };
      });

      const yamlStr = yaml.dump(yamlObj, { lineWidth: -1 });
      return { yamlContent: yamlStr, yamlError: '' };
    } catch (e) {
      console.error('Failed to generate YAML', e);
      return { yamlContent: '', yamlError: 'Failed to generate YAML from the pipeline response.' };
    }
  }, [pipelineState]);

  // Syntax highlighting effect
  useEffect(() => {
    if (yamlContent) {
      document.querySelectorAll('pre code').forEach((block) => {
        // highlight.js modifies DOM, we must re-apply or it loses formatting
        block.removeAttribute('data-highlighted');
        hljs.highlightElement(block);
      });
    }
  }, [yamlContent]);

  const handlePipelineGenerated = (newPipeline) => {
    setPipelineState(newPipeline);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(yamlContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleDownload = () => {
    const element = document.createElement("a");
    const file = new Blob([yamlContent], {type: 'text/yaml'});
    element.href = URL.createObjectURL(file);
    element.download = "anchor.yaml";
    document.body.appendChild(element); 
    element.click();
    document.body.removeChild(element);
  };

  return (
    <div className="app-container">
      <header className="header">
        <div className="logo-container">
          <Anchor className="logo-icon" size={28} />
          <div>
            <h1>YamlAnchor <span>Studio</span></h1>
            <p className="product-subtitle">Typed CI preflight, workflow generation, and local simulation</p>
          </div>
        </div>
        <div className="header-metrics" aria-label="Current pipeline summary">
          <div className="metric-pill">
            <GitBranch size={14} />
            <span>{pipelineStats.jobs} jobs</span>
          </div>
          <div className="metric-pill">
            <TerminalSquare size={14} />
            <span>{pipelineStats.steps} steps</span>
          </div>
          <div className="metric-pill healthy">
            <ShieldCheck size={14} />
            <span>preflight ready</span>
          </div>
        </div>
      </header>

      <main className="workspace">
        <AIGenerator onPipelineGenerated={handlePipelineGenerated} />

        <div className="panel output-panel">
          <div className="panel-header">
            <div className="panel-title">
              <FileCode2 size={16} />
              anchor.yaml
            </div>
            {yamlContent && (
              <div className="panel-actions">
                <button className="icon-btn" onClick={handleCopy} aria-label="Copy generated YAML">
                  {copied ? <Check size={14} color="var(--accent-green)" /> : <Copy size={14} />}
                </button>
                <button className="icon-btn primary" onClick={handleDownload} aria-label="Download anchor.yaml">
                  <Download size={14} />
                </button>
              </div>
            )}
          </div>
          <div className="panel-content yaml-surface">
            {yamlError ? (
              <div className="empty-state danger-text">
                {yamlError}
              </div>
            ) : yamlContent ? (
              <pre className="yaml-preview">
                <code className="language-yaml">
                  {yamlContent}
                </code>
              </pre>
            ) : (
              <div className="empty-state">
                <FileCode2 size={36} />
                <span>Generated anchor.yaml will appear here</span>
              </div>
            )}
          </div>
          <div className="summary-strip">
            <span>{pipelineStats.remoteActions} action shims</span>
            <span>{pipelineStats.shellSteps} shell steps</span>
            <span>{yamlContent ? 'YAML synchronized' : 'Awaiting pipeline'}</span>
          </div>
        </div>

        <VisualGraph pipeline={pipelineState} />
      </main>
    </div>
  );
}

export default App;
