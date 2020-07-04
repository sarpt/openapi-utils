package openapi

import (
	"fmt"
	"strings"
)

const (
	referenceSeparator = '#'
	pathSeparator      = "/"
)

func isLocalReference(path string) bool {
	return strings.IndexRune(path, referenceSeparator) == 0
}

func splitReferenceByHash(path string) []string {
	return strings.Split(path, string(referenceSeparator))
}

func getDocumentPath(path string) string {
	return splitReferenceByHash(path)[0]
}

func getPathToReference(path string) string {
	if isLocalReference(path) {
		return path
	}

	return splitReferenceByHash(path)[1]
}

func convertRemoteToLocalPath(path string) string {
	return fmt.Sprintf("%s%s", string(referenceSeparator), getPathToReference(path))
}

func referencePathToItems(path string) []string {
	componentReference := getPathToReference(path)
	return strings.Split(componentReference, pathSeparator)[1:]
}

func sortReferences(refI, refJ reference) bool {
	isILocal := isLocalReference(refI.path)
	isJLocal := isLocalReference(refJ.path)

	if !isILocal && isJLocal {
		return true
	}

	return false
}
