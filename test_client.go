package main

import (
	"bytes"
	"community-governance-mcp/test"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestClient struct {
	baseURL string
	client  *http.Client
}

type TestCase struct {
	Name     string
	Message  string
	ImageURL string
	Context  string
	Expected string
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (tc *TestClient) sendMessage(message, imageURL, context string) (*test.ChatResponse, error) {
	req := test.ChatRequest{
		Message:  message,
		ImageURL: imageURL,
		Context:  context,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := tc.client.Post(tc.baseURL+"/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var chatResp test.ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}

func main() {
	client := NewTestClient("http://localhost:8080")

	testCases := []TestCase{
		{
			Name:     "Bug 报告测试",
			Message:  "我遇到了一个错误，网关启动失败，出现了空指针异常",
			Expected: "bug_analysis",
		},
		{
			Name:     "图片错误分析",
			Message:  "这个错误截图显示了什么问题？",
			ImageURL: "https://example.com/error-screenshot.png",
			Context:  "Higress 网关启动时的错误",
			Expected: "image_bug_analysis",
		},
		{
			Name:     "Issue 分类",
			Message:  "请帮我分类这个 Issue：新增支持 gRPC 协议的功能",
			Expected: "issue_classification",
		},
		{
			Name:     "GitHub 管理",
			Message:  "请列出所有开放的 Issues",
			Expected: "github_management",
		},
		{
			Name:     "社区统计",
			Message:  "生成一份社区活跃度报告",
			Expected: "community_stats",
		},
		{
			Name:     "知识库搜索",
			Message:  "如何配置 Higress 的 AI 插件？",
			Expected: "knowledge_search",
		},
	}

	fmt.Println("开始测试社区治理 Agent...")

	for i, testCase := range testCases {
		fmt.Printf("\n测试 %d: %s\n", i+1, testCase.Name)
		fmt.Printf("输入: %s\n", testCase.Message)
		if testCase.ImageURL != "" {
			fmt.Printf("图片: %s\n", testCase.ImageURL)
		}
		if testCase.Context != "" {
			fmt.Printf("上下文: %s\n", testCase.Context)
		}

		resp, err := client.sendMessage(testCase.Message, testCase.ImageURL, testCase.Context)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		fmt.Printf("识别意图: %s\n", resp.Intent)
		fmt.Printf("使用工具: %s\n", resp.ToolUsed)
		fmt.Printf("置信度: %.2f\n", resp.Confidence)
		fmt.Printf("响应: %s\n", resp.Response)

		// 验证结果
		if resp.Intent == testCase.Expected {
			fmt.Printf("✅ 测试通过\n")
		} else {
			fmt.Printf("❌ 测试失败 - 期望: %s, 实际: %s\n", testCase.Expected, resp.Intent)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n测试完成！")
}
