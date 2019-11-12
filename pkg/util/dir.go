package util

import "strings"

func AssembleDir(rootDir string, childDir string) string {
	return strings.TrimSuffix(rootDir, "/") + "/" + strings.TrimPrefix(childDir, "/")
}
