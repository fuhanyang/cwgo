package eino

const ReActAgentTemplate = `package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/flow/agent/react"
)

type {{.AgentName}} struct {
	agent *react.Agent
}

func New{{.AgentName}}(ctx context.Context) (*{{.AgentName}}, error) {
	// 1. Initialize ChatModel
	// This is a placeholder. In a real scenario, you would initialize the model based on configuration.
	// chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
	// 	Model:  "{{.ModelName}}",                    // Model version
	// 	APIKey: os.Getenv("OPENAI_API_KEY"), // API Key
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// 2. Initialize Tools
	// var toolList []tool.InvokableTool

	// 3. Create ReAct Agent
	// agent, err := react.NewAgent(ctx, &react.AgentConfig{
    //     Model: chatModel,
    //     ToolsConfig: tool.ToolsConfig{
    //         Tools: toolList,
    //     },
    // })
	// if err != nil {
	// 	return nil, err
	// }

	return &{{.AgentName}}{}, nil
}

func (a *{{.AgentName}}) Run(ctx context.Context, query string) (string, error) {
	// result, err := a.agent.Generate(ctx, query)
	return fmt.Sprintf("AI Agent ({{.ModelProvider}}/{{.ModelName}}) response to: %s", query), nil
}
`
