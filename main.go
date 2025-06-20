package main

import (
	"community-governance-mcp/config"
	"community-governance-mcp/intent"
	"community-governance-mcp/test"
	"community-governance-mcp/tools"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/higress-group/wasm-go/pkg/mcp"
)

func main() {
	// 初始化配置
	cfg := &config.CommunityGovernanceConfig{
		GitHubToken:    os.Getenv("GITHUB_TOKEN"),
		OpenAIKey:      os.Getenv("OPENAI_KEY"),
		ImageAPIKey:    os.Getenv("IMAGE_API_KEY"),
		ImageAPIURL:    "https://api.openai.com/v1/chat/completions",
		KnowledgeDBURL: os.Getenv("KNOWLEDGE_DB_URL"),
		RepoOwner:      "alibaba",
		RepoName:       "higress",
		IntentLLM: config.IntentLLMConfig{
			ServiceName: "dashscope.dns",
			Domain:      "dashscope.aliyuncs.com",
			Port:        443,
			Path:        "/compatible-mode/v1/chat/completions",
			Model:       "qwen-max-0403",
			APIKey:      os.Getenv("OPENAI_KEY"),
			Timeout:     10000,
		},
	}

	// 初始化 MCP 服务器
	mcpServer := mcp.NewMCPServer()
	tools.LoadTools(mcpServer)

	// 初始化意图识别器
	intentRecognizer := intent.NewIntentRecognizer(cfg)

	testServer := &test.TestServer{
		McpServer:        mcpServer,
		Config:           cfg,
		IntentRecognizer: intentRecognizer,
	}

	// 启动 HTTP 服务器
	http.HandleFunc("/chat", testServer.HandleChat)
	http.HandleFunc("/health", testServer.HandleHealth)

	log.Println("Starting community governance agent server on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
