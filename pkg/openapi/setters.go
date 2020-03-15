package openapi

func pathSetter(pathName string, targetDocument Document, referencedDocument Document) {
	targetPath, ok := targetDocument.Root.Paths[pathName]
	if !ok {
		targetPath = &Path{}
	}

	*targetPath = *referencedDocument.Root.Paths[pathName]
}

func responseSetter(responseName string, targetDocument Document, referencedDocument Document) {
	targetResponse, ok := targetDocument.Root.Components.Responses[responseName]
	if !ok {
		targetResponse = &Response{}
	}

	*targetResponse = *referencedDocument.Root.Components.Responses[responseName]
}
