package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Command struct {
	Name string
	Arg1 string
}

// CodeAnalysisAgent is an agent that can automatically execute tasks
type CodeAnalysisAgent struct {
	MaxIter          int // Maximum number of iterations
	Compress         int
	MaxFileReadBytes int

	folder   string
	aiModel  AIModel
	history  []map[string]string
	language string
}

func NewAgent(aiModel AIModel, language string, folder string) *CodeAnalysisAgent {
	return &CodeAnalysisAgent{
		MaxIter:          60,
		MaxFileReadBytes: 10 * 1024, // Maximum read size
		Compress:         9,

		language: language,
		folder:   folder,
		aiModel:  aiModel,
	}
}

// BuildNextPrompt generates next prompt
func (a *CodeAnalysisAgent) BuildNextPrompt(retMsg string, round int) string {
	return fmt.Sprintf("The current round is the %dth conversation. Please try to minimize the number of exchanges to obtain the result.\n.The returned result is as follows. Please draw your conclusion in the \"reply format\".Determine which next command to use, and respond using the format specified above.\nReturn:%s", round, retMsg)
}

// ExtractTag extracts tag part from text
type ReadFileParam struct {
	FilePath  string
	StartLine int
	EndLine   int
}

type ListDirParam struct {
	FilePath string
	Depth    int
	Exts     string
}

type GrepParam struct {
	FilePath string
	Pattern  string
	Context  int
}

func (a *CodeAnalysisAgent) GetHistory() []map[string]string {
	return a.history
}

// RunLoop runs agent
func (a *CodeAnalysisAgent) RunLoop(ctx context.Context, prompt string) error {
	history := []map[string]string{
		{
			"role":    "user",
			"content": prompt,
		},
	}

	// Start main loop
	index := 0
	for {
		if a.MaxIter > 0 && index >= a.MaxIter {
			userPrompt := fmt.Sprintf("Maximum iteration count reached: %d, please complete directly based on history, output command as finish, generate final result", a.MaxIter)
			// Remove last history entry
			if len(history) > 1 {
				history = history[:len(history)-1]
			}
			history = append(history, map[string]string{
				"role":    "user",
				"content": userPrompt,
			})
		}

		slog.Info("Round", "index", index+1)
		index++

		if a.Compress > 0 && len(history) > a.Compress {
			compressed, err := a.compressHistory(ctx, history[1:])
			if err != nil {
				return fmt.Errorf("failed to compress history: %v", err)
			}

			history = append(history[:1],
				map[string]string{
					"role":    "assistant",
					"content": compressed,
				})
		}

		// Call LLM API to generate response
		m := history
		msg, err := a.aiModel.Chat(ctx, m)
		if err != nil {
			return fmt.Errorf("failed to get AI response: %v", err)
		}

		// Add response to history
		history = append(history, map[string]string{
			"role":    "assistant",
			"content": msg,
		})
		a.history = history

		// Try to parse JSON command
		jsonStr := ExtractTag(msg, "command")
		if jsonStr == "" {
			slog.Warn("Command parsing failed, retrying")
			// JSON parsing failed
			history = append(history, map[string]string{
				"role":    "user",
				"content": "Your command output format is incorrect, please answer based on the previous question and reorganize your command output format as <command><name>command name</name><arg><parameters1>parameter content</parameters1><parameters2>parameter content</parameters2></arg></command> and re-answer",
			})
			continue
		}

		var command Command
		command.Name = ExtractTag(jsonStr, "name")
		command.Arg1 = ExtractTag(jsonStr, "arg")

		// Check if command format is correct
		if command.Name == "" || (command.Name != "finish" && command.Arg1 == "") {
			history = append(history, map[string]string{
				"role":    "user",
				"content": "Your command output format is incorrect, please answer based on the previous question and reorganize your command output format as <command><name>command name</name><arg><parameters1>parameter content</parameters1><parameters2>parameter content</parameters2></arg></command> and re-answer",
			})
			continue
		}

		slog.Info("执行命令", "命令", command.Name, "参数", command.Arg1)

		userPrompt, err, finished := a.ExecuteCommand(command)
		if finished {
			return err
		}

		maxLength := 200
		userPrompt2 := []rune(userPrompt)
		if len(userPrompt2) < maxLength {
			maxLength = len(userPrompt2)
		}
		slog.Info("执行命令成功", "结果", string(userPrompt2[:maxLength]))

		// Add user prompt to history
		history = append(history, map[string]string{
			"role":    "user",
			"content": a.BuildNextPrompt(userPrompt, index+1),
		})
		a.history = history
	}
}

