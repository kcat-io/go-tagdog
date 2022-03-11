package plugins

import (
	"context"
	"fmt"
	"github.com/kcat-io/go-plugin/plugin"
	"gopkg.in/yaml.v3"
	"log"
)

func init() {
	plugin.Register("alarm", Alarm())
}

// AlarmConfig Alarm主配置结构
type AlarmConfig struct {
	Root AlarmOption `yaml:"alarm"`
}

// AlarmOption Alarm配置项
type AlarmOption struct {
	Enable bool `yaml:"enable"`
}

// Alarm 用于执行告警通知，建议alarm规则排在第一位（将最后一个在执行after逻辑）
func Alarm() plugin.Plugin {
	return func(ctx context.Context, in, out interface{}, next plugin.NextHandle) (err error) {
		log.Printf("enter alarm")
		// before the next request
		if next != nil {
			err = next(ctx, in, out)
		}
		// after the last response
		isEnabled := false
		c := &AlarmConfig{}
		if input, ok := in.(*Input); ok {
			//解析配置文件，不使用err，避免污染链路返回
			localErr := yaml.Unmarshal(input.ConfigContent, c)
			if localErr != nil {
				log.Printf("ignore config decode err:%v", localErr)
			}
			isEnabled = c.Root.Enable == true
		}

		if isEnabled {
			if data, ok := out.(*Output); ok {
				// 这里只做打印，可以按需替换成其他通知途径，比如发邮件、发企业微信等
				for _, row := range *data {
					fmt.Printf("Alarm!!! %s::%s.%s -> %s\n",
						row.GetField("kind"), row.GetField("pod"), row.GetField("container"), row.Msg)
				}
			}
		}

		log.Printf("exit from alarm")
		return err
	}
}
