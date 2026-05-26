package mail

import (
	_ "embed"
	"fmt"

	gomail "gopkg.in/gomail.v2"

	"github.com/gangantongxue/knowsync/user-server/pkg/config/model"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
)

//go:embed templates/verify_code.html
var verifyCodeTemplate string

// Mailer 邮件发送器
type Mailer struct {
	dialer   *gomail.Dialer
	from     string
	fromName string
	logger   *logger.Logger
}

// NewMailer 创建邮件发送器
func NewMailer(cfg *model.EmailCfg, l *logger.Logger) *Mailer {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.Username, cfg.Password)
	return &Mailer{
		dialer:   dialer,
		from:     cfg.FromAddress,
		fromName: cfg.FromName,
		logger:   l,
	}
}

// SendVerifyCode 发送验证码邮件
func (m *Mailer) SendVerifyCode(to, code string) error {
	body := fmt.Sprintf(verifyCodeTemplate, code)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "KnowSync - 邮箱验证码")
	msg.SetBody("text/html", body)

	if err := m.dialer.DialAndSend(msg); err != nil {
		m.logger.Logger.Error("发送验证码邮件失败", "to", to, "error", err)
		return fmt.Errorf("发送验证码邮件失败: %w", err)
	}

	m.logger.Logger.Info("验证码邮件发送成功", "to", to)
	return nil
}
