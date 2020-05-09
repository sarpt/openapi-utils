package openapi

import "strings"

func isLocalReference(path string) bool {
	return strings.IndexRune(path, '#') == 0
}

func splitReference(path string) []string {
	return strings.Split(path, "#")
}

func getPathToRemoteDocument(path string) string {
	return splitReference(path)[0]
}

func getPathToReference(path string) string {
	return splitReference(path)[1]
}

func referencePathToItems(path string) []string {
	componentReference := getPathToReference(path)
	return strings.Split(componentReference, "/")[1:]
}
