import { Activity, CircleDashed, GitCommitHorizontal, ShieldAlert } from 'lucide-react';

export default function VisualGraph({ pipeline }) {
  const jobs = Array.isArray(pipeline?.jobs)
    ? pipeline.jobs
    : Object.entries(pipeline?.jobs || {}).map(([id, job]) => ({
        id,
        runsOn: job.runsOn || job.runs_on || 'ubuntu-latest',
        steps: job.steps || [],
      }));

  if (!pipeline || jobs.length === 0) {
    return (
      <div className="panel graph-panel">
        <div className="panel-header">
          <div className="panel-title">
            <Activity size={16} />
            Flow Trace
          </div>
        </div>
        <div className="graph-container">
          <div className="empty-state">
            <CircleDashed size={44} />
            <span>Generate a pipeline to inspect the DAG</span>
          </div>
        </div>
      </div>
    );
  }

  // Fault Detection Logic
  const analyzeFaults = (step) => {
    const faults = [];
    if (!step.run && !step.uses) {
      faults.push('Missing "run" or "uses" command');
    }
    if (step.run && step.run.includes('curl') && step.run.includes('| bash')) {
      faults.push('Security Warning: Piping curl to bash is dangerous');
    }
    if (!step.name || step.name.trim() === '') {
      faults.push('Step missing descriptive name');
    }
    return faults;
  };

  // Simple layout calculation
  const startX = 300;
  const startY = 50;
  const jobSpacingX = 250;
  const stepSpacingY = 80;

  // Render SVG nodes
  return (
    <div className="panel graph-panel">
      <div className="panel-header">
        <div className="panel-title">
          <Activity size={16} />
          Flow Trace
        </div>
        <span className="status-pill online">dag ready</span>
      </div>
      
      <div className="panel-content graph-container" style={{ overflow: 'auto' }}>
        {/* Make SVG large enough to scroll if many jobs */}
        <svg width={Math.max(600, jobs.length * jobSpacingX + 100)} height={800} style={{ minWidth: '100%', minHeight: '100%' }}>
          <defs>
            <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
              <polygon points="0 0, 10 3.5, 0 7" fill="var(--text-secondary)" />
            </marker>
            <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="4" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
          </defs>

          {/* Root Node */}
          <rect x={startX - 90} y={startY} width="180" height="44" rx="6" fill="var(--bg-panel)" stroke="var(--accent-blue)" strokeWidth="2" filter="url(#glow)"/>
          <text x={startX} y={startY + 27} fill="#fff" textAnchor="middle" fontSize="12" fontWeight="bold">{pipeline.name || 'Generated Pipeline'}</text>

          {jobs.map((job, jIdx) => {
            // Calculate job X position centering around startX
            const totalWidth = (jobs.length - 1) * jobSpacingX;
            const jobX = startX - (totalWidth / 2) + (jIdx * jobSpacingX);
            const jobY = startY + 100;

            const jobHasNoSteps = !job.steps || job.steps.length === 0;

            return (
              <g key={job.id}>
                {/* Edge from Root to Job */}
                <path 
                  d={`M ${startX} ${startY + 40} C ${startX} ${jobY - 30}, ${jobX} ${startY + 60}, ${jobX} ${jobY}`} 
                  fill="none" 
                  stroke="var(--text-secondary)" 
                  strokeWidth="2" 
                  markerEnd="url(#arrowhead)"
                  className="path-flow"
                />

                {/* Job Node */}
                <rect x={jobX - 90} y={jobY} width="180" height="50" rx="6" fill="var(--bg-panel)" stroke={jobHasNoSteps ? "var(--danger)" : "var(--accent-amber)"} strokeWidth="2" />
                <text x={jobX} y={jobY + 20} fill="#fff" textAnchor="middle" fontSize="12" fontWeight="bold">{job.id}</text>
                <text x={jobX} y={jobY + 38} fill="var(--text-secondary)" textAnchor="middle" fontSize="10">{job.runsOn}</text>

                {jobHasNoSteps && (
                  <text x={jobX} y={jobY + 65} fill="var(--danger)" textAnchor="middle" fontSize="10">Fault: No steps defined!</text>
                )}

                {/* Steps Nodes */}
                {job.steps && job.steps.map((step, sIdx) => {
                  const stepX = jobX;
                  const stepY = jobY + 90 + (sIdx * stepSpacingY);
                  const faults = analyzeFaults(step);
                  const isFaulty = faults.length > 0;
                  const isUses = !!step.uses;
                  
                  const strokeColor = isFaulty ? "var(--danger)" : (isUses ? "var(--accent-green)" : "var(--accent-blue)");

                  return (
                    <g key={step.id || sIdx}>
                      {/* Edge from previous node to this step */}
                      <path 
                        d={`M ${stepX} ${sIdx === 0 ? jobY + 50 : stepY - stepSpacingY + 40} L ${stepX} ${stepY}`} 
                        fill="none" 
                        stroke="var(--text-secondary)" 
                        strokeWidth="2" 
                        markerEnd="url(#arrowhead)"
                      />

                      {/* Step Node */}
                      <rect x={stepX - 85} y={stepY} width="170" height="42" rx="6" fill="var(--bg-input)" stroke={strokeColor} strokeWidth="2" />
                      
                      {/* Icon Status */}
                      {isFaulty ? (
                        <circle cx={stepX - 70} cy={stepY + 20} r="6" fill="var(--danger)" />
                      ) : (
                        <circle cx={stepX - 70} cy={stepY + 20} r="6" fill="var(--accent-green)" />
                      )}

                      <text x={stepX - 55} y={stepY + 24} fill="#fff" fontSize="11">{step.name || 'Unnamed Step'}</text>

                      {/* Fault Annotations */}
                      {isFaulty && (
                        <g>
                          <rect x={stepX + 95} y={stepY} width="140" height={faults.length * 15 + 10} rx="4" fill="rgba(239, 68, 68, 0.1)" stroke="var(--danger)" strokeWidth="1" strokeDasharray="2,2"/>
                          {faults.map((f, fIdx) => (
                            <text key={fIdx} x={stepX + 100} y={stepY + 15 + (fIdx * 15)} fill="var(--danger)" fontSize="9">! {f}</text>
                          ))}
                          {/* Pointer line */}
                          <line x1={stepX + 85} y1={stepY + 20} x2={stepX + 95} y2={stepY + 20} stroke="var(--danger)" strokeWidth="1" strokeDasharray="2,2" />
                        </g>
                      )}
                    </g>
                  );
                })}
              </g>
            );
          })}
        </svg>
      </div>
      <div className="summary-strip">
        <span><GitCommitHorizontal size={13} /> {jobs.length} job nodes</span>
        <span><ShieldAlert size={13} /> inline fault scan</span>
      </div>
    </div>
  );
}
