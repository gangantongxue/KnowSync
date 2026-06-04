package tool

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type CurrentTime struct{}

func NewCurrentTime() *CurrentTime {
	return &CurrentTime{}
}

func (c *CurrentTime) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "current_time",
		Desc:        "获取当前的日期和时间信息，包括年月日、星期、时分秒和时区。当你需要知道当前时间或计算时间相关问题时使用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (c *CurrentTime) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return c.execute()
}

func (c *CurrentTime) execute() (string, error) {
	now := time.Now()
	loc := now.Location()

	type timeInfo struct {
		DateTime    string `json:"datetime"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		Timezone    string `json:"timezone"`
		TimezoneUTC string `json:"timezone_utc"`
		Weekday     string `json:"weekday"`
		Unix        int64  `json:"unix"`
		Year        int    `json:"year"`
		Month       string `json:"month"`
		Day         int    `json:"day"`
		Hour        int    `json:"hour"`
		Minute      int    `json:"minute"`
		Second      int    `json:"second"`
	}

	_, offset := now.Zone()
	utcOffset := ""
	if offset >= 0 {
		utcOffset = "+"
	}
	utcOffset += formatInt(offset/3600) + ":" + formatInt((offset%3600)/60)

	data, _ := json.Marshal(timeInfo{
		DateTime:    now.Format("2006-01-02 15:04:05"),
		Date:        now.Format("2006-01-02"),
		Time:        now.Format("15:04:05"),
		Timezone:    loc.String(),
		TimezoneUTC: "UTC" + utcOffset,
		Weekday:     now.Weekday().String(),
		Unix:        now.Unix(),
		Year:        now.Year(),
		Month:       now.Month().String(),
		Day:         now.Day(),
		Hour:        now.Hour(),
		Minute:      now.Minute(),
		Second:      now.Second(),
	})
	return string(data), nil
}

func formatInt(n int) string {
	if n < 0 {
		n = -n
	}
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
