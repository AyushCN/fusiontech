import React, { useState, useEffect } from 'react';
import AIGenerator from './components/AIGenerator';
import VisualGraph from './components/VisualGraph';
import { Anchor, Download, Copy, Check, FileCode2 } from 'lucide-react';
import hljs from 'highlight.js/lib/core';
import yamlLanguage from 'highlight.js/lib/languages/yaml';
import 'highlight.js/styles/atom-one-dark-reasonable.css'; // Fits the terminal theme better
import yaml from 'js-yaml';

hljs.registerLanguage('yaml', yamlLanguage);

function App() {
  const [pipelineData, setPipelineData] = null; // We'll hold the raw object for the graph
  const [pipelineState, setPipelineState] = useState(null);
  const [yamlContent, setYamlContent] = useState('');
  const [copied, setCopied] = useState(false);

  // Convert pipeline object to YAML string whenever it changes
  useEffect(() => {
    if (!pipelineState) return;
    
    // Format to match YamlAnchor spec
    const yamlObj = {
      name: pipelineState.name,
      on: pipelineState.on,
      jobs: {}
    };

    if (pipelineState.jobs) {
      pipelineState.jobs.forEach(job => {
        yamlObj.jobs[job.id] = {
          'runs-on': job.runsOn,
          steps: job.steps.map(step => {
            const s = { name: step.name };
            if (step.uses) s.uses = step.uses;
            if (step.run) s.run = step.run;
            return s;
          })
        };
      });
    }

    try {
      const yamlStr = yaml.dump(yamlObj, { lineWidth: -1 });
      setYamlContent(yamlStr);
    } catch (e) {
      console.error('Failed to generate YAML', e);
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
          <h1>YamlAnchor <span style={{ color: 'var(--accent-green)' }}>Studio</span></h1>
        </div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)' }}>
          AI_ENGINE: SIMULATED | STATUS: ONLINE
        </div>
      </header>

      <main className="workspace">
        {/* Left Panel: Mock AI Generator */}
        <AIGenerator onPipelineGenerated={handlePipelineGenerated} />

        {/* Middle Panel: YAML Output */}
        <div className="panel">
          <div className="panel-header">
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <FileCode2 size={16} />
              Generated YAML
            </div>
            {yamlContent && (
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button className="btn" onClick={handleCopy} style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}>
                  {copied ? <Check size={14} color="var(--accent-green)" /> : <Copy size={14} />}
                  {copied ? 'COPIED' : 'COPY'}
                </button>
                <button className="btn btn-ai" onClick={handleDownload} style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}>
                  <Download size={14} /> DL
                </button>
              </div>
            )}
          </div>
          <div className="panel-content" style={{ background: '#0d1117' }}>
            {yamlContent ? (
              <pre style={{ margin: 0, padding: '1rem', minHeight: '100%' }}>
                <code className="language-yaml" style={{ background: 'transparent', padding: 0 }}>
                  {yamlContent}
                </code>
              </pre>
            ) : (
              <div style={{ padding: '2rem', color: 'var(--text-secondary)', textAlign: 'center', opacity: 0.5 }}>
                // Output will appear here
              </div>
            )}
          </div>
        </div>

        {/* Right Panel: Visual Graph & Fault Detection */}
        <VisualGraph pipeline={pipelineState} />
      </main>
    </div>
  );
}

export default App;
