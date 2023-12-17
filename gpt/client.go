// ---
// Copyright © 2023 Radoslav Salov <rado.salov@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
// ---

package gpt

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/rammstein4o/git-gpt/git"
	"github.com/rammstein4o/git-gpt/utils"
	"github.com/sashabaranov/go-openai"
)

var (
	tokensMap = map[string]int{
		"gpt-4-32k-0613":         32768,
		"gpt-4-32k-0314":         32768,
		"gpt-4-32k":              32768,
		"gpt-4-0613":             8192,
		"gpt-4-0314":             8192,
		"gpt-4":                  8192,
		"gpt-3.5-turbo-0613":     4096,
		"gpt-3.5-turbo-0301":     4096,
		"gpt-3.5-turbo-16k":      16384,
		"gpt-3.5-turbo-16k-0613": 16384,
		"gpt-3.5-turbo":          4096,
		"gpt-3.5-turbo-instruct": 4096,
	}
)

func countTokens(model string, messages ...openai.ChatCompletionMessage) (int, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, err
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			log.Println("warning: gpt-3.5-turbo may update over time. Returning num tokens assuming gpt-3.5-turbo-0613.")
			return countTokens("gpt-3.5-turbo-0613", messages...)
		} else if strings.Contains(model, "gpt-4") {
			log.Println("warning: gpt-4 may update over time. Returning num tokens assuming gpt-4-0613.")
			return countTokens("gpt-4-0613", messages...)
		}
		return 0, fmt.Errorf("not implemented for model %s", model)
	}

	numTokens := 0
	for _, msg := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(msg.Content, nil, nil))
		numTokens += len(tkm.Encode(msg.Role, nil, nil))
		numTokens += len(tkm.Encode(msg.Name, nil, nil))
		if msg.Name != "" {
			numTokens += tokensPerName
		}
	}
	// every reply is primed with <|start|>assistant<|message|>
	numTokens += 3
	return numTokens, nil
}

func getDeveloperTypeByExtension(fileName string) string {
	var devType string

	extension := strings.ToLower(filepath.Ext(fileName))
	switch extension {
	case ".js", ".jsx", ".mjs", ".cjs", ".mjsx", ".cjsx":
		devType = "JavaScript"
	case ".ts", ".tsx", ".mts", ".cts", ".mtsx", ".ctsx":
		devType = "TypeScript"
	case ".py":
		devType = "Python"
	case ".java", ".jsp":
		devType = "Java"
	case ".scala", ".sc":
		devType = "Scala"
	case ".kt", ".kts":
		devType = "Kotlin"
	case ".groovy", ".gvy", ".gy", ".gsh":
		devType = "Groovy"
	case ".rb":
		devType = "Ruby"
	case ".php", ".phtml":
		devType = "PHP"
	case ".r":
		devType = "R-Lang"
	case ".c":
		devType = "C"
	case ".cs":
		devType = "C#"
	case ".cpp", ".cc", ".cxx", ".h", ".hpp":
		devType = "C++"
	case ".go":
		devType = "Go"
	case ".aspx", ".ascx", ".cshtml":
		devType = "ASP.NET"
	case ".sh", ".bash", ".bat", ".ps1", ".cmd":
		devType = "Shell or batch scripts"
	case ".html", ".htm", ".css", ".less", ".scss", ".sass", ".styl", ".stylus", ".vue", ".ejs":
		devType = "Frontend"
	case ".rs":
		devType = "Rust"
	case ".sql":
		devType = "SQL"
	default:
		devType = ""
	}

	if devType != "" {
		return fmt.Sprintf("expert %s developer", devType)
	}

	return "expert programmer"
}

type Gpt interface {
	SummarizeFile(ctx context.Context, op git.GitOperation, fileName, fileContent string) (string, error)
	SummarizeDiff(ctx context.Context, fileName, diff string) (string, error)
	SummarizeChanges(ctx context.Context, changes []string) (string, error)
	FinalizeCommitMsg(ctx context.Context, prompt string) (string, error)
	GetStats(ctx context.Context) *Stats
}

// Ensure, that gitcmd does implement Git.
var _ Gpt = &client{}

type client struct {
	model        string
	stream       bool
	maxTokens    int
	temperature  float32
	topP         float32
	maxChunkSize int
	client       *openai.Client
	stats        *Stats
}

func (c *client) createChatCompletion(ctx context.Context, content string, systemMessages ...string) (openai.ChatCompletionResponse, error) {
	messages := make([]openai.ChatCompletionMessage, 0)

	for _, msg := range systemMessages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: strings.TrimSpace(msg),
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: strings.TrimSpace(content),
	})

	tokenLimit := tokensMap[c.model]
	numTokens, _ := countTokens(c.model, messages...)
	if numTokens > tokenLimit-c.maxTokens {
		return openai.ChatCompletionResponse{}, fmt.Errorf("too many tokens used %d (%d)", numTokens, tokenLimit)
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      c.stream,
		N:           1,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		TopP:        c.topP,
	}

	return c.client.CreateChatCompletion(ctx, req)
}

