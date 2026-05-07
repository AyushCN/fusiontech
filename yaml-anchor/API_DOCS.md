# YamlAnchor API Documentation

## Base URL
```
http://localhost:8080
```

Start the server with: `anchor server` or `cd yaml-anchor && go run main.go server`

---

## Endpoints

### `GET /health`

Health check for the API server.

**Response:**
```json
{ "status": "ok", "version": "0.1.0" }
```

---

### `POST /api/analyze`

Analyzes code content and returns language, framework, and CI/CD suggestions.

**Request:**
```json
{
  "code": "package main\nimport \"fmt\"\nfunc main() {}",
  "file_type": "go"
}
```

**Supported `file_type` values:**

| Value | Description |
|:---|:---|
| `go` | Go source file |
| `js`, `jsx`, `ts`, `tsx` | JavaScript / TypeScript |
| `python`, `py` | Python |
| `package.json` | Node.js manifest |
| `dockerfile` | Dockerfile |
| `go.mod` | Go module definition |

**Response:**
```json
{
  "language": "go",
  "framework": "gin",
  "dependencies": {
    "github.com/gin-gonic/gin": "remote"
  },
  "suggestions": [
    "go build, go test, docker build suggested"
  ]
}
```

---

### `POST /api/generate`

Analyzes code and generates a `schema.Pipeline` object ready to use as `anchor.yaml`.

**Request:**
```json
{
  "code": "FROM golang:1.21\nCOPY . .\nRUN go build -o app main.go",
  "file_type": "dockerfile"
}
```

**Response:** A full `schema.Pipeline` JSON object:
```json
{
  "name": "Generated Pipeline",
  "on": { "push": { "branches": ["main"] } },
  "jobs": {
    "build": {
      "name": "Dockerfile Build",
      "runs_on": "ubuntu-latest",
      "steps": [
        { "name": "Checkout", "uses": "actions/checkout@v4" },
        { "name": "Build Image", "run": "docker build -t myapp:latest ." }
      ]
    }
  }
}
```

---

### `POST /api/validate`

Validates a pipeline object against YamlAnchor's schema rules (DAG check, required fields, etc).

**Request:**
```json
{
  "pipeline": { ... } 
}
```

**Response:**
```json
{
  "valid": true,
  "errors": []
}
```

On failure:
```json
{
  "valid": false,
  "errors": ["pipeline must have a name", "job \"build\" must have at least one step"]
}
```

---

## Error Responses

All endpoints return standard HTTP status codes:

| Code | Meaning |
|:---|:---|
| `400 Bad Request` | Missing/invalid JSON body |
| `405 Method Not Allowed` | Wrong HTTP method |
| `500 Internal Server Error` | Server-side failure |

Error format:
```json
{ "error": "description of the error", "status": 400 }
```

---

## CORS

All endpoints include `Access-Control-Allow-Origin: *` headers, enabling use from the Studio UI or any local frontend at a different port.
