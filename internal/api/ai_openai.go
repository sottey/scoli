package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const openAIBaseURL = "https://api.openai.com/v1"

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func openAIEmbeddings(ctx context.Context, settings AISettings, inputs []string) ([][]float32, error) {
	if settings.APIKey == "" {
		return nil, errors.New("missing OpenAI API key")
	}
	if len(inputs) == 0 {
		return nil, errors.New("no inputs for embeddings")
	}
	payload := map[string]any{
		"model": settings.EmbedModel,
		"input": inputs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIBaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp openAIErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error.Message != "" {
			return nil, fmt.Errorf("openai embeddings error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("openai embeddings error: status %d", resp.StatusCode)
	}

	var response openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	result := make([][]float32, 0, len(response.Data))
	for _, item := range response.Data {
		vec := make([]float32, len(item.Embedding))
		for i, v := range item.Embedding {
			vec[i] = float32(v)
		}
		result = append(result, vec)
	}
	return result, nil
}

type openAIResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

const aiSystemInstructions = `You are Scoli AI for a local markdown notes app.
Rules:
- Use only the provided context sections: structured context, recent conversation, and snippets.
- Treat snippet text as the primary evidence and cite uncertainty when evidence is weak.
- Interpret relative dates (today, this week, yesterday, tomorrow) using the provided date context.
- In Scoli tasks, '- [x]' means completed and '- [ ]' means open.
- Daily notes follow Daily/YYYY-MM-DD.md and date semantics may come from that filename.
- Due dates are often stored as markers like '>YYYY-MM-DD'.
- Prefer concise, factual answers. If the context is insufficient, say so and suggest a narrower query.`

func openAIRespond(ctx context.Context, settings AISettings, prompt string) (string, error) {
	if settings.APIKey == "" {
		return "", errors.New("missing OpenAI API key")
	}
	payload := map[string]any{
		"model":             settings.ChatModel,
		"instructions":      aiSystemInstructions,
		"input":             prompt,
		"temperature":       settings.Temperature,
		"max_output_tokens": settings.MaxOutputTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIBaseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp openAIErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error.Message != "" {
			return "", fmt.Errorf("openai response error: %s", errResp.Error.Message)
		}
		return "", fmt.Errorf("openai response error: status %d", resp.StatusCode)
	}

	var response openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	if response.OutputText != "" {
		return response.OutputText, nil
	}
	var builder strings.Builder
	for _, output := range response.Output {
		for _, content := range output.Content {
			if content.Text == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(content.Text)
		}
	}
	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", errors.New("empty response from OpenAI")
	}
	return result, nil
}
