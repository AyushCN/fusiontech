# fusiontech

**fusiontech** is a collection of developer productivity tools and infrastructure utilities built in Go.

---

## Projects

### ⚓ [yaml-anchor](./yaml-anchor)

> Treat CI/CD pipelines as Type-Safe Code, not Indentation-Sensitive Text.

YamlAnchor eliminates "YAML Hell" by letting you define GitHub Actions pipelines in a simple `anchor.yaml` config file. It provides:

- 🔍 **Compile-time validation** — structural errors caught before they reach CI
- 🔐 **Secret scanning** — blocks generation if hardcoded credentials are detected
- 🐳 **Local execution** — runs your pipeline in Docker containers via Dagger, no pushing required
- 🖥️ **Live TUI dashboard** — interactive Pulse dashboard with real-time log streaming

```bash
cd yaml-anchor

# Define your pipeline
cp anchor.yaml /your-project/anchor.yaml
# (edit it)

# Generate .github/workflows/main.yml
anchor generate --config anchor.yaml

# Run the pipeline locally with the live dashboard
anchor local --config anchor.yaml

# Clean up Docker/Dagger caches
anchor clean
```

---

## Getting Started

```bash
git clone https://github.com/ayushcn/fusiontech.git
cd fusiontech/yaml-anchor
go build -o anchor main.go
sudo mv anchor /usr/local/bin/
```

**Prerequisites:** Go 1.21+ and Docker (running locally).
