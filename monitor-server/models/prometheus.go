package models

type RuleFile struct {
	Groups  []*RFGroup  `yaml:"groups"`
}

// RF -> RuleFile
type RFGroup struct {
	Name  string  `yaml:"name"`
	Rules  []*RFRule  `yaml:"rules"`
}

type RFRule struct {
	Alert  string  `yaml:"alert"`
	Expr  string  `yaml:"expr"`
	For  string  `yaml:"for"`
	Labels  map[string]string  `yaml:"labels"`
	Annotations  RFAnnotation  `yaml:"annotations"`
}

type RFAnnotation struct {
	Summary  string  `yaml:"summary"`
	Description  string  `yaml:"description"`
}
