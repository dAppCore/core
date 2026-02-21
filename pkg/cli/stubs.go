package cli

// ──────────────────────────────────────────────────────────────────────────────
// Form (stubbed — simple fallback, will use charmbracelet/huh later)
// ──────────────────────────────────────────────────────────────────────────────

// FieldType defines the type of a form field.
type FieldType string

const (
	FieldText     FieldType = "text"
	FieldPassword FieldType = "password"
	FieldConfirm  FieldType = "confirm"
	FieldSelect   FieldType = "select"
)

// FormField describes a single field in a form.
type FormField struct {
	Label       string
	Key         string
	Type        FieldType
	Default     string
	Placeholder string
	Options     []string // For FieldSelect
	Required    bool
	Validator   func(string) error
}

// Form presents a multi-field form and returns the values keyed by FormField.Key.
// Currently falls back to sequential Question()/Confirm()/Select() calls.
// Will be replaced with charmbracelet/huh interactive form later.
//
//	results, err := cli.Form([]cli.FormField{
//	    {Label: "Name", Key: "name", Type: cli.FieldText, Required: true},
//	    {Label: "Password", Key: "pass", Type: cli.FieldPassword},
//	    {Label: "Accept terms?", Key: "terms", Type: cli.FieldConfirm},
//	})
func Form(fields []FormField) (map[string]string, error) {
	results := make(map[string]string, len(fields))

	for _, f := range fields {
		switch f.Type {
		case FieldPassword:
			val := Question(f.Label+":", WithDefault(f.Default))
			results[f.Key] = val
		case FieldConfirm:
			if Confirm(f.Label) {
				results[f.Key] = "true"
			} else {
				results[f.Key] = "false"
			}
		case FieldSelect:
			val, err := Select(f.Label, f.Options)
			if err != nil {
				return nil, err
			}
			results[f.Key] = val
		default: // FieldText
			var opts []QuestionOption
			if f.Default != "" {
				opts = append(opts, WithDefault(f.Default))
			}
			if f.Required {
				opts = append(opts, RequiredInput())
			}
			if f.Validator != nil {
				opts = append(opts, WithValidator(f.Validator))
			}
			results[f.Key] = Question(f.Label+":", opts...)
		}
	}

	return results, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// FilePicker (stubbed — will use charmbracelet/filepicker later)
// ──────────────────────────────────────────────────────────────────────────────

// FilePickerOption configures FilePicker behaviour.
type FilePickerOption func(*filePickerConfig)

type filePickerConfig struct {
	dir        string
	extensions []string
}

// InDirectory sets the starting directory for the file picker.
func InDirectory(dir string) FilePickerOption {
	return func(c *filePickerConfig) {
		c.dir = dir
	}
}

// WithExtensions filters to specific file extensions (e.g. ".go", ".yaml").
func WithExtensions(exts ...string) FilePickerOption {
	return func(c *filePickerConfig) {
		c.extensions = exts
	}
}

// FilePicker presents a file browser and returns the selected path.
// Currently falls back to a text prompt. Will be replaced with an
// interactive file browser later.
//
//	path, err := cli.FilePicker(cli.InDirectory("."), cli.WithExtensions(".go"))
func FilePicker(opts ...FilePickerOption) (string, error) {
	cfg := &filePickerConfig{dir: "."}
	for _, opt := range opts {
		opt(cfg)
	}

	hint := "File path"
	if cfg.dir != "." {
		hint += " (from " + cfg.dir + ")"
	}
	return Question(hint + ":"), nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Tabs (stubbed — will use bubbletea model later)
// ──────────────────────────────────────────────────────────────────────────────

// TabItem describes a tab with a title and content.
type TabItem struct {
	Title   string
	Content string
}

// Tabs displays tabbed content. Currently prints all tabs sequentially.
// Will be replaced with an interactive tab switcher later.
//
//	cli.Tabs([]cli.TabItem{
//	    {Title: "Overview", Content: summaryText},
//	    {Title: "Details", Content: detailText},
//	})
func Tabs(items []TabItem) error {
	for i, tab := range items {
		if i > 0 {
			Blank()
		}
		Section(tab.Title)
		Println("%s", tab.Content)
	}
	return nil
}
