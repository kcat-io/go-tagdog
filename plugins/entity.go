package plugins

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

const configFile = "./config.yaml"

// Input 输入资源结构体
type Input struct {
	KubeConfig    string   `yaml:"kube_config"`
	Namespace     string   `yaml:"namespace"`
	Timeout       int      `yaml:"timeout"`
	Plugins       []string `yaml:"plugins,flow"`
	ClientSet     *kubernetes.Clientset
	ConfigContent []byte
}

// Output 输出资源数组
type Output []Result

// Result 输出资源结构体
type Result struct {
	Msg   string
	Field map[string]string
}

// GetField 获取Field中数据
func (r *Result) GetField(name string) string {
	if s, ok := r.Field[name]; ok {
		return s
	}
	return ""
}

// GetTimeout 获取超时时间
func (i *Input) GetTimeout() time.Duration {
	return time.Duration(i.Timeout)
}

// Load 加载配置
func (c *Input) Load() error {
	// 加载配置文件
	cf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	c.ConfigContent = cf

	//解析配置文件
	err = yaml.Unmarshal(cf, c)
	if err != nil {
		return err
	}

	// kubeconfig 检查
	if c.KubeConfig == "" {
		return errors.New("kube_config is required")
	}

	if _, err = os.Stat(c.KubeConfig); err != nil {
		errStr := fmt.Sprintf("file %s does not exist or not readable", c.KubeConfig)
		return errors.New(errStr)
	}

	// 构建k8s client
	cmd, err := clientcmd.BuildConfigFromFlags("", c.KubeConfig)
	if err != nil {
		return err
	}

	c.ClientSet, err = kubernetes.NewForConfig(cmd)
	if err != nil {
		return err
	}

	// 必要条件检查
	if len(c.Plugins) == 0 {
		return errors.New("there is no rule in config")
	}
	return nil
}
