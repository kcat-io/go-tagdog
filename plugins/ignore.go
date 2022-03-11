package plugins

import (
	"context"
	"fmt"
	"github.com/kcat-io/go-plugin/plugin"
	"gopkg.in/yaml.v3"
	"log"
	"time"
)

func init() {
	plugin.Register("ignore", Ignore())
}

type IgnoreConfig struct {
	Root IgnoreList `yaml:"ignore"`
}

type IgnoreList struct {
	Enable bool              `yaml:"enable"`
	List   map[string]string `yaml:"list"`
}

// IsIgnorePath 判定路径是否在忽略名单中
func (c *IgnoreConfig) IsIgnorePath(path string) bool {
	now := time.Now()
	loc, _ := time.LoadLocation("Local")
	//fmt.Println(path)
	// 找到配置
	if stime, ok := c.Root.List[path]; ok {
		//fmt.Println(path)
		// 判定时间
		if "" == stime {
			return true
		}
		stime, err := time.ParseInLocation("2006-01-02 15:04:05", stime, loc)
		if err != nil {
			// 时间格式配置错误，当未配置时间处理
			log.Printf("datetime(%s) format err:%v", stime, err)
			return true
		}
		// 时间还没有过，还在屏蔽期间
		if now.Sub(stime) < 0 {
			return true
		}
	}
	return false
}

// Ignore 按配置进行忽略，建议该规则排在最后
func Ignore() plugin.Plugin {
	return func(ctx context.Context, in, out interface{}, next plugin.NextHandle) (err error) {
		log.Printf("enter ignore")
		isEnabled := false
		c := &IgnoreConfig{}
		if input, ok := in.(*Input); ok {
			//解析配置文件
			err = yaml.Unmarshal(input.ConfigContent, c)
			if err != nil {
				log.Printf("ignore config decode err:%v", err)
			}
			isEnabled = c.Root.Enable == true
		}
		// 若规则启用和存在过滤清单
		if isEnabled && len(c.Root.List) > 0 {
			if data, ok := out.(*Output); ok {
				// 遍历out数据
				for i := 0; i < len(*data); i++ {
					path := fmt.Sprintf("%s::%s.%s",
						(*data)[i].GetField("kind"), (*data)[i].GetField("pod"), (*data)[i].GetField("container"))
					// 判断path是否在过滤清单中
					// log.Printf("is path(%s) in ignore list?", path)
					if c.IsIgnorePath(path) {
						log.Printf("path(%s) in ignore list", path)
						*data = append((*data)[:i], (*data)[i+1:]...)
						i-- // form the remove item index to start iterate next item
					}
				}
			}

		}
		// before the next request
		if next != nil {
			err = next(ctx, in, out)
		}
		// after the last response
		log.Printf("exit from ignore")
		return err
	}
}
