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
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/service"
	_ "net/http/pprof"
)

var Version string

type program struct {
	httpServer *http.Server
}

func (p *program) Start(s service.Service) error {
	defer println("********* START *********")
	go p.run()
	return nil
}

func (p *program) run() {
	sync()
}

func sync() {
	v := utils.Config.Timetype
	t := utils.Config.Timeduration
	var chSy chan int
	var tickerSync *time.Ticker
	switch v {
	case "h":
		tickerSync = time.NewTicker(time.Hour * time.Duration(t))
	case "m":
		tickerSync = time.NewTicker(time.Minute * time.Duration(t))
	case "s":
		tickerSync = time.NewTicker(time.Second * time.Duration(t))
	default:
		tickerSync = time.NewTicker(time.Hour * time.Duration(t))
	}
	defer tickerSync.Stop()

	mysql := models.GetMysql()
	defer models.CloseMysql(mysql)

	loggerU := logging.GetMyLogger("user_type")
	loggerR := logging.GetMyLogger("remote")
	loggerL := logging.GetMyLogger("loc")
	var userTypes []models.UserType
	var requestUserTypesJson []byte
	var remoteDevs []models.RemoteDev
	var requestRemoteDevsJson []byte
	var locs []models.Loc
	var requestLocsJson []byte
	go func() {
		for range tickerSync.C {
			if err := utils.GetToken(); err != nil {
				fmt.Println(err)
			}
			if utils.GetAppInfoCache() == nil {
				fmt.Println("app info nil")
			}
			models.RemoteSync(remoteDevs, requestRemoteDevsJson, mysql, loggerR)
			models.LocSync(locs, requestLocsJson, mysql, loggerL)
			models.UserTypeSync(userTypes, requestUserTypesJson, mysql, loggerU)
		}
		chSy <- 1
	}()
	<-chSy
}

func (p *program) Stop(s service.Service) error {
	defer log.Println("********** STOP **********")
	return nil
}

var Action = flag.String("action", "", "程序操作指令")

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	logger := logging.GetMyLogger("common")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [command]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  -action= <install remove start stop restart version remote_sync loc_sync user_type_sync cache_clear>\n")
		fmt.Fprintf(os.Stderr, "    程序操作指令 \n")
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
		logger.Error(err)
	}

	if err != nil {
		logger.Error(err)
	}

	if *Action == "install" {
		err = s.Install()
		if err != nil {
			logger.Error("服务安装错误：", err)
		}
		err = s.Start()
		if err != nil {
			logger.Error("服务启动错误", err)
		}
		logger.Info("服务安装并启动")
		return
	}

	if *Action == "remote_sync" {
		if err = utils.GetToken(); err != nil {
			logger.Error("get token err ", err)
		}
		if utils.GetAppInfoCache() == nil {
			logger.Error("app info is empty")
		}
		mysql := models.GetMysql()
		defer models.CloseMysql(mysql)
		loggerR := logging.GetMyLogger("remote")
		var remoteDevs []models.RemoteDev
		var requestRemoteDevsJson []byte
		models.RemoteSync(remoteDevs, requestRemoteDevsJson, mysql, loggerR)
		logger.Info("探视数据同步")

		return
	}

	if *Action == "loc_sync" {
		if err = utils.GetToken(); err != nil {
			logger.Error("get token err ", err)
		}
		if utils.GetAppInfoCache() == nil {
			logger.Error("app info is empty")
		}
		mysql := models.GetMysql()
		defer models.CloseMysql(mysql)
		loggerL := logging.GetMyLogger("loc")
		var locs []models.Loc
		var requestLocsJson []byte
		models.LocSync(locs, requestLocsJson, mysql, loggerL)

		logger.Info("科室数据同步")

		return
	}

	if *Action == "user_type_sync" {
		if err = utils.GetToken(); err != nil {
			logger.Error("get token err ", err)
		}
		if utils.GetAppInfoCache() == nil {
			logger.Error("app info is empty")
		}
		mysql := models.GetMysql()
		defer models.CloseMysql(mysql)
		loggerU := logging.GetMyLogger("user_type")
		var userTypes []models.UserType
		var requestUserTypesJson []byte
		models.UserTypeSync(userTypes, requestUserTypesJson, mysql, loggerU)
		logger.Info("职称数据同步")

		return
	}

	if *Action == "cache_clear" {
		utils.GetCache().Delete(fmt.Sprintf("XToken_%s", utils.Config.Appid))
		utils.GetCache().Delete(fmt.Sprintf("APPINFO_%s", utils.Config.Appid))
		utils.GetCache().DeleteExpired()
		logger.Info("清除缓存")
		return
	}

	if *Action == "remove" {
		status, _ := s.Status()
		if status == service.StatusRunning {
			err = s.Stop()
			if err != nil {
				logger.Error("服务停止错误：", err)
			}
		}

		err = s.Uninstall()
		if err != nil {
			logger.Error("服务卸载错误：", err)
		}
		logger.Info("服务卸载成功")
		return
	}

	if *Action == "start" {
		err = s.Start()
		if err != nil {
			logger.Error("服务启动错误：", err)
		}
		logger.Info("服务启动成功")
		return
	}

	//if *Action == "auto_migrate" {
	//	err = models.GetSqlite().AutoMigrate(&models.RemoteDev{}, &models.Loc{}, &models.UserType{})
	//	if err != nil {
	//		fmt.Println(fmt.Sprintf("database model init error:%+v", err))
	//	}
	//}

	if *Action == "stop" {
		err = s.Stop()
		if err != nil {
			logger.Error("服务停止错误：", err)
		}
		logger.Info("服务停止成功")
		return
	}

	if *Action == "restart" {
		err = s.Restart()
		if err != nil {
			logger.Error("服务重启错误：", err)
		}

		logger.Info("服务重启成功")
		return
	}

	if *Action == "version" {
		logger.Info(fmt.Sprintf("版本号：%s", Version))
		return
	}

	s.Run()

}
