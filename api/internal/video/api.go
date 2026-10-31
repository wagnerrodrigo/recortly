package video

import "github.com/recortly/recortly/api/internal/video/template"

// Aliases para simplificar o uso no worker e na API.
type TemplateSpec = template.Spec
type ZoneBox = template.ZoneBox

func TemplateDefault() template.Spec { return template.Default() }

func ValidateTemplate(s template.Spec) error { return template.Validate(s) }

func ParseTemplate(raw []byte) (template.Spec, error) { return template.Parse(raw) }

func BuildLayout(s template.Spec) template.Layout { return template.BuildLayout(s) }
