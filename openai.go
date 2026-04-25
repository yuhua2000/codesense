package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AIModel interface {
	ChatStream(ctx context.Context, history []map[string]string) <-chan string
	Chat(ctx context.Context, history []map[string]string) (string, error)
}

type OpenAI struct {
	Key     string
	BaseUrl string
	Model   string
}

func NewOpenAI(key string, model string, url string) AIModel {
	if url == "" {
		url = "https://api.openai.com/"
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return &OpenAI{
		Key:     key,
		BaseUrl: url,
		Model:   model,
	}
}

func (o *OpenAI) Vaild(ctx context.Context) error {
	client := openai.NewClient(option.WithBaseURL(o.BaseUrl), option.WithAPIKey(o.Key))
	res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("only return '1'"),
		},
		Model: o.Model,
		Seed:  openai.Int(24),
	})
	if err != nil {
		return err
	}
	if len(res.Choices) == 0 {
		return errors.New("no response")
	}
	if len(res.Choices[0].Message.Content) == 0 {
		return errors.New("invalid response")
	}
	return nil
}

func (o *OpenAI) Chat(ctx context.Context, history []map[string]string) (string, error) {
	var modelResponse string

	slog.Debug("Calling AI model once")

	responseChannel := o.ChatStream(ctx, history)
	for response := range responseChannel {
		modelResponse += response
	}

	if modelResponse == "" {
		slog.Warn("Received empty response from AI model")
		return "", fmt.Errorf("empty response from AI model")
	}

	slog.Info("Received model response successfully")
	slog.Debug("Received model response", "response", modelResponse)

	return modelResponse, nil
}

func (o *OpenAI) ChatStream(ctx context.Context, history []map[string]string) <-chan string {
	client := openai.NewClient(option.WithBaseURL(o.BaseUrl), option.WithAPIKey(o.Key))
	resp := make(chan string)
	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0)
	totalLength := 0
	for _, item := range history {
		role := item["role"]
		content := item["content"]
		switch role {
		case "assistant":
			totalLength += len(content)
			chatMessages = append(chatMessages, openai.AssistantMessage(content))
		case "user":
			totalLength += len(content)
			chatMessages = append(chatMessages, openai.UserMessage(content))
		}
	}

	slog.Info("GPT request", "length", totalLength)
	slog.Debug("GPT request", "data", history, "length", totalLength)
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: chatMessages,
		Seed:     openai.Int(24),
		Model:    o.Model,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	})

	go func() {
		var usage openai.CompletionUsage
		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				word := evt.Choices[0].Delta.Content
				resp <- word
			}
			if evt.Usage.TotalTokens != 0 {
				usage = evt.Usage
			}
		}

		slog.Info("GPT token usage", "model", o.Model, "prompt_tokens", usage.PromptTokens, "completion_tokens", usage.CompletionTokens,
			"total_tokens", usage.TotalTokens)

		if stream.Err() != nil {
			slog.Error("ChatStream error", "错误", stream.Err())
		}
		close(resp)
	}()
	return resp
}
