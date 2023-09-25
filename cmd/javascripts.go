package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/klippa-app/pdfium-cli/pdf"

	"github.com/klippa-app/go-pdfium/requests"
	"github.com/spf13/cobra"
)

func init() {
	addGenericPDFOptions(javascriptsCmd)
	rootCmd.AddCommand(javascriptsCmd)
}

var javascriptsCmd = &cobra.Command{
	Use:   "javascripts [input] [output-folder]",
	Short: "Extract the javascripts of a PDF",
	Long:  "Extract the javascripts of a PDF and store them as file.\n[input] can either be a file path or - for stdin.\n[output-folder] can be either a folder or - for stdout. In the case of stdout, multiple files will be delimited by the value of the std-file-delimiter, with a newline before and after it.",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return newExitCodeError(err, ExitCodeInvalidArguments)
		}

		if err := validFile(args[0]); err != nil {
			return fmt.Errorf("could not open input file %s: %w\n", args[0], newExitCodeError(err, ExitCodeInvalidInput))
		}

		if args[1] != stdFilename {
			folderStat, err := os.Stat(args[1])
			if err != nil {
				return fmt.Errorf("could not open output folder %s: %w\n", args[1], newExitCodeError(err, ExitCodeInvalidOutput))
			}

			if !folderStat.IsDir() {
				return newExitCodeError(fmt.Errorf("output folder %s is not a folder\n", args[1]), ExitCodeInvalidOutput)
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := pdf.LoadPdfium()
		if err != nil {
			handleError(cmd, fmt.Errorf("could not load pdfium: %w\n", newPdfiumError(err)), ExitCodePdfiumError)
			return
		}
		defer pdf.ClosePdfium()

		document, closeFile, err := openFile(args[0])
		if err != nil {
			handleError(cmd, fmt.Errorf("could not open input file %s: %w\n", args[0], err), ExitCodeInvalidInput)
			return
		}
		defer closeFile()

		javascripts, err := pdf.PdfiumInstance.GetJavaScriptActions(&requests.GetJavaScriptActions{
			Document: document.Document,
		})
		if err != nil {
			handleError(cmd, fmt.Errorf("could not get javascript count for PDF %s: %w\n", args[0], newPdfiumError(err)), ExitCodePdfiumError)
			return
		}

		if len(javascripts.JavaScriptActions) > 0 {
			for i := 0; i < len(javascripts.JavaScriptActions); i++ {
				if args[1] != stdFilename {
					filePath := path.Join(args[1], fmt.Sprintf("%s.js", javascripts.JavaScriptActions[i].Name))
					outFile, err := os.Create(filePath)
					if err != nil {
						handleError(cmd, fmt.Errorf("could not create output file for javascript %d for PDF %s: %w\n", i, args[0], err), ExitCodeInvalidOutput)
						return
					}

					outFile.Write([]byte(javascripts.JavaScriptActions[i].Script))
					outFile.Close()

					cmd.Printf("Exported javascript %d into %s\n", i+1, filePath)
				} else {
					if i > 0 {
						os.Stdout.WriteString("\n")
						os.Stdout.WriteString(stdFileDelimiter)
						os.Stdout.WriteString("\n")
					}
					os.Stdout.Write([]byte(javascripts.JavaScriptActions[i].Script))
				}
			}
		}
	},
}
