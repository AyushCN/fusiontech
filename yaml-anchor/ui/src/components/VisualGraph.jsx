import { Activity, AlertTriangle, CircleDashed, GitCommitHorizontal, Route, ShieldAlert, Workflow } from 'lucide-react';

function normalizeNeeds(needs) {
  if (!needs) return [];
  if (Array.isArray(needs)) return needs;
  return [needs];
}

function normalizeJobs(pipeline) {
  return Array.isArray(pipeline?.jobs)
    ? pipeline.jobs.map((job) => ({
        ...job,
        runsOn: job.runsOn || job.runs_on || 'ubuntu-latest',
        needs: normalizeNeeds(job.needs),
        steps: job.steps || [],
      }))
    : Object.entries(pipeline?.jobs || {}).map(([id, job]) => ({
        id,
        name: job.name,
        runsOn: job.runsOn || job.runs_on || 'ubuntu-latest',
        needs: normalizeNeeds(job.needs),
        steps: job.steps || [],
      }));
}

function analyzeStep(step) {
  const faults = [];
  const run = step.run || '';
  if (!step.run && !step.uses) faults.push('Missing run or uses');
  if (run.includes('curl') && run.includes('| bash')) faults.push('curl piped to shell');
  if (!step.name || step.name.trim() === '') faults.push('Unnamed step');
  if (/password|secret|token/i.test(run) && !run.includes('secrets.')) faults.push('Possible inline secret');
  return faults;
}

function buildGraph(jobs) {
  const ids = new Set(jobs.map((job) => job.id));
  const missingNeeds = [];
  const jobMeta = new Map();

  jobs.forEach((job) => {
    job.needs.forEach((need) => {
      if (!ids.has(need)) missingNeeds.push(`${job.id} needs missing job ${need}`);
    });
  });

  const getLevel = (job, visiting = new Set()) => {
    if (jobMeta.has(job.id)) return jobMeta.get(job.id).level;
    if (visiting.has(job.id)) return 0;
    visiting.add(job.id);
    const validNeeds = job.needs
      .map((need) => jobs.find((candidate) => candidate.id === need))
      .filter(Boolean);
    const level = validNeeds.length === 0 ? 0 : 1 + Math.max(...validNeeds.map((needJob) => getLevel(needJob, visiting)));
    visiting.delete(job.id);
    jobMeta.set(job.id, { level });
    return level;
  };

  jobs.forEach((job) => getLevel(job));
  const levels = new Map();
  jobs.forEach((job) => {
    const level = jobMeta.get(job.id)?.level || 0;
    if (!levels.has(level)) levels.set(level, []);
    levels.get(level).push(job);
  });

  return { levels, missingNeeds };
}

