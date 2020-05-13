package openapi

import "strings"

func isLocalReference(path string) bool {
	return strings.IndexRune(path, '#') == 0
}

func splitReference(path string) []string {
	return strings.Split(path, "#")
}

func getDocumentPath(path string) string {
	return splitReference(path)[0]
}

func getPathToReference(path string) string {
	return splitReference(path)[1]
}

func referencePathToItems(path string) []string {
	componentReference := getPathToReference(path)
	return strings.Split(componentReference, "/")[1:]
}

func sortReferences(refI, refJ reference) bool {
	isILocal := isLocalReference(refI.path)
	isJLocal := isLocalReference(refJ.path)

	if !isILocal && isJLocal {
		return true
	}

	return false
}
