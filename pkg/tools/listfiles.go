package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/appclacks/maizai/pkg/tools/aggregates"
)

type ToolListFilesParams struct {
	BaseDirectory string `json:"directory"`
}

type ToolListFiles struct {
}

func (t *ToolListFiles) Execute(jsonInput string) (string, error) {
	params := ToolListFilesParams{}
	err := json.Unmarshal([]byte(jsonInput), &params)
	if err != nil {
		return "", err
	}
	if params.BaseDirectory == "" {
		params.BaseDirectory = "."
	}
	result := []string{}
	err = filepath.Walk(params.BaseDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || strings.Contains(path, ".git/") {
				return nil
			}
			result = append(result, path)
			return nil
		})
	if err != nil {
		return "", err
	}
	return strings.Join(result, "\n"), nil
}

func (t *ToolListFiles) Spec() aggregates.Tool {
	return aggregates.Tool{
		Name:        "ListFiles",
		Description: "List all files recursively contained in the provided directory.",
		Properties: map[string]interface{}{
			"directory": map[string]interface{}{
				"type":        "string",
				"description": "The base directory that should be used to list files. THe parameter is optional, you shouldn't provide anything if the user is asking for the default value.",
			},
		},
	}
}
