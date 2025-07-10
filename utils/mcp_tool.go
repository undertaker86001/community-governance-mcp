package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/higress-group/wasm-go/pkg/mcp/server"
) // SendHTTPRequest 发送 HTTP 请求的辅助函数

func SendHTTPRequest(ctx server.HttpContext, method, url string, headers map[string]string, body string) (string, error) {
	// 转换 headers 格式为 [][2]string
	headerPairs := make([][2]string, 0, len(headers))
	for key, value := range headers {
		headerPairs = append(headerPairs, [2]string{key, value})
	}

	// 使用 channel 来同步异步调用
	resultChan := make(chan httpResult, 1)

	// 使用 RouteCall 发送请求
	err := ctx.RouteCall(method, url, headerPairs, []byte(body),
		func(statusCode int, responseHeaders [][2]string, responseBody []byte) {
			result := httpResult{
				statusCode: statusCode,
				body:       string(responseBody),
			}

			if statusCode < 200 || statusCode >= 300 {
				result.err = fmt.Errorf("HTTP request failed with status %d", statusCode)
			}

			// 发送结果到 channel
			select {
			case resultChan <- result:
			default:
				// channel 已满，避免阻塞
			}
		})

	if err != nil {
		return "", fmt.Errorf("failed to initiate HTTP request: %v", err)
	}

	// 等待响应，设置超时
	select {
	case result := <-resultChan:
		if result.err != nil {
			return "", result.err
		}
		return result.body, nil
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("HTTP request timeout")
	}
}

// httpResult 用于传递 HTTP 响应结果
type httpResult struct {
	statusCode int
	body       string
	err        error
}

// SendHTTPRequestWithTimeout 带超时的 HTTP 请求函数
func SendHTTPRequestWithTimeout(ctx server.HttpContext, method, url string, headers map[string]string, body string, timeoutMs uint32) (string, error) {
	headerPairs := make([][2]string, 0, len(headers))
	for key, value := range headers {
		headerPairs = append(headerPairs, [2]string{key, value})
	}

	resultChan := make(chan httpResult, 1)

	err := ctx.RouteCall(method, url, headerPairs, []byte(body),
		func(statusCode int, responseHeaders [][2]string, responseBody []byte) {
			result := httpResult{
				statusCode: statusCode,
				body:       string(responseBody),
			}

			if statusCode < 200 || statusCode >= 300 {
				result.err = fmt.Errorf("HTTP request failed with status %d", statusCode)
			}

			select {
			case resultChan <- result:
			default:
			}
		})

	if err != nil {
		return "", fmt.Errorf("failed to initiate HTTP request: %v", err)
	}

	timeout := time.Duration(timeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second // 默认超时
	}

	select {
	case result := <-resultChan:
		if result.err != nil {
			return "", result.err
		}
		return result.body, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("HTTP request timeout after %v", timeout)
	}
}

// SendHTTPRequestAsync 异步 HTTP 请求函数，直接使用回调
func SendHTTPRequestAsync(ctx server.HttpContext, method, url string, headers map[string]string, body string, callback func(string, error)) error {
	headerPairs := make([][2]string, 0, len(headers))
	for key, value := range headers {
		headerPairs = append(headerPairs, [2]string{key, value})
	}

	return ctx.RouteCall(method, url, headerPairs, []byte(body),
		func(statusCode int, responseHeaders [][2]string, responseBody []byte) {
			if statusCode < 200 || statusCode >= 300 {
				callback("", fmt.Errorf("HTTP request failed with status %d", statusCode))
				return
			}
			callback(string(responseBody), nil)
		})
}

// SendMCPToolTextResult 发送MCP工具文本结果
// 用于向MCP客户端返回工具执行结果
func SendMCPToolTextResult(ctx server.HttpContext, toolName string, result string, success bool) error {
	// 构建结果消息
	response := map[string]interface{}{
		"tool_name": toolName,
		"result":    result,
		"success":   success,
		"timestamp": time.Now().Unix(),
	}

	// 转换为JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal MCP tool result: %v", err)
	}

	// 发送HTTP响应
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, err = SendHTTPRequest(ctx, "POST", "/mcp/tool/result", headers, string(responseJSON))
	if err != nil {
		return fmt.Errorf("failed to send MCP tool result: %v", err)
	}

	return nil
}

// SendMCPToolErrorResult 发送MCP工具错误结果
// 用于向MCP客户端返回工具执行错误
func SendMCPToolErrorResult(ctx server.HttpContext, toolName string, errorMsg string) error {
	return SendMCPToolTextResult(ctx, toolName, errorMsg, false)
}

// SendMCPToolSuccessResult 发送MCP工具成功结果
// 用于向MCP客户端返回工具执行成功结果
func SendMCPToolSuccessResult(ctx server.HttpContext, toolName string, result string) error {
	return SendMCPToolTextResult(ctx, toolName, result, true)
}
