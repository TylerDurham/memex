package document

import (
	"errors"
	"fmt"
)

func ErrorGeneratingLauncURL(reason string, doc *IndexDocument) error {
	return errors.New(fmt.Sprintf("cannot generate launch url for indexable document: '%s': %s", doc.DocPath, reason))
}
