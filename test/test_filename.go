package main

import (
	"fmt"
	"make_image_by_ai/utils"
)

func main() {
	fmt.Println("=== 文件名生成测试 ===")

	// 测试中文提示词 vs 英文提示词
	testCases := []struct {
		original string
		english  string
		expected string
	}{
		{
			original: "画一条中华龙,要写实",
			english:  "Draw a Chinese dragon in a realistic style",
			expected: "draw_chinese_dragon_realistic",
		},
		{
			original: "一只可爱的小猫",
			english:  "A cute little kitten",
			expected: "cute_little_kitten",
		},
		{
			original: "美丽的风景画",
			english:  "Beautiful landscape painting",
			expected: "beautiful_landscape_painting",
		},
		{
			original: "宫崎骏风格的动画角色",
			english:  "Miyazaki style animated character",
			expected: "miyazaki_style_animated_character",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n--- 测试用例 %d ---\n", i+1)
		fmt.Printf("原始提示词: %s\n", tc.original)
		fmt.Printf("英文提示词: %s\n", tc.english)

		// 测试旧的中文关键词提取
		chineseKeywords := extractKeywords(tc.original)
		fmt.Printf("中文关键词: %s\n", chineseKeywords)

		// 测试新的英文关键词提取
		englishKeywords := utils.ExtractEnglishKeywords(tc.english)
		fmt.Printf("英文关键词: %s\n", englishKeywords)

		fmt.Printf("预期结果: %s\n", tc.expected)
		fmt.Printf("匹配度: %t\n", englishKeywords == tc.expected)
	}

	fmt.Println("\n=== 测试完成 ===")
}

// 复制utils包中的函数用于测试
func extractKeywords(prompt string) string {
	// 这里简化实现，实际使用utils包中的函数
	return "chinese_extracted_keywords"
}
