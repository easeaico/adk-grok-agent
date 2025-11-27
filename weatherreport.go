package main

import (
	"adk-weatherreport/main/llm"
	"context"
	"log"
	"os"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

type getWeatherReportArgs struct {
	City string `json:"city" jsonschema:"The city for which to get the weather report."`
}

type getWeatherReportResult struct {
	Status string `json:"status"`
	Report string `json:"report,omitempty"`
}

func getWeatherReport(ctx tool.Context, args getWeatherReportArgs) (getWeatherReportResult, error) {
	if strings.ToLower(args.City) == "london" {
		return getWeatherReportResult{Status: "success", Report: "The current weather in London is cloudy with a temperature of 18 degrees Celsius and a chance of rain."}, nil
	}
	if strings.ToLower(args.City) == "paris" {
		return getWeatherReportResult{Status: "success", Report: "The weather in Paris is sunny with a temperature of 25 degrees Celsius."}, nil
	}
	return getWeatherReportResult{Status: "error"}, nil
}

type analyzeSentimentArgs struct {
	Text string `json:"text" jsonschema:"The text to analyze for sentiment."`
}

type analyzeSentimentResult struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
}

func analyzeSentiment(ctx tool.Context, args analyzeSentimentArgs) (analyzeSentimentResult, error) {
	if strings.Contains(strings.ToLower(args.Text), "good") || strings.Contains(strings.ToLower(args.Text), "sunny") {
		return analyzeSentimentResult{Sentiment: "positive", Confidence: 0.8}, nil
	}
	if strings.Contains(strings.ToLower(args.Text), "rain") || strings.Contains(strings.ToLower(args.Text), "bad") {
		return analyzeSentimentResult{Sentiment: "negative", Confidence: 0.7}, nil
	}
	return analyzeSentimentResult{Sentiment: "neutral", Confidence: 0.6}, nil
}

func NewWeatherSentimentAgent(ctx context.Context) (agent.Agent, error) {
	model, err := llm.NewGrokModel(ctx, "grok-4-1-fast", &genai.ClientConfig{
		APIKey: os.Getenv("XAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
		return nil, err
	}

	weatherTool, err := functiontool.New(
		functiontool.Config{
			Name:        "get_weather_report",
			Description: "Retrieves the current weather report for a specified city.",
		},
		getWeatherReport,
	)
	if err != nil {
		log.Fatal(err)
	}

	sentimentTool, err := functiontool.New(
		functiontool.Config{
			Name:        "analyze_sentiment",
			Description: "Analyzes the sentiment of the given text.",
		},
		analyzeSentiment,
	)
	if err != nil {
		log.Fatal(err)
	}

	weatherSentimentAgent, err := llmagent.New(llmagent.Config{
		Name:        "weather_sentiment_agent",
		Model:       model,
		Instruction: "You are a helpful assistant that provides weather information and analyzes the sentiment of user feedback. **If the user asks about the weather in a specific city, use the 'get_weather_report' tool to retrieve the weather details.** **If the 'get_weather_report' tool returns a 'success' status, provide the weather report to the user.** **If the 'get_weather_report' tool returns an 'error' status, inform the user that the weather information for the specified city is not available and ask if they have another city in mind.** **After providing a weather report, if the user gives feedback on the weather (e.g., 'That's good' or 'I don't like rain'), use the 'analyze_sentiment' tool to understand their sentiment.** Then, briefly acknowledge their sentiment. You can handle these tasks sequentially if needed.",
		Tools:       []tool.Tool{weatherTool, sentimentTool},
	})
	if err != nil {
		log.Fatal(err)
	}

	return weatherSentimentAgent, nil
}
