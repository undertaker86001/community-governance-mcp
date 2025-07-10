package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TestAgent 测试Agent功能
func TestAgent() {
	// 测试用例
	testCases := []struct {
		name    string
		request map[string]interface{}
	}{
		{
			name: "Issue测试",
			request: map[string]interface{}{
				"type":     "issue",
				"title":    "Gateway配置问题",
				"content":  "我在配置Higress Gateway时遇到了路由问题，具体错误是：404 Not Found。",
				"author":   "test-user",
				"priority": "medium",
				"tags":     []string{"gateway", "routing", "404"},
			},
		},
		{
			name: "PR测试",
			request: map[string]interface{}{
				"type":     "pr",
				"title":    "Add new plugin feature",
				"content":  "我添加了一个新的插件功能，包括配置验证和错误处理。",
				"author":   "test-contributor",
				"priority": "medium",
				"tags":     []string{"plugin", "feature"},
			},
		},
		{
			name: "图文问题测试",
			request: map[string]interface{}{
				"type":     "text",
				"title":    "Kubernetes部署问题",
				"content":  "我想了解如何在Kubernetes中部署Higress，需要哪些配置文件和步骤？",
				"author":   "k8s-user",
				"priority": "low",
				"tags":     []string{"kubernetes", "deployment"},
			},
		},
	}

	// 运行测试
	for _, tc := range testCases {
		fmt.Printf("\n=== 测试: %s ===\n", tc.name)
		
		// 发送请求
		response, err := sendRequest(tc.request)
		if err != nil {
			fmt.Printf("❌ 测试失败: %v\n", err)
			continue
		}
		
		// 打印结果
		fmt.Printf("✅ 测试成功\n")
		fmt.Printf("响应ID: %s\n", response["id"])
		fmt.Printf("处理时间: %s\n", response["processing_time"])
		fmt.Printf("置信度: %.2f\n", response["confidence"])
		fmt.Printf("融合分数: %.2f\n", response["fusion_score"])
		
		// 打印建议
		if recommendations, ok := response["recommendations"].([]interface{}); ok {
			fmt.Printf("建议:\n")
			for i, rec := range recommendations {
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}
	}
}

// sendRequest 发送请求到Agent
func sendRequest(request map[string]interface{}) (map[string]interface{}, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/process", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}
	
	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	return response, nil
}

// TestHealth 测试健康检查
func TestHealth() {
	fmt.Println("\n=== 健康检查测试 ===")
	
	resp, err := http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ 健康检查通过")
	} else {
		fmt.Printf("❌ 健康检查失败，状态码: %d\n", resp.StatusCode)
	}
}

// TestConfig 测试配置接口
func TestConfig() {
	fmt.Println("\n=== 配置信息测试 ===")
	
	resp, err := http.Get("http://localhost:8080/api/v1/config")
	if err != nil {
		fmt.Printf("❌ 配置信息获取失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		var config map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
			fmt.Printf("❌ 解析配置信息失败: %v\n", err)
			return
		}
		
		fmt.Println("✅ 配置信息获取成功")
		fmt.Printf("Agent名称: %s\n", config["name"])
		fmt.Printf("版本: %s\n", config["version"])
		fmt.Printf("端口: %.0f\n", config["port"])
	} else {
		fmt.Printf("❌ 配置信息获取失败，状态码: %d\n", resp.StatusCode)
	}
}

func main() {
	fmt.Println("🚀 开始测试Higress社区治理Agent")
	
	// 等待服务器启动
	fmt.Println("⏳ 等待服务器启动...")
	time.Sleep(2 * time.Second)
	
	// 运行测试
	TestHealth()
	TestConfig()
	TestAgent()
	
	fmt.Println("\n🎉 测试完成！")
} 