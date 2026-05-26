package model

// EmailCfg 邮件配置项
type EmailCfg struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	FromAddress  string `yaml:"from_address"`
	FromName     string `yaml:"from_name"`
}
