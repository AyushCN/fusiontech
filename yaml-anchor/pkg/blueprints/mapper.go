package blueprints

import (
	"fmt"
	"yaml-anchor/pkg/detector"
)

// MapToYAML returns a recommended anchor.yaml content based on the project profile.
func MapToYAML(profile *detector.ProjectProfile) string {
	name := "My Project"
	if profile.Stack != "" {
		name = fmt.Sprintf("%s Pipeline", profile.Stack)
	}

	yaml := fmt.Sprintf("name: %q\n\n", name)

	if profile.Version != "" {
		yaml += fmt.Sprintf("# Detected version: %s\n", profile.Version)
	}
	if profile.ModuleName != "" {
		yaml += fmt.Sprintf("# Detected module: %s\n", profile.ModuleName)
	}
	if profile.Framework != "" {
		yaml += fmt.Sprintf("# Detected framework: %s\n", profile.Framework)
	}
	if len(profile.InferredScripts) > 0 {
		yaml += fmt.Sprintf("# Available scripts: %v\n", profile.InferredScripts)
	}

	yaml += "\njobs:\n"

	switch profile.Stack {
	case "go":
		yaml += `  build-and-test:
    blueprint: "go-app"
`
	case "node":
		yaml += `  build-and-test:
    blueprint: "node-app"
`
	default:
		// Generic fallback
		yaml += `  main:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Hello"
        run: "echo 'Welcome to YamlAnchor!'"
`
	}

	return yaml
}
