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

package cmd

import (
	"fmt"
	"html"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/rammstein4o/git-gpt/git"
	"github.com/rammstein4o/git-gpt/gpt"
	"github.com/rammstein4o/git-gpt/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Auto generate commit message",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := viper.GetString("mode")

		if err := utils.CwdToGitRoot(); err != nil {
			return err
		}

		gitHelper := git.New(
			git.WithExcludeList(viper.GetStringSlice("git.exclude_list")),
		)

		var topP float32
		if err := viper.UnmarshalKey("completion.top_p", &topP); err != nil {
			topP = 1.0
		}

		var temperature float32
		if err := viper.UnmarshalKey("completion.temperature", &temperature); err != nil {
			temperature = 0.4
		}

		gptOptions := []gpt.Option{
			gpt.WithMaxTokens(viper.GetInt("completion.max_tokens")),
			gpt.WithTopP(topP),
			gpt.WithTemperature(temperature),
			gpt.WithStream(viper.GetBool("completion.stream")),
			gpt.WithMaxChunkSize(viper.GetInt("commit.maxChunkSize")),
		}

		if mode == "azure_open_ai" {
			gptOptions = append(gptOptions, gpt.WithAzureOpenAI(
				viper.GetString("azure_open_ai.api_key"),
				viper.GetString("azure_open_ai.endpoint"),
				viper.GetString("azure_open_ai.model"),
				viper.GetString("azure_open_ai.alias"),
			))
		} else {
			gptOptions = append(gptOptions, gpt.WithOpenAI(
				viper.GetString("open_ai.api_key"),
				viper.GetString("open_ai.model"),
			))
		}

		gptHelper := gpt.New(
			gptOptions...,
		)

		names, err := gitHelper.DiffNames()
		if err != nil {
			return err
		}
		if names == "" {
			return fmt.Errorf("please add your staged changes using git add <files...>")
		}

		color.Green("Summarize the stashed changes")

		status, err := gitHelper.Status()
		if err != nil {
			return err
		}

		addedFiles, removedFiles, modifiedFiles := git.ParseGitStatus(status)

		changeSummaries := make([]string, 0)

		for _, fileName := range addedFiles {
			if utils.IsBinaryFile(fileName) {
				changeSummaries = append(changeSummaries, fmt.Sprintf("Added binary file `%s`", fileName))
				continue
			}

			fileContent, err := utils.ReadFile(fileName)
			if err != nil {
				return err
			}

			summary, err := gptHelper.SummarizeFile(
				cmd.Context(),
				git.OPERATION_ADD,
				fileName,
				strings.TrimSpace(fileContent),
			)
			if err != nil {
				return err
			}

			changeSummaries = append(changeSummaries, summary)
		}

		for _, fileName := range removedFiles {
			if utils.IsBinaryFile(fileName) {
				changeSummaries = append(changeSummaries, fmt.Sprintf("Removed binary file `%s`", fileName))
				continue
			}

			fileContent, err := gitHelper.ShowDeletedFile(fileName)
			if err != nil {
				return err
			}

			summary, err := gptHelper.SummarizeFile(
				cmd.Context(),
				git.OPERATION_DEL,
				fileName,
				strings.TrimSpace(fileContent),
			)
			if err != nil {
				return err
			}

			changeSummaries = append(changeSummaries, summary)
		}

		for _, fileName := range modifiedFiles {
			if utils.IsBinaryFile(fileName) {
				changeSummaries = append(changeSummaries, fmt.Sprintf("Replaced binary file `%s`", fileName))
				continue
			}

			diff, err := gitHelper.DiffFile(fileName)
			if err != nil {
				return err
			}

			summary, err := gptHelper.SummarizeDiff(cmd.Context(), fileName, diff)
			if err != nil {
				return err
			}

			changeSummaries = append(changeSummaries, summary)
		}

		summary, err := gptHelper.SummarizeChanges(cmd.Context(), changeSummaries)
		if err != nil {
			return err
		}

		commitMessage, err := gptHelper.FinalizeCommitMsg(cmd.Context(), summary)
		if err != nil {
			return err
		}

		// unescape html entities in commit message
		commitMessage = html.UnescapeString(commitMessage)

		// Output commit summary data from AI
		color.Yellow("================Commit Summary====================")
		color.Yellow("\n" + strings.TrimSpace(commitMessage) + "\n\n")
		color.Yellow("==================================================")

		stats := gptHelper.GetStats(cmd.Context())
		color.Magenta(stats.String())

		// pricePerThousand := 0.0
		// if tmp, ok := costMap[selectedModel]; ok {
		// 	pricePerThousand = tmp
		// }
		// if pricePerThousand > 0.0 {
		// 	color.Magenta(
		// 		"Total cost: ~$%.2f\n",
		// 		calculateCost(totalTokens, pricePerThousand),
		// 	)
		// } else {
		// 	color.Magenta("Unknown cost\n")
		// }

		outputFile := viper.GetString("commit.file")
		if outputFile == "" {
			out, err := gitHelper.GitDir()
			if err != nil {
				return err
			}
			outputFile = path.Join(strings.TrimSpace(out), "COMMIT_EDITMSG")
		}
		color.Cyan("Write the commit message to " + outputFile + " file")
		// write commit message to git staging file
		err = os.WriteFile(outputFile, []byte(commitMessage), 0o644)
		if err != nil {
			return err
		}

		if viper.GetBool("commit.preview") {
			return nil
		}

		// git commit automatically
		color.Cyan("Git record changes to the repository")
		output, err := gitHelper.Commit(commitMessage)
		if err != nil {
			return err
		}
		color.Yellow(output)

		return nil
	},
}
