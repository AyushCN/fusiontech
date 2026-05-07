import React, { useState, useEffect } from 'react';
import PipelineBuilder from './components/PipelineBuilder';
import { Download, Copy, Check, Anchor, Terminal } from 'lucide-react';
import hljs from 'highlight.js/lib/core';
import yaml from 'highlight.js/lib/languages/yaml';
import 'highlight.js/styles/github-dark.css';

hljs.registerLanguage('yaml', yaml);

function App() {
  const [yamlContent, setYamlContent] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (yamlContent) {
      document.querySelectorAll('pre code').forEach((block) => {
        hljs.highlightElement(block);
      });
    }
  }, [yamlContent]);

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
    document.body.appendChild(element); // Required for this to work in FireFox
    element.click();
    document.body.removeChild(element);
  };

  return (
    <div className="app-container">
      <header className="header">
        <div className="logo-container">
          <Anchor className="logo-icon" size={32} />
          <h1>YamlAnchor Studio</h1>
        </div>
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Design CI/CD visually.
          </span>
        </div>
      </header>

      <PipelineBuilder onYamlChange={setYamlContent} />

      <div className="preview-section glass-panel" style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column' }}>
        <div className="section-header" style={{ justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Terminal size={20} color="var(--accent-primary)" />
            Generated anchor.yaml
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn" onClick={handleCopy}>
              {copied ? <Check size={16} color="var(--success)" /> : <Copy size={16} />}
              {copied ? 'Copied!' : 'Copy'}
            </button>
            <button className="btn btn-primary" onClick={handleDownload}>
              <Download size={16} /> Download
            </button>
          </div>
        </div>
        
        <div className="code-preview" style={{ flex: 1, margin: 0 }}>
          <pre style={{ margin: 0, height: '100%' }}>
            <code className="language-yaml" style={{ background: 'transparent', padding: 0 }}>
              {yamlContent || '# Start building your pipeline to see the output here...'}
            </code>
          </pre>
        </div>
      </div>
    </div>
  );
}

export default App;
