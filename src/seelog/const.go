package seelog

import (
	"github.com/cihub/seelog"
	"github.com/go-solana-parse/src/config"
)

var (
	appLogger      seelog.LoggerInterface
	cashbackLogger seelog.LoggerInterface
)

func LoadLoggerConfig() {
	appLogger = loadAppLogger()
	cashbackLogger = loadCashBackLogger()
}

func loadAppLogger() seelog.LoggerInterface {
	seelog.RegisterCustomFormatter("ServiceName", createAppNameFormatter)
	var logger seelog.LoggerInterface
	var err error
	if config.SvcConfig.Env == "test" {
		logger, err = seelog.LoggerFromConfigAsString(AppSeelogTestConfigForDate)
	} else {
		logger, err = seelog.LoggerFromConfigAsString(AppSeelogProConfigForDate)
	}
	if err != nil {
		panic(err)
	}
	return logger
}

func loadCashBackLogger() seelog.LoggerInterface {
	seelog.RegisterCustomFormatter("ServiceName", createAppNameFormatter)
	var logger seelog.LoggerInterface
	var err error
	if config.SvcConfig.Env == "test" {
		logger, err = seelog.LoggerFromConfigAsString(CashbackSeelogTestConfigForDate)
	} else {
		logger, err = seelog.LoggerFromConfigAsString(CashbackSeelogProConfigForDate)
	}
	if err != nil {
		panic(err)
	}
	return logger
}

const AppSeelogTestConfigForSize string = `
<seelog minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
			 <rollingfile formatid="fmt_info" type="size" filename="/data/logs/smartx/app/app.log" 
maxsize="33554432" maxrolls="64"/>
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`

const CashbackSeelogTestConfigForSize string = `
<seelog type="asyncloop" minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
			 <rollingfile formatid="fmt_info" type="size" filename="/data/logs/smartx/cashback/app.log" 
maxsize="33554432" maxrolls="64"/>
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`

const AppSeelogTestConfigForDate string = `
<seelog minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
			 <rollingfile formatid="fmt_info" type="date" filename="/data/logs/smartx/app/app.log" datepattern="2006-01-02" maxrolls="14"/>
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`

const CashbackSeelogTestConfigForDate string = `
<seelog minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
			 <rollingfile formatid="fmt_info" type="date" filename="/data/logs/smartx/cashback/app.log" datepattern="2006-01-02" maxrolls="14"/>
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`

const AppSeelogProConfigForDate string = `
<seelog minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
		     <console />
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`

const CashbackSeelogProConfigForDate string = `
<seelog minlevel="trace">
	<outputs formatid="fmt_info">
         <filter levels="trace,debug,info,warn,error,critical">
		 	 <console />
         </filter>
	</outputs>
	<formats>
		<format id="fmt_info" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
		<format id="fmt_err" format="%Date(2006-01-02 15:04:05.999) %ServiceName %Msg%n" />
	</formats>
</seelog>`
