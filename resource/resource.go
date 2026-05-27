package resource

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
	metaui "github.com/viant/mcp-ui/meta"
)

// ContentHash returns the canonical UI resource content hash for the supplied
// text payload.
func ContentHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// NewHTMLResource constructs a UI resource declaration for a `ui://...` HTML
// resource.
func NewHTMLResource(uri, name string, description, title *string, annotations *schema.Annotations, ui metaui.ResourceUI) (*schema.Resource, error) {
	parsed, err := ValidateUIURI(uri)
	if err != nil {
		return nil, err
	}
	_ = parsed
	if name == "" {
		return nil, fmt.Errorf("ui resource: name is required")
	}
	mimeType := capabilities.ResourceMimeType
	res := &schema.Resource{
		Name:        name,
		Uri:         uri,
		MimeType:    &mimeType,
		Description: description,
		Title:       title,
		Annotations: annotations,
	}
	metaui.SetResourceUI(res, ui)
	return res, nil
}

// NewHTMLContents constructs a text resource contents entry for HTML UI.
func NewHTMLContents(uri, html string, ui metaui.ResourceUI) (*schema.TextResourceContents, error) {
	if _, err := ValidateUIURI(uri); err != nil {
		return nil, err
	}
	if html == "" {
		return nil, fmt.Errorf("ui resource contents: html text is required")
	}
	mimeType := capabilities.ResourceMimeType
	contents := &schema.TextResourceContents{
		Uri:      uri,
		MimeType: &mimeType,
		Text:     html,
	}
	metaui.SetTextResourceContentsUI(contents, ui)
	return contents, nil
}

// NewReadResultHTMLContents constructs a resources/read payload element for
// HTML UI content.
func NewReadResultHTMLContents(uri, html string, ui metaui.ResourceUI) (*schema.ReadResourceResultContentsElem, error) {
	if _, err := ValidateUIURI(uri); err != nil {
		return nil, err
	}
	if html == "" {
		return nil, fmt.Errorf("ui read contents: html text is required")
	}
	mimeType := capabilities.ResourceMimeType
	contents := &schema.ReadResourceResultContentsElem{
		Uri:      uri,
		MimeType: &mimeType,
		Text:     html,
	}
	metaui.SetReadResultContentsUI(contents, ui)
	return contents, nil
}

// NewEmbeddedHTMLResource constructs an embedded HTML UI resource for optional
// compatibility flows.
func NewEmbeddedHTMLResource(uri, html string, annotations *schema.Annotations, ui metaui.ResourceUI) (*schema.EmbeddedResource, error) {
	if _, err := ValidateUIURI(uri); err != nil {
		return nil, err
	}
	if html == "" {
		return nil, fmt.Errorf("embedded ui resource: html text is required")
	}
	mimeType := capabilities.ResourceMimeType
	embedded := &schema.EmbeddedResource{
		Type:        "resource",
		Annotations: annotations,
		Resource: schema.EmbeddedResourceResource{
			Uri:      uri,
			MimeType: &mimeType,
			Text:     html,
		},
	}
	metaui.SetEmbeddedResourceUI(embedded, ui)
	return embedded, nil
}
