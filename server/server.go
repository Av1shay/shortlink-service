package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/url"
	"shortlink-service/shortlink"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	shortnerClient ShortnerClient
}

type ShortnerClient interface {
	GenerateShortLink(ctx context.Context, data *shortlink.Input) (string, error)
	GetLongURL(ctx context.Context, key string, t time.Time, kt shortlink.KeyType, incVisits bool) (string, error)
	GelAllShortLinks(ctx context.Context) ([]*shortlink.Item, error)
	DeleteShortLink(ctx context.Context, key string) error
}

func New(ctx context.Context, shortnerClient ShortnerClient, router chi.Router) (*Server, error) {
	s := Server{shortnerClient: shortnerClient}
	router.Post("/s/generate", s.ShortlinkGenerateHandler)
	router.Get("/{shortlink}", s.ShortlinkRedirectHandler)
	router.Get("/u/{uuid}", s.ShortlinkRedirectUuidHandler)
	router.Get("/cron/checkRedirects", s.CheckRedirectsHandler)
	return &s, nil
}

func (s *Server) ShortlinkGenerateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in shortlink.Input

	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		fmt.Printf("error decode shortlink: %v\n", err)
		http.Error(w, "failed to decode body", http.StatusBadRequest)
		return
	}

	err = validateURLs(&in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortLink, err := s.shortnerClient.GenerateShortLink(ctx, &in)
	if err != nil {
		fmt.Printf("error generating url: %v\n", err)
		http.Error(w, "error generating url", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprint(shortLink)))
}

func (s *Server) ShortlinkRedirectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqTime := time.Now()
	key := chi.URLParam(r, "shortlink")

	redirectUrl, err := s.shortnerClient.GetLongURL(ctx, key, reqTime, shortlink.KeyTypeStandard, true)
	if err != nil {
		fmt.Printf("error getting url by key: %v\n", err)
		http.Error(w, "error getting url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusFound)
	return
}

func (s *Server) ShortlinkRedirectUuidHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqTime := time.Now()
	key := chi.URLParam(r, "uuid")

	redirectUrl, err := s.shortnerClient.GetLongURL(ctx, key, reqTime, shortlink.KeyTypeUuid, true)
	if err != nil {
		fmt.Printf("error getting url by uuid: %v\n", err)
		http.Error(w, "error getting url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusFound)
	return
}

func (s *Server) CheckRedirectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := s.CheckRedirects(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprint("ok")))
}

func (s *Server) CheckRedirects(ctx context.Context) error {
	defer elapsed("CheckRedirects()")()
	items, err := s.shortnerClient.GelAllShortLinks(ctx)
	if err != nil {
		return err
	}

	numJobs := len(items)
	itemJobs := make(chan *shortlink.Item, numJobs)
	workers := 1000
	hc := &http.Client{Timeout: 10 * time.Second}
	wg := sync.WaitGroup{}
	wg.Add(workers)
	var doneCount uint64

	go backgroundTask(ctx, numJobs+1, &doneCount)

	for i := 0; i < workers; i++ {
		go func(ctxI context.Context, wgI *sync.WaitGroup) {
			defer wgI.Done()

			for item := range itemJobs {
				numSubJobs := len(item.Redirects)
				urlJobs := make(chan string, numSubJobs)
				subWorkers := 2
				subWg := sync.WaitGroup{}
				subWg.Add(subWorkers)

				for j := 0; j < subWorkers; j++ {
					go s.urlCheckWorker(ctxI, urlJobs, hc, item.Key, &subWg)
				}

				for _, r := range item.Redirects {
					urlJobs <- r.URL
				}
				close(urlJobs)

				subWg.Wait()
				atomic.AddUint64(&doneCount, 1)
			}
		}(ctx, &wg)
	}

	for _, item := range items {
		itemJobs <- item
	}
	close(itemJobs)

	wg.Wait()

	return nil
}

func (s *Server) urlCheckWorker(ctx context.Context, urlJobs <-chan string, hc *http.Client, itemKey string, wg *sync.WaitGroup) {
	defer wg.Done()
	for u := range urlJobs {
		resp, err := hc.Get(u)
		if err != nil {
			fmt.Printf("error getting URL: %v\n", err)
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("invalid status code %d for url %s with item key %s, deleting item\n", resp.StatusCode, u, itemKey)
			err := s.shortnerClient.DeleteShortLink(ctx, itemKey)
			if err != nil {
				fmt.Printf("error deleting item: %v\n", err)
			}
		}
	}
}

func backgroundTask(ctx context.Context, total int, done *uint64) {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case _ = <-ticker.C:
				fmt.Printf("progress (%d,%d)\n", atomic.LoadUint64(done), total)
			}
		}
	}()
}

func validateURLs(in *shortlink.Input) error {
	for i, r := range in.Redirects {
		if r.URL == "" {
			return fmt.Errorf("url not provided at index %d", i)
		}
		_, err := url.ParseRequestURI(r.URL)
		if err != nil {
			return err
		}
	}
	return nil
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
