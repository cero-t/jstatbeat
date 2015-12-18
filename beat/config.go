package beat

type JstatConfig struct {
	Interval *string `yaml:"interval"`
	Name     *string `yaml:"name"`
}

type ConfigSettings struct {
	Input JstatConfig
}
