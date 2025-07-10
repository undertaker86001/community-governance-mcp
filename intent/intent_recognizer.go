package intent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/community-governance-mcp-higress/config"
)

type IntentRecognizer struct {
	config *config.CommunityGovernanceConfig
}

type IntentRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type IntentResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type IntentResult struct {
	Intent     string  `json:"intent"`
	ToolName   string  `json:"tool_name"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func NewIntentRecognizer(cfg *config.CommunityGovernanceConfig) *IntentRecognizer {
	return &IntentRecognizer{
		config: cfg,
	}
}

func (ir *IntentRecognizer) RecognizeIntent(message, imageURL, context string) (*IntentResult, error) {
	// 构建意图识别的提示词
	prompt := ir.buildIntentPrompt(message, imageURL, context)

	// 调用LLM进行意图识别
	llmResponse, err := ir.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %v", err)
	}

	// 解析LLM响应
	intentResult, err := ir.parseIntentResponse(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("解析意图响应失败: %v", err)
	}

	return intentResult, nil
}

func (ir *IntentRecognizer) buildIntentPrompt(message, imageURL, context string) string {
	prompt := `你是Higress社区治理助手的意图识别模块。请分析用户的输入，识别其意图并选择最合适的工具。  
  
可用的工具和对应意图：  
1. github_manager - GitHub仓库管理（创建/更新Issue、PR管理、仓库操作）  
2. issue_classifier - Issue智能分类（自动标签推荐、问题类型识别）  
3. bug_analyzer - Bug分析诊断（错误分析、解决方案推荐）  
4. image_analyzer - 图片分析（错误截图、架构图、日志图片分析）  
5. knowledge_base - 知识库搜索（文档查询、最佳实践、历史问题搜索）  
6. community_stats - 社区统计（活跃度报告、贡献者分析、项目健康度）  
  
用户输入：  
消息: %s`

	if imageURL != "" {
		prompt += fmt.Sprintf("\n图片URL: %s", imageURL)
	}

	if context != "" {
		prompt += fmt.Sprintf("\n上下文: %s", context)
	}

	prompt += `  
  
请以JSON格式返回结果，包含以下字段：  
{  
  "intent": "具体的意图描述",  
  "tool_name": "选择的工具名称",  
  "confidence": 0.95,  
  "reasoning": "选择该工具的理由"  
}  
  
注意：  
- confidence为0-1之间的数值，表示识别的置信度  
- 如果有图片URL，优先考虑image_analyzer工具  
- 如果提到错误、异常、堆栈等，优先考虑bug_analyzer工具  
- 如果需要GitHub操作，选择github_manager工具  
- 如果需要搜索信息，选择knowledge_base工具`

	return fmt.Sprintf(prompt, message)
}

func (ir *IntentRecognizer) callLLM(prompt string) (string, error) {
	// 构建请求
	request := IntentRequest{
		Model: ir.config.IntentLLM.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// 这里需要实现HTTP调用，参考ai-intent插件的实现方式
	// 由于在独立进程模式下，我们使用标准的HTTP客户端
	return ir.makeHTTPRequest(string(requestBody))
}

func (ir *IntentRecognizer) makeHTTPRequest(requestBody string) (string, error) {

	// 构建完整的请求URL
	url := fmt.Sprintf("https://%s%s", ir.config.IntentLLM.Domain, ir.config.IntentLLM.Path)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(ir.config.IntentLLM.Timeout) * time.Millisecond,
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ir.config.IntentLLM.APIKey)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API 调用失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func (ir *IntentRecognizer) parseIntentResponse(response string) (*IntentResult, error) {
	var llmResp IntentResponse
	if err := json.Unmarshal([]byte(response), &llmResp); err != nil {
		return nil, err
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM响应为空")
	}

	content := llmResp.Choices[0].Message.Content

	// 解析JSON格式的意图结果
	var result IntentResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
