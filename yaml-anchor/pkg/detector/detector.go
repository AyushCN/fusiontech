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
	Version         string
	ModuleName      string
	Framework       string
	HasDocker       bool
	HasExistingCI   bool
	InferredScripts []string
	Root            string
}

// Detect scans the directory and returns a ProjectProfile.
func Detect(root string) (*ProjectProfile, error) {
	profile := &ProjectProfile{
		Root: root,
	}

	// 1. Detect Go
	if exists(filepath.Join(root, "go.mod")) {
		profile.Stack = "go"
		parseGoMod(filepath.Join(root, "go.mod"), profile)
	} else if exists(filepath.Join(root, "package.json")) {
		// 2. Detect Node
		profile.Stack = "node"
		parsePackageJSON(filepath.Join(root, "package.json"), profile)
		if profile.Version == "" && exists(filepath.Join(root, ".nvmrc")) {
			if data, err := os.ReadFile(filepath.Join(root, ".nvmrc")); err == nil {
				profile.Version = strings.TrimSpace(string(data))
			}
		}
	} else if exists(filepath.Join(root, "requirements.txt")) || exists(filepath.Join(root, "pyproject.toml")) {
		// 3. Detect Python
		profile.Stack = "python"
		profile.InferredScripts = []string{"install", "test"}
	} else if exists(filepath.Join(root, "Cargo.toml")) {
		// 4. Detect Rust
		profile.Stack = "rust"
		profile.InferredScripts = []string{"build", "test"}
	}

	// 5. Detect Docker
	if exists(filepath.Join(root, "Dockerfile")) {
		profile.HasDocker = true
	}

	// 6. Detect existing CI
	if exists(filepath.Join(root, ".github", "workflows")) {
		profile.HasExistingCI = true
	}

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
		Name    string `json:"name"`
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
			if _, ok := pkg.Scripts[scriptName]; ok {
				profile.InferredScripts = append(profile.InferredScripts, scriptName)
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
