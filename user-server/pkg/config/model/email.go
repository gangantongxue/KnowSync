package model

// EmailCfg 邮件配置项.
type EmailCfg struct {
	SMTPHost    string `yaml:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort    int    `yaml:"smtp_port" mapstructure:"smtp_port"`
	Username    string `yaml:"username" mapstructure:"username"`
	Password    string `yaml:"password" mapstructure:"password"`
	FromAddress string `yaml:"from_address" mapstructure:"from_address"`
	FromName    string `yaml:"from_name" mapstructure:"from_name"`
}
