package tools

import (
	"community-governance-mcp/config"
	mcp_tool "community-governance-mcp/utils"
	"encoding/json"
	"fmt"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
	"github.com/higress-group/wasm-go/pkg/mcp/utils"
)

type ImageAnalyzer struct {
	ImageURL     string `json:"image_url" jsonschema_description:"图片URL地址"`
	ImageBase64  string `json:"image_base64,omitempty" jsonschema_description:"图片Base64编码"`
	Context      string `json:"context,omitempty" jsonschema_description:"图片相关的文字描述或上下文"`
	AnalysisType string `json:"analysis_type" jsonschema_description:"分析类型：error_screenshot, architecture_diagram, log_image, general" jsonschema:"example=error_screenshot"`
}

func (t ImageAnalyzer) Description() string {
	return `图片分析工具，支持分析错误截图、架构图、日志图片等，结合文字上下文提供智能诊断。`
}

func (t ImageAnalyzer) InputSchema() map[string]any {
	return server.ToInputSchema(&ImageAnalyzer{})
}

func (t ImageAnalyzer) Create(params []byte) server.Tool {
	analyzer := &ImageAnalyzer{}
	json.Unmarshal(params, analyzer)
	return analyzer
}

func (t ImageAnalyzer) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	if serverConfig.ImageAPIKey == "" {
		return fmt.Errorf("图片分析 API 密钥未配置")
	}

	// 根据分析类型选择不同的分析策略
	var analysis string
	var err error

	switch t.AnalysisType {
	case "error_screenshot":
		analysis, err = t.analyzeErrorScreenshot(ctx, serverConfig)
	case "architecture_diagram":
		analysis, err = t.analyzeArchitectureDiagram(ctx, serverConfig)
	case "log_image":
		analysis, err = t.analyzeLogImage(ctx, serverConfig)
	default:
		analysis, err = t.generalImageAnalysis(ctx, serverConfig)
	}

	if err != nil {
		return err
	}

	// 结合上下文信息
	if t.Context != "" {
		analysis = fmt.Sprintf("## 上下文信息\n%s\n\n%s", t.Context, analysis)
	}

	utils.SendMCPToolTextResult(ctx, analysis)
	return nil
}

func (t ImageAnalyzer) analyzeErrorScreenshot(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	prompt := `这是一个错误截图。请分析图片中的错误信息，包括：  
1. 错误类型和错误代码  
2. 可能的原因分析  
3. 建议的解决步骤  
4. 相关的调试方法  
  
请提供详细的分析结果。`

	return t.callVisionAPI(ctx, config, prompt)
}

func (t ImageAnalyzer) analyzeArchitectureDiagram(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	prompt := `这是一个系统架构图。请分析图片中的架构设计，包括：  
1. 系统组件和模块  
2. 数据流和调用关系  
3. 可能的性能瓶颈  
4. 架构优化建议  
5. 潜在的问题点  
  
请提供详细的架构分析。`

	return t.callVisionAPI(ctx, config, prompt)
}

func (t ImageAnalyzer) analyzeLogImage(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	prompt := `这是一个日志截图。请分析图片中的日志信息，包括：  
1. 提取关键的日志内容  
2. 识别错误和警告信息  
3. 分析日志模式和趋势  
4. 提供问题诊断建议  
  
请提供详细的日志分析结果。`

	return t.callVisionAPI(ctx, config, prompt)
}

func (t ImageAnalyzer) generalImageAnalysis(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	prompt := `请分析这张图片的内容，特别关注技术相关的信息，包括：  
1. 图片中的文字内容  
2. 技术组件或界面元素  
3. 可能的问题或异常  
4. 相关的技术建议  
  
请提供详细的分析结果。`

	return t.callVisionAPI(ctx, config, prompt)
}

func (t ImageAnalyzer) callVisionAPI(ctx server.HttpContext, config *config.CommunityGovernanceConfig, prompt string) (string, error) {
	var imageContent string
	if t.ImageBase64 != "" {
		imageContent = t.ImageBase64
	} else if t.ImageURL != "" {
		imageContent = t.ImageURL
	} else {
		return "", fmt.Errorf("未提供图片URL或Base64编码")
	}

	requestBody := map[string]interface{}{
		//换成智谱的意图理解
		"model": "glm-4v-flash",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageContent,
						},
					},
				},
			},
		},
		"max_tokens": 500,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	headers := map[string]string{
		"Authorization": "Bearer " + config.OpenAIKey,
		"Content-Type":  "application/json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "POST", "https://api.openai.com/v1/chat/completions", headers, string(bodyBytes))
	if err != nil {
		return "", err
	}

	var aiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		return "", err
	}

	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("AI 响应格式错误")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return fmt.Sprintf("# 图片分析结果\n\n%s", content), nil
}
