// API service layer — connects the YamlAnchor Studio to the Go backend server
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function parseJsonResponse(response, action) {
  let payload = null;
  try {
    payload = await response.json();
  } catch {
    // Keep the original status error below when the body is empty or invalid.
  }

  if (!response.ok) {
    throw new Error(payload?.error || `${action} failed: ${response.status}`);
  }

  return payload;
}

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
    return parseJsonResponse(response, 'Analyze');
  },

  async generatePipeline(code, fileType = 'go') {
    const response = await fetch(`${API_BASE_URL}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, prompt: code, file_type: fileType }),
    });
    return parseJsonResponse(response, 'Generate');
  },

  async validatePipeline(pipeline) {
    const response = await fetch(`${API_BASE_URL}/api/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pipeline }),
    });
    return parseJsonResponse(response, 'Validate');
  },
};

export const generatePipeline = (input) => api.generatePipeline(input);
export const checkHealth = () => api.checkHealth();
