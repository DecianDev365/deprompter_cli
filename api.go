package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const systemInstruction = "You are an expert AI image prompt reverse engineer with deep knowledge of Midjourney, DALL-E, Stable Diffusion, and Flux. Analyze this image and reconstruct the most accurate and detailed prompt that could have generated it. Return only the prompt text, nothing else. Make it detailed enough that someone could paste it directly into Midjourney, DALL-E, or Stable Diffusion and get a very similar result."

func GeneratePrompt(provider, apiKey, base64Image, mimeType string) (string, error) {
	switch strings.ToLower(provider) {
	case "groq":
		return groqPrompt(apiKey, base64Image, mimeType)
	case "gemini":
		return geminiPrompt(apiKey, base64Image, mimeType)
	case "openrouter":
		return generatePromptFromOpenRouter(apiKey, base64Image, mimeType)
	default:
		return "", fmt.Errorf("unsupported provider: %s (use groq or gemini)", provider)
	}
}

type groqMessage struct {
	Role    string        `json:"role"`
	Content []groqContent `json:"content"`
}

type groqContent struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	ImageURL *groqImageURL  `json:"image_url,omitempty"`
}

type groqImageURL struct {
	URL string `json:"url"`
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func groqPrompt(apiKey, base64Image, mimeType string) (string, error) {
	body := groqRequest{
		Model: "meta-llama/llama-4-scout-17b-16e-instruct",
		Messages: []groqMessage{
			{
				Role: "user",
				Content: []groqContent{
					{
						Type: "text",
						Text: systemInstruction,
					},
					{
						Type:     "image_url",
						ImageURL: &groqImageURL{URL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)},
					},
				},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq API request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("groq API returned status %d: %s", resp.StatusCode, string(raw))
	}

	var gr groqResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("failed to parse groq response: %w", err)
	}
	if len(gr.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return strings.TrimSpace(gr.Choices[0].Message.Content), nil
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string         `json:"text,omitempty"`
	InlineData *geminiData    `json:"inline_data,omitempty"`
}

type geminiData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func geminiPrompt(apiKey, base64Image, mimeType string) (string, error) {
	body := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: systemInstruction},
					{
						InlineData: &geminiData{
							MimeType: mimeType,
							Data:     base64Image,
						},
					},
				},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(raw))
	}

	var gr geminiResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("failed to parse gemini response: %w", err)
	}
	if len(gr.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}
	if len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content parts")
	}
	return strings.TrimSpace(gr.Candidates[0].Content.Parts[0].Text), nil
}

type openrouterMessage struct {
	Role    string              `json:"role"`
	Content []openrouterContent `json:"content"`
}

type openrouterContent struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	ImageURL *openrouterImageURL `json:"image_url,omitempty"`
}

type openrouterImageURL struct {
	URL string `json:"url"`
}

type openrouterRequest struct {
	Model    string              `json:"model"`
	Messages []openrouterMessage `json:"messages"`
}

type openrouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func generatePromptFromOpenRouter(apiKey, imageBase64, mimeType string) (string, error) {
	body := openrouterRequest{
		Model: "meta-llama/llama-4-scout",
		Messages: []openrouterMessage{
			{
				Role: "user",
				Content: []openrouterContent{
					{
						Type: "text",
						Text: systemInstruction,
					},
					{
						Type:     "image_url",
						ImageURL: &openrouterImageURL{URL: fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64)},
					},
				},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter API request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("openrouter API returned status %d: %s", resp.StatusCode, string(raw))
	}

	var or openrouterResponse
	if err := json.Unmarshal(raw, &or); err != nil {
		return "", fmt.Errorf("failed to parse openrouter response: %w", err)
	}
	if len(or.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned no choices")
	}
	return strings.TrimSpace(or.Choices[0].Message.Content), nil
}
