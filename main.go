package main

import (
	"github.com/community-governance-mcp-higress/config"
	"github.com/community-governance-mcp-higress/intent"
	"github.com/community-governance-mcp-higress/test"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/higress-group/wasm-go/pkg/mcp"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := &config.CommunityGovernanceConfig{
		GitHubToken:    os.Getenv("GITHUB_TOKEN"),
		OpenAIKey:      os.Getenv("OPENAI_KEY"),
		ImageAPIKey:    os.Getenv("IMAGE_API_KEY"),
		ImageAPIURL:    "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		KnowledgeDBURL: os.Getenv("KNOWLEDGE_DB_URL"),
		RepoOwner:      "alibaba",
		RepoName:       "higress",
		IntentLLM: config.IntentLLMConfig{
			ServiceName: "zhipu",
			Domain:      "https://open.bigmodel.cn",
			Port:        443,
			Path:        "/api/paas/v4/chat/completions",
			Model:       "glm-4v-flash",
			APIKey:      os.Getenv("OPENAI_KEY"),
			Timeout:     10000,
		},
	}

	// 初始化 MCP 服务器
	mcpServer := mcp.NewMCPServer()

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
