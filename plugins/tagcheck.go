package plugins

import (
	"context"
	"fmt"
	"github.com/kcat-io/go-plugin/plugin"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strings"
)

func init() {
	plugin.Register("tagcheck", TagCheck())
}

// TagConfig 配置文件结构
type TagConfig struct {
	Root TagOption `yaml:"tagcheck"`
}

// TagOption 配置文件结构
type TagOption struct {
	Enable       bool              `yaml:"enable"`
	List         map[string]string `yaml:"list,flow"`
	CheckCronjob bool              `yaml:"cronjob"`
}

// TagCheck 按配置校验指定的镜像版本是否正确
// 自动校验cronjob中镜像tag与deployment是否一致
func TagCheck() plugin.Plugin {
	return func(ctx context.Context, in, out interface{}, next plugin.NextHandle) (err error) {
		log.Printf("enter tagcheck")

		isEnabled := false
		c := &TagConfig{}
		if input, ok := in.(*Input); ok {
			//解析配置文件，不使用err，避免污染链路返回
			localErr := yaml.Unmarshal(input.ConfigContent, c)
			if localErr != nil {
				log.Printf("ignore config decode err:%v", localErr)
			}
			isEnabled = c.Root.Enable == true
		}
		pIn := in.(*Input)
		pOut := out.(*Output)
		if isEnabled && (c.Root.CheckCronjob || len(c.Root.List) > 0) {
			// 获取deployment列表
			deploymentList, localErr := (*pIn).ClientSet.AppsV1().Deployments((*pIn).Namespace).List(ctx, metav1.ListOptions{})
			if localErr != nil {
				log.Printf("get deployment list err:%v", localErr)
			}

			imageMap := make(map[string]string)
			for _, d := range deploymentList.Items {
				// 获取deployment详情
				deployment, localErr := (*pIn).ClientSet.AppsV1().Deployments((*pIn).Namespace).Get(ctx, d.Name, metav1.GetOptions{})
				if err != nil {
					log.Printf("deployment get error:%v", localErr)
					continue
				}

				containers := &deployment.Spec.Template.Spec.Containers
				for _, container := range *containers {
					// 提取镜像名称和Tag
					image := strings.Split(container.Image, ":")[0]
					tag := strings.Split(container.Image, ":")[1]
					if _, ok := imageMap[image]; !ok {
						// 存储镜像与tag以备检查cronjob中tag的一致性
						imageMap[image] = tag
					}
					if targetTag, ok := c.Root.List[image]; ok && targetTag != tag {
						// 指定镜像的tag不匹配
						o := Result{}
						o.Msg = fmt.Sprintf("tag not match, need(%s) ,got(%s)", targetTag, tag)
						o.Field = map[string]string{
							"kind":      "deployment",
							"pod":       d.Name,
							"container": container.Name,
							"image":     image,
							"tag":       tag,
						}
						*pOut = append(*pOut, o)
					}
				}
			}

			// 检查cronjob中tag的一致性
			if c.Root.CheckCronjob && len(imageMap) > 0 {
				cronjobList, localErr := (*pIn).ClientSet.BatchV1beta1().CronJobs((*pIn).Namespace).List(ctx, metav1.ListOptions{})
				if localErr != nil {
					log.Printf("get cronjob list err:%v", localErr)
				}

				for _, cj := range cronjobList.Items {
					cronjob, localErr := (*pIn).ClientSet.BatchV1beta1().CronJobs((*pIn).Namespace).Get(ctx, cj.Name, metav1.GetOptions{})
					if localErr != nil {
						log.Printf("deployment get error:%v", localErr)
						continue
					}

					if *cronjob.Spec.Suspend {
						// 跳过暂停的任务
						continue
					}
					containers := &cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers
					for _, container := range *containers {
						image := strings.Split(container.Image, ":")[0]
						tag := strings.Split(container.Image, ":")[1]
						if _, ok := imageMap[image]; ok && tag != imageMap[image] {
							// 指定镜像的tag不匹配
							o := Result{}
							o.Msg = fmt.Sprintf("tag not match, need(%s) ,got(%s)", imageMap[image], tag)
							o.Field = map[string]string{
								"kind":      "cronjob",
								"pod":       cj.Name,
								"container": container.Name,
								"image":     image,
								"tag":       tag,
							}
							*pOut = append(*pOut, o)
						}
					}
				}
			}
		}

		// before the next request
		if next != nil {
			err = next(ctx, in, out)
		}
		// after the last response
		log.Printf("exit from tagcheck")
		return err
	}
}
