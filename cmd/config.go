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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".git-gpt" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".git-gpt")

	cobra.CheckErr(viper.ReadInConfig())
}

func init() {
	cobra.OnInitialize(initConfig)

	commitCmd.PersistentFlags().StringP("file", "f", "", "commit message file")
	viper.BindPFlag("commit.file", commitCmd.PersistentFlags().Lookup("file"))

	commitCmd.PersistentFlags().Bool("preview", false, "preview commit message")
	viper.BindPFlag("commit.preview", commitCmd.PersistentFlags().Lookup("preview"))

	commitCmd.PersistentFlags().Int("maxChunkSize", 6000, "split big diffs into chunks with this maximum size")
	viper.BindPFlag("commit.maxChunkSize", commitCmd.PersistentFlags().Lookup("maxChunkSize"))

	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(completionCmd)
}
