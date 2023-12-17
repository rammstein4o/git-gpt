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

import "github.com/sashabaranov/go-openai"

type Option func(*client)

func WithOpenAI(token, model string) Option {
	return func(c *client) {
		c.model = model
		c.client = openai.NewClient(token)
	}
}

func WithAzureOpenAI(token, endpoint, model, alias string) Option {
	config := openai.DefaultAzureConfig(token, endpoint)

	modelName := model
	if alias != "" {
		modelName = alias
	}

	// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	config.AzureModelMapperFunc = func(_ string) string {
		return modelName
	}

	return func(c *client) {
		c.model = model
		c.client = openai.NewClientWithConfig(config)
	}
}

func WithStream(stream bool) Option {
	return func(c *client) {
		c.stream = stream
	}
}

func WithMaxTokens(maxTokens int) Option {
	return func(c *client) {
		c.maxTokens = maxTokens
	}
}

// A low temperature makes the model more confident in its top choices, while temperatures greater than 1 decrease confidence in its top choices.
// An even higher temperature corresponds to more uniform sampling (total randomness).
// A temperature of 0 is equivalent to argmax/max likelihood, or the highest probability token.
func WithTemperature(temperature float32) Option {
	return func(c *client) {
		c.temperature = temperature
	}
}

// TopP computes the cumulative probability distribution, and cut off as soon as that distribution exceeds the value of TopP.
// For example, a TopP of 0.3 means that only the tokens comprising the top 30% probability mass are considered.
func WithTopP(topP float32) Option {
	return func(c *client) {
		c.topP = topP
	}
}

func WithMaxChunkSize(maxChunkSize int) Option {
	return func(c *client) {
		c.maxChunkSize = maxChunkSize
	}
}
