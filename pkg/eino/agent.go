package eino

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cloudwego/cwgo/config"
)

type AgentTemplateData struct {
	GoModule      string
	AgentName     string
	ModelProvider string
	ModelName     string
	Tools         []string
	EnableRAG     bool
}

func GenerateEinoAgentModule(c *config.ServerArgument) error {
	agentDir := filepath.Join(c.OutDir, "internal", "agent")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return err
	}

	switch c.AgentType {
	case "react":
		return generateReActAgent(c, agentDir)
	case "rag":
		// Placeholder for RAG
		return generateReActAgent(c, agentDir)
	case "multi-agent":
		// Placeholder for Multi-Agent
		return generateReActAgent(c, agentDir)
	default:
		return generateReActAgent(c, agentDir)
	}
}

func generateReActAgent(c *config.ServerArgument, outDir string) error {
	serviceName := c.ServerName
	if serviceName == "" {
		serviceName = "Service"
	}

	data := AgentTemplateData{
		GoModule:      c.GoMod,
		AgentName:     toCamel(serviceName) + "Agent",
		ModelProvider: c.ModelProvider,
		ModelName:     c.ModelName,
		Tools:         c.EnableTools,
		EnableRAG:     c.EnableRAG,
	}

	tmpl, err := template.New("agent").Parse(ReActAgentTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	filename := filepath.Join(outDir, "agent.go")
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

func toCamel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}
