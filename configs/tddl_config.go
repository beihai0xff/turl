package configs

// TDDLConfig is the configuration for tddl
type TDDLConfig struct {
	// Step is the step of the sequence
	Step uint64 `json:"step" yaml:"step" mapstructure:"step"`
	// SeqName is the name of the sequence
	SeqName string `json:"seq_name" yaml:"seq_name" mapstructure:"seq_name"`
	// StartNum is the start number of the sequence
	StartNum uint64 `json:"start_num" yaml:"start_num" mapstructure:"start_num"`
}