export default function VisualGraph({ pipeline }) {
  const jobs = normalizeJobs(pipeline);

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

  const { levels, missingNeeds } = buildGraph(jobs);
  const allStepFaults = jobs.flatMap((job) =>
    job.steps.flatMap((step, index) => analyzeStep(step).map((fault) => `${job.id} step ${index + 1}: ${fault}`))
  );
  const actionCount = jobs.reduce((total, job) => total + job.steps.filter((step) => step.uses).length, 0);
  const shellCount = jobs.reduce((total, job) => total + job.steps.filter((step) => step.run).length, 0);
  const maxLevel = Math.max(...levels.keys());
  const levelWidth = 280;
  const rowHeight = 140;
  const nodeWidth = 220;
  const nodeHeight = 88;
  const marginX = 80;
  const marginY = 70;
  const maxRows = Math.max(...Array.from(levels.values()).map((levelJobs) => levelJobs.length));
  const width = Math.max(760, (maxLevel + 1) * levelWidth + marginX * 2);
  const height = Math.max(430, maxRows * rowHeight + marginY * 2);
  const positions = new Map();

  Array.from(levels.entries()).forEach(([level, levelJobs]) => {
    const columnX = marginX + level * levelWidth;
    const columnHeight = (levelJobs.length - 1) * rowHeight;
    const startY = height / 2 - columnHeight / 2;
    levelJobs.forEach((job, index) => {
      positions.set(job.id, {
        x: columnX,
        y: startY + index * rowHeight,
      });
    });
  });

  const issueCount = missingNeeds.length + allStepFaults.length;

  return (
    <div className="panel graph-panel">
      <div className="panel-header">
        <div className="panel-title">
          <Activity size={16} />
          Flow Trace
        </div>
        <span className={`status-pill ${issueCount === 0 ? 'online' : 'offline'}`}>
          {issueCount === 0 ? 'dag ready' : `${issueCount} warnings`}
        </span>
      </div>

      <div className="graph-workbench">
        <div className="graph-sidebar">
          <div className="graph-stat">
            <Workflow size={16} />
            <div>
              <strong>{jobs.length}</strong>
              <span>jobs</span>
            </div>
          </div>
          <div className="graph-stat">
            <GitCommitHorizontal size={16} />
            <div>
              <strong>{actionCount + shellCount}</strong>
              <span>steps</span>
            </div>
          </div>
          <div className="graph-stat">
            <Route size={16} />
            <div>
              <strong>{maxLevel + 1}</strong>
              <span>stages</span>
            </div>
          </div>

          <div className="graph-section">
            <h3>Execution Stages</h3>
            {Array.from(levels.entries()).map(([level, levelJobs]) => (
              <div className="stage-row" key={level}>
                <span>Stage {level + 1}</span>
                <strong>{levelJobs.map((job) => job.id).join(', ')}</strong>
              </div>
            ))}
          </div>

          <div className="graph-section">
            <h3>Preflight</h3>
            {issueCount === 0 ? (
              <p className="graph-ok">No structural warnings found.</p>
            ) : (
              [...missingNeeds, ...allStepFaults].slice(0, 5).map((issue) => (
                <p className="graph-warning" key={issue}>
                  <AlertTriangle size={13} />
                  {issue}
                </p>
              ))
            )}
          </div>
        </div>

        <div className="panel-content graph-container graph-canvas">
          <svg width={width} height={height} style={{ minWidth: '100%', minHeight: '100%' }}>
            <defs>
              <marker id="job-arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="var(--accent-blue)" />
              </marker>
            </defs>

            {jobs.flatMap((job) =>
              job.needs.map((need) => {
                const from = positions.get(need);
                const to = positions.get(job.id);
                if (!from || !to) return null;
                const fromX = from.x + nodeWidth;
                const fromY = from.y + nodeHeight / 2;
                const toX = to.x;
                const toY = to.y + nodeHeight / 2;
                const midX = fromX + (toX - fromX) / 2;
                return (
                  <path
                    key={`${need}-${job.id}`}
                    d={`M ${fromX} ${fromY} C ${midX} ${fromY}, ${midX} ${toY}, ${toX} ${toY}`}
                    fill="none"
                    stroke="var(--accent-blue)"
                    strokeWidth="2"
                    markerEnd="url(#job-arrowhead)"
                    className="path-flow"
                  />
                );
              })
            )}

            {jobs.map((job) => {
              const pos = positions.get(job.id);
              const faults = job.steps.flatMap(analyzeStep);
              const hasIssue = faults.length > 0 || job.needs.some((need) => !positions.has(need));
              return (
                <g key={job.id}>
                  <rect
                    x={pos.x}
                    y={pos.y}
                    width={nodeWidth}
                    height={nodeHeight}
                    rx="8"
                    fill="var(--bg-panel)"
                    stroke={hasIssue ? 'var(--danger)' : 'var(--accent-green)'}
                    strokeWidth="2"
                  />
                  <text x={pos.x + 14} y={pos.y + 25} fill="#fff" fontSize="13" fontWeight="700">
                    {job.id}
                  </text>
                  <text x={pos.x + 14} y={pos.y + 45} fill="var(--text-secondary)" fontSize="10">
                    {job.runsOn}
                  </text>
                  <text x={pos.x + 14} y={pos.y + 67} fill="var(--accent-blue)" fontSize="10">
                    {job.steps.length} steps · {job.steps.filter((step) => step.uses).length} actions · {job.steps.filter((step) => step.run).length} shell
                  </text>
                  {job.needs.length > 0 && (
                    <text x={pos.x + 14} y={pos.y + 82} fill="var(--text-muted)" fontSize="9">
                      needs: {job.needs.join(', ')}
                    </text>
                  )}
                </g>
              );
            })}
          </svg>
        </div>
      </div>

      <div className="summary-strip">
        <span><GitCommitHorizontal size={13} /> {jobs.length} job nodes</span>
        <span><Route size={13} /> {maxLevel + 1} execution stages</span>
        <span><ShieldAlert size={13} /> {issueCount} warnings</span>
      </div>
    </div>
  );
}
