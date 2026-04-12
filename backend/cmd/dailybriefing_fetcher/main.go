package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"ai-blog/backend/internal/bootstrap"
	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/repository"
	"ai-blog/backend/internal/service"
)

func main() {
	date := flag.String("date", "", "briefing date in YYYY-MM-DD")
	limit := flag.Int("limit", 10, "number of briefing items to fetch")
	flag.Parse()

	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := bootstrap.InitDatabase(appConfig)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}

	repo := repository.NewGormDailyBriefingRepository(db)
	briefingService := service.NewDailyBriefingService(repo)

	result, err := briefingService.FetchNow(context.Background(), *date, *limit, "script")
	if err != nil {
		fmt.Fprintf(os.Stderr, "daily briefing fetch failed: %v\n", err)
		if result.Message != "" {
			fmt.Fprintf(os.Stderr, "%s\n", result.Message)
		}
		os.Exit(1)
	}

	fmt.Printf("%s\n", result.Message)
}
