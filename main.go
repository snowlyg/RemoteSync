package main

import (
	"flag"
	"fmt"
	"github.com/snowlyg/RemoteSync/logging"
	"github.com/snowlyg/RemoteSync/models"
	"github.com/snowlyg/RemoteSync/utils"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/service"
)

var Version string

type program struct {
	httpServer *http.Server
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	//err := models.Init()
	//if err != nil {
	//	panic(err)
	//}
	//err = routers.Init()
	//if err != nil {
	//	panic(err)
	//}

	//go syncDeviceLog()
	//go syncDevice()

}

//func syncDevice() {
//	t := utils.Conf().Section("time").Key("sync_data_time").MustInt64(1)
//	v := utils.Conf().Section("time").Key("sync_data").MustString("h")
//	var chSy chan int
//	var tickerSync *time.Ticker
//	switch v {
//	case "h":
//		tickerSync = time.NewTicker(time.Hour * time.Duration(t))
//	case "m":
//		tickerSync = time.NewTicker(time.Minute * time.Duration(t))
//	case "s":
//		tickerSync = time.NewTicker(time.Second * time.Duration(t))
//	default:
//		tickerSync = time.NewTicker(time.Hour * time.Duration(t))
//	}
//	go func() {
//		for range tickerSync.C {
//			utils.GetToken()
//			sync.SyncDevice()
//		}
//		chSy <- 1
//	}()
//	<-chSy
//}
//
//func syncDeviceLog() {
//	var ch chan int
//	var t int64
//	t = utils.Conf().Section("time").Key("sync_log_time").MustInt64(4)
//	v := utils.Conf().Section("time").Key("sync_log").MustString("m")
//	var ticker *time.Ticker
//
//	ticker = time.NewTicker(time.Hour * time.Duration(t))
//	switch v {
//	case "h":
//		ticker = time.NewTicker(time.Hour * time.Duration(t))
//	case "m":
//		ticker = time.NewTicker(time.Minute * time.Duration(t))
//	case "s":
//		ticker = time.NewTicker(time.Second * time.Duration(t))
//	default:
//		ticker = time.NewTicker(time.Minute * time.Duration(t))
//	}
//	sync.NotFirst = false
//	go func() {
//		for range ticker.C {
//			utils.GetToken()
//			go func() {
//				sync.CheckRestful()
//			}()
//			go func() {
//				sync.CheckService()
//			}()
//			// 进入当天目录,跳过 23点45 当天凌晨 0点15 分钟，给设备创建目录的时间
//			if !((time.Now().Hour() == 0 && time.Now().Minute() < 15) || (time.Now().Hour() == 23 && time.Now().Minute() > 45)) {
//				go func() {
//					sync.SyncDeviceLog()
//				}()
//			}
//			sync.NotFirst = true
//		}
//		ch <- 1
//	}()
//	<-ch
//}

//func (p *program) StopHTTP() (err error) {
//	if p.httpServer == nil {
//		err = fmt.Errorf("HTTP Server Not Found")
//		return
//	}
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err = p.httpServer.Shutdown(ctx); err != nil {
//		return
//	}
//	return
//}

func (p *program) Stop(s service.Service) error {
	defer log.Println("********** STOP **********")
	//defer utils.CloseLogWriter()
	//_ = p.StopHTTP()
	//models.Close()
	return nil
}

var Action = flag.String("action", "", "程序操作指令")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [command]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  -action <install remove start stop restart version>\n")
		fmt.Fprintf(os.Stderr, "    程序操作指令\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	// 初始化日志目录
	exeName := utils.EXEName()
	svcConfig := &service.Config{
		Name:        exeName,    //服务显示名称
		DisplayName: exeName,    //服务名称
		Description: "远程探视数据同步", //服务描述
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logging.Err.Error(err)
	}

	if err != nil {
		logging.Err.Error(err)
	}

	if *Action == "install" {
		err := s.Install()
		if err != nil {
			panic(err)
		}
		logging.Dbug.Info("服务安装成功")
		return
	}

	if *Action == "sync" {
		err := models.Sync()
		if err != nil {
			panic(err)
		}
		logging.Dbug.Info("同步数据")
		return
	}

	if *Action == "remove" {
		err := s.Uninstall()
		if err != nil {
			panic(err)
		}
		logging.Dbug.Info("服务卸载成功")
		return
	}

	if *Action == "start" {
		err := s.Start()
		if err != nil {
			panic(err)
		}
		logging.Dbug.Info("服务启动成功")
		return
	}

	if *Action == "stop" {
		err := s.Stop()
		if err != nil {
			panic(err)
		}
		logging.Dbug.Info("服务停止成功")
		return
	}

	if *Action == "restart" {
		err := s.Restart()
		if err != nil {
			panic(err)
		}

		logging.Dbug.Info("服务重启成功")
		return
	}

	if *Action == "version" {
		logging.Dbug.Info(fmt.Sprintf("版本号：%s", Version))
		return
	}

	s.Run()

}