func (c *client) SummarizeFile(ctx context.Context, op git.GitOperation, fileName, fileContent string) (string, error) {
	result := make([]string, 0)

	var str string
	switch op {
	case git.OPERATION_ADD:
		str = "Added"
	case git.OPERATION_DEL:
		str = "Removed"
	default:
		str = "Modified"
	}

	result = append(result, fmt.Sprintf("%s file `%s`: ", str, fileName))

	prevChunkSummary := ""
	chunks := utils.SplitText(fileContent, c.maxChunkSize)
	for _, chunk := range chunks {
		systemMsgs := make([]string, 0)
		tmpMsg, err := utils.GetTemplateByString(
			SummarizeFileTemplate,
			utils.Data{
				"operation":        op,
				"devType":          getDeveloperTypeByExtension(fileName),
				"file":             filepath.Base(fileName),
				"prevChunkSummary": prevChunkSummary,
			},
		)
		if err != nil {
			return "", err
		}

		systemMsgs = append(systemMsgs, tmpMsg)

		if prevChunkSummary != "" {
			tmpMsg, err := utils.GetTemplateByString(
				PrevChunkSummaryTemplate,
				utils.Data{
					"prevChunkSummary": prevChunkSummary,
				},
			)
			if err != nil {
				return "", err
			}

			systemMsgs = append(systemMsgs, tmpMsg)
		}

		resp, err := c.createChatCompletion(ctx, chunk, systemMsgs...)
		if err != nil {
			return "", err
		}
		c.stats.NumRequests += 1
		c.stats.PromptTokens += resp.Usage.PromptTokens
		c.stats.CompletionTokens += resp.Usage.CompletionTokens
		c.stats.TotalTokens += resp.Usage.TotalTokens
		completion := strings.TrimSpace(resp.Choices[0].Message.Content)
		prevChunkSummary = completion
		result = append(result, completion)
	}

	c.stats.NumFiles += 1
	return strings.TrimSpace(strings.Join(result, " ")), nil
}

func (c *client) SummarizeDiff(ctx context.Context, fileName, diff string) (string, error) {
	result := make([]string, 0)
	result = append(result, fmt.Sprintf("Modified file `%s`: ", fileName))

	prevChunkSummary := ""
	chunks := utils.SplitText(diff, c.maxChunkSize)
	for _, chunk := range chunks {
		systemMsgs := make([]string, 0)
		tmpMsg, err := utils.GetTemplateByString(
			SummarizeDiffTemplate,
			utils.Data{
				"devType":          getDeveloperTypeByExtension(fileName),
				"file":             filepath.Base(fileName),
				"prevChunkSummary": prevChunkSummary,
			},
		)
		if err != nil {
			return "", err
		}

		systemMsgs = append(systemMsgs, tmpMsg)

		if prevChunkSummary != "" {
			tmpMsg, err := utils.GetTemplateByString(
				PrevChunkSummaryTemplate,
				utils.Data{
					"prevChunkSummary": prevChunkSummary,
				},
			)
			if err != nil {
				return "", err
			}

			systemMsgs = append(systemMsgs, tmpMsg)
		}

		resp, err := c.createChatCompletion(ctx, chunk, systemMsgs...)
		if err != nil {
			return "", err
		}
		c.stats.NumRequests += 1
		c.stats.PromptTokens += resp.Usage.PromptTokens
		c.stats.CompletionTokens += resp.Usage.CompletionTokens
		c.stats.TotalTokens += resp.Usage.TotalTokens
		completion := strings.TrimSpace(resp.Choices[0].Message.Content)
		prevChunkSummary = completion
		result = append(result, completion)
	}

	c.stats.NumFiles += 1
	return strings.TrimSpace(strings.Join(result, " ")), nil
}

func (c *client) SummarizeChanges(ctx context.Context, changes []string) (string, error) {
	systemMsg, err := utils.GetTemplateByString(
		SummarizeChangesTemplate,
		utils.Data{},
	)
	if err != nil {
		return "", err
	}

	prompt := ""
	result := make([]string, 0)

	for _, summary := range changes {
		if len(prompt)+len(summary) > c.maxChunkSize {
			resp, err := c.createChatCompletion(ctx, prompt, systemMsg)
			if err != nil {
				return "", err
			}
			c.stats.NumRequests += 1
			c.stats.PromptTokens += resp.Usage.PromptTokens
			c.stats.CompletionTokens += resp.Usage.CompletionTokens
			c.stats.TotalTokens += resp.Usage.TotalTokens
			result = append(result, strings.TrimSpace(resp.Choices[0].Message.Content))

			prompt = ""
		}

		prompt = fmt.Sprintf("%s\n%s", prompt, summary)
	}

	if strings.TrimSpace(prompt) != "" {
		resp, err := c.createChatCompletion(ctx, prompt, systemMsg)
		if err != nil {
			return "", err
		}
		c.stats.NumRequests += 1
		c.stats.PromptTokens += resp.Usage.PromptTokens
		c.stats.CompletionTokens += resp.Usage.CompletionTokens
		c.stats.TotalTokens += resp.Usage.TotalTokens
		result = append(result, strings.TrimSpace(resp.Choices[0].Message.Content))
	}

	return strings.TrimSpace(strings.Join(result, "\n")), nil
}

func (c *client) FinalizeCommitMsg(ctx context.Context, prompt string) (string, error) {
	systemMsg, err := utils.GetTemplateByString(
		SummarizeChangesTemplate,
		utils.Data{},
	)
	if err != nil {
		return "", err
	}

	resp, err := c.createChatCompletion(ctx, prompt, systemMsg)
	if err != nil {
		return "", err
	}
	c.stats.NumRequests += 1
	c.stats.PromptTokens += resp.Usage.PromptTokens
	c.stats.CompletionTokens += resp.Usage.CompletionTokens
	c.stats.TotalTokens += resp.Usage.TotalTokens

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (c *client) GetStats(ctx context.Context) *Stats {
	return c.stats
}

func New(opts ...Option) Gpt {
	cl := &client{
		stats: &Stats{},
	}

	// Loop through each option passed as argument and apply it to the config object
	for _, fn := range opts {
		fn(cl)
	}

	return cl
}
