package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	owm "github.com/briandowns/openweathermap"
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

	apiKey := os.Getenv("OWM_API_KEY")
	w, err := owm.NewCurrent("K", "EN", apiKey)
	if err != nil {
		log.Fatalln(err)
	}

	err = w.CurrentByName("Phoenix,AZ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(w)

	weatherAgent, err := NewWeatherSentimentAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(weatherAgent),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
