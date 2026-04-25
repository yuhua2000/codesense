package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 初始化 slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 参数
	codePath := flag.String("path", ".", "代码仓库路径")
	modelName := flag.String("model", "", "AI模型名称")
	apiKey := flag.String("key", "", "OpenAI API密钥")
	baseURL := flag.String("url", "", "API基础URL")
	language := flag.String("lang", "zh", "输出语言 (zh/en)")
	outputFile := flag.String("output", "info.md", "输出报告文件路径 (可选)")

	flag.Parse()

	// API Key 校验
	if *apiKey == "" {
		*apiKey = os.Getenv("OPENAI_API_KEY")
		if *apiKey == "" {
			slog.Error("missing API key", "flag", "-key", "env", "OPENAI_API_KEY")
			os.Exit(1)
		}
	}

	// 路径处理
	absPath, err := filepath.Abs(*codePath)
	if err != nil {
		slog.Error("resolve path failed", "path", *codePath, "err", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		slog.Error("path not exist", "path", absPath)
		os.Exit(1)
	}

	slog.Info("start analysis", "path", absPath, "model", *modelName, "lang", *language)

	// 初始化 AI
	aiModel := NewOpenAI(*apiKey, *modelName, *baseURL)

	task := &CodeAnalysisTask{
		CodePath: absPath,
		AIModel:  aiModel,
		Language: *language,
	}

	// 执行分析
	ctx := context.Background()
	report, err := task.RunAnalysis(ctx)
	if err != nil {
		slog.Error("analysis failed", "err", err)
		os.Exit(1)
	}

	// 输出结果
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, []byte(report), 0644)
		if err != nil {
			slog.Error("write file failed", "file", *outputFile, "err", err)
			os.Exit(1)
		}

		slog.Info("report saved", "file", *outputFile)
		return
	}

	// 打印到终端
	slog.Info("print report to stdout")

	slog.Info("代码仓库分析报告开始")

	slog.Info(strings.Repeat("=", 80))
	slog.Info("代码仓库分析报告")
	slog.Info(strings.Repeat("=", 80))

	slog.Info("report", "content", report)

	slog.Info("代码仓库分析报告结束")
}

type CodeAnalysisTask struct {
	CodePath string
	AIModel  AIModel
	Language string // zh / en
}

// ---------------------- 代码仓库分析 ---------------------------

type AnalysisTemplate struct {
	CodePath              string
	DirectoryStructure    string
	StaticAnalysisResults string
	OriginalReports       string
}

func (t *CodeAnalysisTask) RunAnalysis(ctx context.Context) (string, error) {
	// 获取目录结构
	dirPrompt, err := ListDir(t.CodePath, 2, "")
	if err != nil {
		slog.Error("读取目录失败", "目录", t.CodePath, "错误", err)
		return "", err
	}

	// 解析模板
	tpl, err := template.New("analysisTemplate").Parse(AnalysisPrompt)
	if err != nil {
		slog.Error("创建模板失败", "错误", err)
		return "", err
	}

	// 执行模板
	var buf bytes.Buffer
	err = tpl.Execute(&buf, AnalysisTemplate{
		CodePath:           t.CodePath,
		DirectoryStructure: dirPrompt,
	})
	if err != nil {
		slog.Debug("执行模板失败", "错误", err)
		return "", err
	}

	// 创建 AnalysisAgent 代理进行分析
	agent := NewAgent(t.AIModel, t.Language, t.CodePath)
	err = agent.RunLoop(ctx, buf.String())
	if err != nil {
		slog.Error("AI分析失败", "错误", err)
		return "", err
	}

	// 生成总结报告
	report, err := t.SummaryChat(ctx, agent)
	return report, err
}

func (t *CodeAnalysisTask) SummaryChat(ctx context.Context, agent *CodeAnalysisAgent) (string, error) {
	history := agent.GetHistory()
	history = append(history, map[string]string{
		"role":    "user",
		"content": fmt.Sprintf(SummaryPrompt, LanguagePrompt(t.Language)),
	})

	slog.Info("生成总结报告")

	result, err := t.AIModel.Chat(ctx, history)
	if err != nil {
		return "", fmt.Errorf("获取AI响应失败: %v", err)
	}

	return result, nil
}

func LanguagePrompt(language string) string {
	if language == "zh" {
		return "请使用中文回复。"
	}
	return "Please respond in English."
}
