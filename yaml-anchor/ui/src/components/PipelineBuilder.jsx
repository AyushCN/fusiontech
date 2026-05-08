import { useState, useEffect } from 'react';
import { Plus, Trash2, Box, Settings, Play, Server, Layers } from 'lucide-react';
import yaml from 'js-yaml';

export default function PipelineBuilder({ onYamlChange }) {
  const [pipeline, setPipeline] = useState({
    name: 'CI Pipeline',
    on: {
      push: { branches: ['main'] },
      pull_request: { branches: ['main'] }
    },
    jobs: [
      {
        id: 'test',
        name: 'Test Job',
        runsOn: 'ubuntu-latest',
        steps: [
          { id: 1, name: 'Checkout Code', uses: 'actions/checkout@v4' },
          { id: 2, name: 'Run Tests', run: 'go test ./...' }
        ]
      }
    ]
  });

  useEffect(() => {
    // Generate valid YamlAnchor configuration
    const yamlObj = {
      name: pipeline.name,
      on: pipeline.on,
      jobs: {}
    };

    pipeline.jobs.forEach(job => {
      yamlObj.jobs[job.id] = {
        'runs-on': job.runsOn,
        steps: job.steps.map(step => {
          const s = { name: step.name };
          if (step.uses) s.uses = step.uses;
          if (step.run) s.run = step.run;
          if (step.env) s.env = step.env;
          return s;
        })
      };
    });

    try {
      const yamlStr = yaml.dump(yamlObj, { lineWidth: -1 });
      onYamlChange(yamlStr);
    } catch (e) {
      console.error('Failed to generate YAML', e);
    }
  }, [pipeline, onYamlChange]);

  const addJob = () => {
    const newJobId = `job_${Date.now()}`;
    setPipeline(prev => ({
      ...prev,
      jobs: [
        ...prev.jobs,
        {
          id: newJobId,
          name: 'New Job',
          runsOn: 'ubuntu-latest',
          steps: []
        }
      ]
    }));
  };

  const removeJob = (jobId) => {
    setPipeline(prev => ({
      ...prev,
      jobs: prev.jobs.filter(j => j.id !== jobId)
    }));
  };

  const updateJob = (jobId, field, value) => {
    setPipeline(prev => ({
      ...prev,
      jobs: prev.jobs.map(j => j.id === jobId ? { ...j, [field]: value } : j)
    }));
  };

  const addStep = (jobId) => {
    setPipeline(prev => ({
      ...prev,
      jobs: prev.jobs.map(j => {
        if (j.id === jobId) {
          return {
            ...j,
            steps: [...j.steps, { id: Date.now(), name: 'New Step', run: 'echo "Hello World"' }]
          };
        }
        return j;
      })
    }));
  };

  const removeStep = (jobId, stepId) => {
    setPipeline(prev => ({
      ...prev,
      jobs: prev.jobs.map(j => {
        if (j.id === jobId) {
          return { ...j, steps: j.steps.filter(s => s.id !== stepId) };
        }
        return j;
      })
    }));
  };

  const updateStep = (jobId, stepId, field, value) => {
    setPipeline(prev => ({
      ...prev,
      jobs: prev.jobs.map(j => {
        if (j.id === jobId) {
          return {
            ...j,
            steps: j.steps.map(s => s.id === stepId ? { ...s, [field]: value } : s)
          };
        }
        return j;
      })
    }));
  };

  return (
    <div className="builder-section">
      <div className="section-header">
        <Layers className="logo-icon" size={24} />
        Pipeline Configuration
      </div>
      <div className="scroll-area">
        
        <div className="card animate-fade-in">
          <div className="card-header">
            <h3 className="card-title">Global Settings</h3>
            <Settings size={18} color="var(--text-secondary)" />
          </div>
          <div className="form-group">
            <label className="form-label">Pipeline Name</label>
            <input 
              type="text" 
              className="form-input" 
              value={pipeline.name}
              onChange={e => setPipeline(p => ({ ...p, name: e.target.value }))}
            />
          </div>
        </div>

        <div className="section-header" style={{ marginTop: '2rem' }}>
          <Server size={20} />
          Jobs
        </div>

        {pipeline.jobs.length === 0 ? (
          <div className="empty-state">
            <Box size={48} />
            <p>No jobs defined.</p>
            <button className="btn btn-primary" onClick={addJob} style={{ marginTop: '1rem' }}>
              <Plus size={16} /> Add Job
            </button>
          </div>
        ) : (
          pipeline.jobs.map((job) => (
            <div key={job.id} className="card animate-fade-in">
              <div className="card-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Box size={18} color="var(--accent-primary)" />
                  <input 
                    type="text" 
                    className="form-input" 
                    value={job.id}
                    onChange={e => updateJob(job.id, 'id', e.target.value)}
                    style={{ padding: '0.25rem 0.5rem', width: 'auto' }}
                  />
                </div>
                <button className="btn btn-danger btn-icon" onClick={() => removeJob(job.id)}>
                  <Trash2 size={16} />
                </button>
              </div>

              <div className="form-group">
                <label className="form-label">Runs On (Docker Image)</label>
                <input 
                  type="text" 
                  className="form-input" 
                  value={job.runsOn}
                  onChange={e => updateJob(job.id, 'runsOn', e.target.value)}
                />
              </div>

              <div className="step-list">
                <h4 style={{ margin: '1rem 0 0.5rem 0', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>Steps</h4>
                
                {job.steps.map((step, idx) => (
                  <div key={step.id} className="step-item animate-fade-in" style={{ animationDelay: `${idx * 0.1}s` }}>
                    <div className="step-header">
                      <div className="step-title">
                        <Play size={14} color="var(--accent-secondary)" />
                        Step {idx + 1}
                      </div>
                      <button className="btn btn-danger btn-icon" style={{ border: 'none' }} onClick={() => removeStep(job.id, step.id)}>
                        <Trash2 size={14} />
                      </button>
                    </div>
                    
                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <input 
                        type="text" 
                        className="form-input" 
                        placeholder="Step Name"
                        value={step.name}
                        onChange={e => updateStep(job.id, step.id, 'name', e.target.value)}
                        style={{ marginBottom: '0.5rem' }}
                      />
                      
                      {step.uses !== undefined ? (
                        <input 
                          type="text" 
                          className="form-input" 
                          placeholder="Uses (e.g. actions/checkout@v4)"
                          value={step.uses}
                          onChange={e => updateStep(job.id, step.id, 'uses', e.target.value)}
                        />
                      ) : (
                        <input 
                          type="text" 
                          className="form-input" 
                          placeholder="Run (e.g. go test ./...)"
                          value={step.run || ''}
                          onChange={e => updateStep(job.id, step.id, 'run', e.target.value)}
                        />
                      )}
                    </div>
                  </div>
                ))}
                
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                  <button className="btn" onClick={() => addStep(job.id)}>
                    <Plus size={14} /> Add Run Step
                  </button>
                  <button className="btn" onClick={() => {
                    setPipeline(prev => ({
                      ...prev,
                      jobs: prev.jobs.map(j => j.id === job.id ? {
                        ...j, steps: [...j.steps, { id: Date.now(), name: 'Action Step', uses: 'actions/setup-node@v3' }]
                      } : j)
                    }));
                  }}>
                    <Plus size={14} /> Add Uses Step
                  </button>
                </div>
              </div>
            </div>
          ))
        )}

        {pipeline.jobs.length > 0 && (
          <button className="btn btn-primary" onClick={addJob} style={{ width: '100%', justifyContent: 'center' }}>
            <Plus size={18} /> Add Another Job
          </button>
        )}
      </div>
    </div>
  );
}
