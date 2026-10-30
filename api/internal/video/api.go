package video

import "github.com/smashcut/smashcut/api/internal/video/template"

func TemplateDefault() template.Spec { return template.Default() }

func ValidateTemplate(s template.Spec) error { return template.Validate(s) }

func ParseTemplate(raw []byte) (template.Spec, error) { return template.Parse(raw) }
