package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Define Prometheus metrics
	rankMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "website_data",
			Help: "rank of market value",
		},
		[]string{"id", "name"},
	)
	portionMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "TWSE_portion",
			Help: "portion of tatal market value",
		},
		[]string{"id", "name"},
	)
	errorMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "error_total",
			Help: "number of errors when parsing website",
		},
		[]string{"error"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(rankMetrics)
	prometheus.MustRegister(portionMetrics)
}

func main() {
	// Parse website and extract data
	go func() {
		for {
			fmt.Println("Parsing website...")
			if err := parseWebsite(); err != nil {
				log.Println("Error parsing website:", err)
			}
			// Sleep for some time before parsing again
			time.Sleep(6 * time.Hour)
		}
	}()

	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseWebsite() error {
	// Load the website
	doc, err := goquery.NewDocument("https://www.taifex.com.tw/cht/9/futuresQADetail")
	if err != nil {
		return err
	}

	// Extract data from the website
	doc.Find(".table_c tr").Each(func(i int, s *goquery.Selection) {
		rankStr := strings.TrimSpace(s.Find("td").Eq(0).Text())
		if rankStr == "" {
			return
		}
		rank, err := strconv.ParseInt(rankStr, 10, 32)
		if err != nil {
			errorMetrics.WithLabelValues("rank").Inc()
			log.Println("Error parsing rank:", err)
			return
		}
		stockSymbol := strings.TrimSpace(s.Find("td").Eq(1).Text())
		// Extract stock name and market value weight
		stockName := strings.TrimSpace(s.Find("td").Eq(2).Text())
		marketPortionStr := strings.TrimSpace(s.Find("td").Eq(3).Text())
		marketPortion, err := strconv.ParseFloat(strings.TrimSuffix(marketPortionStr, "%"), 64)
		if err != nil {
			log.Println("Error parsing market value:", err)
			return
		}

		// Update Prometheus metrics
		rankMetrics.WithLabelValues(stockSymbol, stockName).Set(float64(rank))
		portionMetrics.WithLabelValues(stockSymbol, stockName).Set(marketPortion)
	})

	return nil
}