func (a *CodeAnalysisAgent) ExecuteCommand(command Command) (string, error, bool) {
	var userPrompt string
	switch command.Name {
	case "list_dir":
		var parameter ListDirParam
		err := ParseCommandParam(command.Arg1, &parameter)
		if err != nil {
			userPrompt = fmt.Sprintf("Failed to parse command parameters: %v", err)
			break
		}
		err = a.inFolder(parameter.FilePath)
		if err != nil {
			userPrompt = err.Error()
			break
		}
		data, err := ListDir(parameter.FilePath, parameter.Depth, parameter.Exts)
		if err != nil {
			userPrompt = fmt.Sprintf("Failed to read directory: %v Please confirm if the directory path is correct", err)
		} else {
			userPrompt = fmt.Sprintf("Directory reading completed, dir path:%s\nDir tree:\n%s\n", parameter.FilePath, data)
		}
	case "read_file":
		var parameter ReadFileParam
		err := ParseCommandParam(command.Arg1, &parameter)
		if err != nil {
			userPrompt = fmt.Sprintf("Failed to parse command parameters: %v", err)
			break
		}
		err = a.inFolder(parameter.FilePath)
		if err != nil {
			userPrompt = err.Error()
			break
		}
		// Get file information
		fileInfo, err := os.Stat(parameter.FilePath)
		if err != nil {
			slog.Error("读取文件失败 ", "文件", parameter.FilePath, "错误", err)
			userPrompt = fmt.Sprintf("Failed to read file %s: %v", parameter.FilePath, err)
			break
		}
		// Check file size, decide how to read
		if fileInfo.Size() < int64(a.MaxFileReadBytes) {
			// Small file, read all at once
			data, err := os.ReadFile(parameter.FilePath)
			if err != nil {
				userPrompt = fmt.Sprintf("Failed to read file: %v", err)
			} else {
				userPrompt = fmt.Sprintf("File reading completed, filename:%s\nfile content:\n%s", parameter.FilePath, string(data))
			}
		} else {
			startline := parameter.StartLine
			endline := parameter.EndLine
			if startline == 0 && endline == 0 {
				startline = 0
				endline = 200
			}
			if endline < startline {
				userPrompt = fmt.Sprintf("endline(%d) cannot be less than startline(%d)", endline, startline)
				break
			}
			// Large file, read in chunks
			content, err := ReadFileChunk(parameter.FilePath, startline, endline, a.MaxFileReadBytes)
			if err != nil {
				userPrompt = fmt.Sprintf("Failed to read file: %v", err)
			} else {
				userPrompt = content
			}
		}
	case "grep":
		var parameter GrepParam
		err := ParseCommandParam(command.Arg1, &parameter)
		if err != nil {
			userPrompt = fmt.Sprintf("Failed to parse command parameters: %v", err)
			break
		}
		err = a.inFolder(parameter.FilePath)
		if err != nil {
			userPrompt = err.Error()
			break
		}
		// Execute grep
		results, err := Grep(parameter.FilePath, parameter.Pattern, parameter.Context)
		if err != nil {
			userPrompt = fmt.Sprintf("Search failed: %v", err)
		} else {
			pattern := parameter.Pattern
			patternDesc := strings.ReplaceAll(pattern, ",", "', '")
			if strings.Contains(pattern, ",") {
				patternDesc = fmt.Sprintf("['%s']", patternDesc)
			} else {
				patternDesc = fmt.Sprintf("'%s'", pattern)
			}
			userPrompt = fmt.Sprintf("Search results (path: %s, pattern: %s, context lines: %d):\n%s",
				parameter.FilePath, patternDesc, parameter.Context, results)
		}
	case "finish":
		return "", nil, true
	default:
		userPrompt = fmt.Sprintf("Unknown command: %s You can only use read_file, list_dir, grep, finish commands", command.Name)
	}
	return userPrompt, nil, false
}

func (a *CodeAnalysisAgent) inFolder(arg1 string) error {
	folder, err := filepath.Abs(arg1)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}
	if !strings.HasPrefix(folder, a.folder) {
		return fmt.Errorf("security policy restriction, you cannot read directories outside %s, %s is not in the current directory", a.folder, folder)
	}
	return nil
}

func (a *CodeAnalysisAgent) compressHistory(ctx context.Context, history []map[string]string) (string, error) {
	slog.Info("Compressing history", "len", len(history))
	newHistory := append([]map[string]string(nil), history...)
	newHistory = append(newHistory, map[string]string{
		"role":    "user",
		"content": CompressionPrompt,
	})

	msg, err := a.aiModel.Chat(ctx, newHistory)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %v", err)
	}

	slog.Info("Successfully compressed history", "Result", msg)
	return msg, nil
}
