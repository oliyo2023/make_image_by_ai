//go:build testclient
// +build testclient

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// 测试请求结构
type TestRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

// 测试响应结构
type TestResponse struct {
	TranslatedPrompt string `json:"translated_prompt"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
}

func main() {
	baseURL := "http://127.0.0.1:8000"

	// 测试健康检查
	fmt.Println("=== 测试健康检查 ===")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("健康检查响应: %s\n\n", string(body))

	// 测试翻译功能
	fmt.Println("=== 测试翻译功能 ===")
	translateReq := TestRequest{
		Prompt: "一只可爱的小猫",
	}

	jsonData, _ := json.Marshal(translateReq)
	resp, err = http.Post(baseURL+"/translate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("翻译请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("翻译响应: %s\n\n", string(body))

	// 测试图像生成功能
	fmt.Println("=== 测试图像生成功能 ===")
	generateReq := TestRequest{
		Prompt: "一只可爱的小猫",
		Model:  "google/gemini-2.5-flash-image-preview:free",
	}

	jsonData, _ = json.Marshal(generateReq)
	resp, err = http.Post(baseURL+"/generate-image", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("图像生成请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("图像生成响应: %s\n\n", string(body))

	// 测试图片列表功能
	fmt.Println("=== 测试图片列表功能 ===")
	resp, err = http.Get(baseURL + "/images")
	if err != nil {
		fmt.Printf("图片列表请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("图片列表响应: %s\n", string(body))
}
