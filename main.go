package main

import (
	"adk-weatherreport/main/agents"
	"context"
	"log"
	"os"
	"syscall"

	"github.com/maxence-charriere/go-app/v10/pkg/cli"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
)

func main() {
	// 创建带有信号处理的上下文
	ctx, cancel := cli.ContextWithSignals(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// 创建天气报告智能体
	weatherReportAgent, err := agents.NewWeatherReportAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(weatherReportAgent),
	}

	// 创建Launcher,运行测试环境
	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
