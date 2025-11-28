package main

import (
	"adk-weatherreport/main/llm"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

// WeatherResponse struct maps the JSON response from OpenWeather API
// Contains city name and main temperature details
type WeatherResponse struct {
	CityName string `json:"name"`
	Weather  []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
}

type ErrorResponse struct {
	Code    string `json:"cod"`
	Message string `json:"message"`
}

func getWeatherReport(ctx tool.Context, args getWeatherReportArgs) (getWeatherReportResult, error) {
	resp, err := getTemperature(args.City)
	if err != nil {
		return getWeatherReportResult{Status: "error", Report: err.Error()}, nil
	}

	report := fmt.Sprintf("The weather in %s is %s with a temperature of %f degrees Celsius.", resp.CityName, resp.Weather[0].Description, resp.Main.Temp)
	return getWeatherReportResult{Status: "success", Report: report}, nil
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

func getTemperature(city string) (*WeatherResponse, error) {
	apiKey := os.Getenv("OWM_API_KEY")
	apiURL := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?appid=%s&q=%s&units=metric", apiKey, city)
	res, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if bytes.Index(body, []byte("{\"cod\"")) == 0 {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}

		return nil, fmt.Errorf("failed to get weather report: %s", errorResponse.Message)
	}

	var weatherData WeatherResponse
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &weatherData, nil
}
