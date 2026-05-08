package detector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectProfile represents the detected stack of a project.
type ProjectProfile struct {
	Stack           string
	Stacks          []string
	Version         string
	ModuleName      string
	Framework       string
	HasGo           bool
	HasNode         bool
	HasPython       bool
	HasRust         bool
	HasDocker       bool
	HasExistingCI   bool
	InferredScripts []string
	ScriptCommands  map[string]string
	ProjectTree     []string
	ContextFiles    map[string]string
	ExistingCI      []string
	Root            string
}

// Detect scans the directory and returns a ProjectProfile.
func Detect(root string) (*ProjectProfile, error) {
	profile := &ProjectProfile{
		Root:           root,
		ScriptCommands: make(map[string]string),
		ContextFiles:   make(map[string]string),
	}

	if exists(filepath.Join(root, "go.mod")) {
		profile.HasGo = true
		profile.Stacks = append(profile.Stacks, "go")
		parseGoMod(filepath.Join(root, "go.mod"), profile)
		addContextFile(root, profile, "go.mod")
	}

	if exists(filepath.Join(root, "package.json")) {
		profile.HasNode = true
		profile.Stacks = append(profile.Stacks, "node")
		parsePackageJSON(filepath.Join(root, "package.json"), profile)
		addContextFile(root, profile, "package.json")
		if profile.Version == "" && exists(filepath.Join(root, ".nvmrc")) {
			if data, err := os.ReadFile(filepath.Join(root, ".nvmrc")); err == nil {
				profile.Version = strings.TrimSpace(string(data))
			}
		}
	}

	if exists(filepath.Join(root, "requirements.txt")) || exists(filepath.Join(root, "pyproject.toml")) {
		profile.HasPython = true
		profile.Stacks = append(profile.Stacks, "python")
		if len(profile.InferredScripts) == 0 {
			profile.InferredScripts = []string{"install", "test"}
		}
		addContextFile(root, profile, "requirements.txt")
		addContextFile(root, profile, "pyproject.toml")
	}

	if exists(filepath.Join(root, "Cargo.toml")) {
		profile.HasRust = true
		profile.Stacks = append(profile.Stacks, "rust")
		profile.InferredScripts = []string{"build", "test"}
		addContextFile(root, profile, "Cargo.toml")
	}

	if exists(filepath.Join(root, "Dockerfile")) {
		profile.HasDocker = true
		addContextFile(root, profile, "Dockerfile")
	}

	workflowDir := filepath.Join(root, ".github", "workflows")
	if exists(workflowDir) {
		profile.HasExistingCI = true
		entries, _ := os.ReadDir(workflowDir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
				rel := filepath.Join(".github", "workflows", name)
				profile.ExistingCI = append(profile.ExistingCI, rel)
				addContextFile(root, profile, rel)
			}
		}
	}

	if len(profile.Stacks) > 0 {
		profile.Stack = profile.Stacks[0]
	}
	profile.ProjectTree = collectProjectTree(root, 80)

	return profile, nil
}

func parseGoMod(path string, profile *ProjectProfile) {
	file, err := os.Open(path)
	if err != nil {
		profile.InferredScripts = []string{"build", "test"}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			profile.ModuleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		} else if strings.HasPrefix(line, "go ") {
			profile.Version = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	profile.InferredScripts = []string{"build", "test"}
}

func parsePackageJSON(path string, profile *ProjectProfile) {
	data, err := os.ReadFile(path)
	if err != nil {
		profile.InferredScripts = []string{"install", "build", "test"}
		return
	}

	var pkg struct {
		Name    string            `json:"name"`
		Scripts map[string]string `json:"scripts"`
		Engines map[string]string `json:"engines"`
		Deps    map[string]string `json:"dependencies"`
		DevDeps map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err == nil {
		profile.ModuleName = pkg.Name

		if nodeVer, ok := pkg.Engines["node"]; ok {
			profile.Version = nodeVer
		}

		// Detect Framework
		if _, ok := pkg.Deps["next"]; ok {
			profile.Framework = "Next.js"
		} else if _, ok := pkg.Deps["express"]; ok {
			profile.Framework = "Express"
		} else if _, ok := pkg.Deps["react"]; ok {
			profile.Framework = "React"
		}

		// Extract available scripts
		for _, scriptName := range []string{"lint", "test", "build"} {
			if command, ok := pkg.Scripts[scriptName]; ok {
				profile.InferredScripts = append(profile.InferredScripts, scriptName)
				profile.ScriptCommands[scriptName] = command
			}
		}
	}

	if len(profile.InferredScripts) == 0 {
		profile.InferredScripts = []string{"install", "build", "test"}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func addContextFile(root string, profile *ProjectProfile, rel string) {
	if rel == "" {
		return
	}
	path := filepath.Join(root, rel)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() > 64*1024 {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	profile.ContextFiles[filepath.ToSlash(rel)] = string(data)
}

func collectProjectTree(root string, limit int) []string {
	var tree []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || len(tree) >= limit {
			return nil
		}
		name := d.Name()
		if d.IsDir() && shouldSkipDir(name) {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			rel += "/"
		}
		tree = append(tree, rel)
		return nil
	})
	return tree
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "dist", "build", "coverage", ".next", ".venv", "vendor":
		return true
	default:
		return false
	}
}
