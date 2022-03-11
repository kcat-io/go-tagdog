package main

import (
	"context"
	"github.com/kcat-io/go-plugin/plugin"
	"github.com/kcat-io/go-tagdog/plugins"
	"log"
	"time"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	in := &plugins.Input{}
	out := &plugins.Output{}

	// 加载配置
	err := in.Load()
	if err != nil {
		log.Fatal(err)
		return
	}

	// 加载规则
	var r = plugin.Plugins{}
	for _, p := range in.Plugins {
		f := plugin.GetPlugin(p)
		if f == nil {
			log.Fatalf("Plugin %s is not registered yet.", p)
		}
		r = append(r, f)
	}

	// 创建会话
	ctx, cancel := context.WithTimeout(context.Background(), in.GetTimeout()*time.Second)
	defer cancel()

	// 执行规则
	err = r.Handle(ctx, in, out, nil)
	if err != nil {
		log.Fatal(err)
	}

}
