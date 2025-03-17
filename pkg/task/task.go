package task

import (
	"fmt"
	"github.com/aoaostar/mooc/pkg/config"
	"github.com/aoaostar/mooc/pkg/yinghua"
	"github.com/aoaostar/mooc/pkg/yinghua/types"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"net/smtp"
)

type Task struct {
	User   config.User
	Course types.CoursesList
	Status bool
}

var Tasks []Task

// SendQQEmail 用于通过 QQ 邮箱发送邮件
func SendQQEmail(from, password, to, subject, body string) error {
    // QQ 邮箱的 SMTP 服务器地址和端口
    smtpServer := "smtp.qq.com:587"

    // 构建邮件内容
    message := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s", to, from, subject, body))

    // 创建认证信息
    auth := smtp.PlainAuth("", from, password, "smtp.qq.com")

    // 发送邮件
    err := smtp.SendMail(smtpServer, auth, from, []string{to}, message)
    if err != nil {
        return err
    }
    return nil
}

func Start() {
	limit := int(math.Min(float64(config.Conf.Global.Limit), float64(len(Tasks))))
	jobs := make(chan Task, limit)
	wg := sync.WaitGroup{}
	for i := 0; i < limit; i++ {
		go func() {
			defer wg.Done()
			for job := range jobs {
				work(job)
			}
		}()
		wg.Add(1)
	}

	logrus.Infof("任务系统启动成功, 协程数: %d, 任务数: %d", limit, len(Tasks))

	for _, task := range Tasks {
		jobs <- task
	}
	close(jobs)
	wg.Wait()
	from := "panxun188@qq.com"
    	password := "pasknjesvbzqdebh"
    	to := "panxun188@qq.com"
    	subject := "刷课完成"
    	body := "所有任务已经完成。"
	logrus.Infof("恭喜您, 所有任务都已全部完成~~~")
	SendQQEmail(from, password, to, subject, body)
        fmt.Println("邮件发送成功")
    	
}

func work(task Task) {
	instance := yinghua.New(task.User)
	err := instance.Login()
	if err != nil {
		logrus.Fatal(err)
	}

	instance.Output("登录成功")

	if task.Course.Progress == 1 {
		instance.Output(fmt.Sprintf("当前课程[%s][%d] 进度: %s, 跳过", task.Course.Name, task.Course.ID, task.Course.Progress1))
		return
	}
	if task.Course.State == 2 {
		instance.Output(fmt.Sprintf("当前课程[%s][%d] 已结束, 跳过", task.Course.Name, task.Course.ID))
		return
	}
	instance.Output(fmt.Sprintf("当前课程[%s][%d] 进度: %s", task.Course.Name, task.Course.ID, task.Course.Progress1))
	err = instance.StudyCourse(task.Course)
	if err != nil {
		instance.OutputWith(fmt.Sprintf("课程[%s][%d]: %s", task.Course.Name, task.Course.ID, err.Error()), logrus.Errorf)
	}

}
