// API service layer — connects the YamlAnchor Studio to the Go backend server
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = {
  async checkHealth() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`, { signal: AbortSignal.timeout(2000) });
      return response.ok;
    } catch {
      return false;
    }
  },

  async analyzeCode(code, fileType = 'go') {
    const response = await fetch(`${API_BASE_URL}/api/analyze`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, file_type: fileType }),
    });
    if (!response.ok) throw new Error(`Analyze failed: ${response.status}`);
    return response.json();
  },

  async generatePipeline(code, fileType = 'go') {
    const response = await fetch(`${API_BASE_URL}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, file_type: fileType }),
    });
    if (!response.ok) throw new Error(`Generate failed: ${response.status}`);
    return response.json();
  },

  async validatePipeline(pipeline) {
    const response = await fetch(`${API_BASE_URL}/api/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pipeline }),
    });
    if (!response.ok) throw new Error(`Validate failed: ${response.status}`);
    return response.json();
  },
};
