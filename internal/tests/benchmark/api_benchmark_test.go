package benchmark

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const addr = "http://localhost:8080/api/shorten"

func readDomain(b *testing.B) []string {
	file, err := os.Open("domain.txt")
	if err != nil {
		b.Error(err)
	}

	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func Benchmark_Create(b *testing.B) {
	domains := readDomain(b)

	wg, maxRequest := sync.WaitGroup{}, 100
	count := atomic.Int64{}
	ch, stop := make(chan string, maxRequest), make(chan int)

	tr := &http.Transport{
		MaxIdleConns:    200,
		MaxConnsPerHost: 200,
		IdleConnTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
	}

	wg.Add(maxRequest)
	for range maxRequest {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				case url := <-ch:
					func() {
						resp, err := client.Post(addr, "application/json", strings.NewReader(fmt.Sprintf(`{"long_url": "%s"}`, url)))
						if err != nil {
							b.Error(err)
							return
						}
						defer resp.Body.Close()

						body, err := io.ReadAll(resp.Body)
						if err != nil {
							log.Fatalln(err)
						}
						if resp.StatusCode != http.StatusOK {
							b.Errorf("Error: %d, resp body: %s", resp.StatusCode, body)
							return
						}

						count.Add(1)
					}()
				}
			}
		}()
	}

	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		for range time.NewTicker(time.Second).C {
			b.Log("send requests:", count.Load())
		}
	}()
	start := time.Now()
	for i := range 10 {
		s := sha256.Sum256([]byte(strconv.Itoa(i)))
		for j := range len(domains) {
			ch <- fmt.Sprintf("https://%s/test/api/image/1234567890/%x", domains[j], s)
		}
	}

	for len(ch) > 0 {
		time.Sleep(100 * time.Millisecond)
	}

	close(stop)
	wg.Wait()

	b.Log("success requests: ", count.Load(), "costs", time.Since(start).String())
}
