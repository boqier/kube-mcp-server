package sendmessage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendToFeishu(analysis string, feishuWebhookURL string) (string, error) {
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": analysis,
		},
	}

	// 将消息内容序列化为 JSON
	messageBody, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", feishuWebhookURL, bytes.NewBuffer(messageBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发出请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to send Feishu alert, status code: %d", resp.StatusCode)
	}

	return "send success", nil
}
