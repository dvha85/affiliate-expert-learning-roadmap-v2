package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestCampaignBoundedFiles(t *testing.T) {
	for _, name := range []string{"manifest", "attempt-001.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "campaign")
			if err := initAdvisorCampaign(path); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(path, name), []byte(strings.Repeat("x", 4096)), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := reserveAdvisorAttempt(path); err == nil {
				t.Fatal("oversized file accepted")
			}
		})
	}
}

func TestCampaignDurableBudget(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if _, err := reserveAdvisorAttempt(p); err == nil {
		t.Fatal("auto initialized")
	}
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 6; i++ {
		n, err := reserveAdvisorAttempt(p)
		if err != nil || n != i {
			t.Fatal(n, err)
		}
	}
	if _, err := reserveAdvisorAttempt(p); err == nil {
		t.Fatal("budget exceeded")
	}
	if err := initAdvisorCampaign(p); err == nil {
		t.Fatal("reset existing campaign")
	}
}
func TestCampaignCorruptionAndStaleLock(t *testing.T) {
	for _, name := range []string{"lock", "attempt-001.json", "manifest"} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "campaign")
			if err := initAdvisorCampaign(p); err != nil {
				t.Fatal(err)
			}
			if name == "lock" {
				if err := os.Mkdir(filepath.Join(p, name), 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(filepath.Join(p, name), []byte("broken"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := reserveAdvisorAttempt(p); err == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
}
func TestCampaignConcurrentReservations(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := map[int]bool{}
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := reserveAdvisorAttempt(p)
			if err == nil {
				mu.Lock()
				defer mu.Unlock()
				if seen[n] {
					t.Error("duplicate")
				}
				seen[n] = true
			}
		}()
	}
	wg.Wait()
	if len(seen) > 6 || len(seen) == 0 {
		t.Fatal(seen)
	}
}

type failedCampaignProvider struct{}

func (failedCampaignProvider) identity() providerIdentity {
	return providerIdentity{"test", "test", "test"}
}
func (failedCampaignProvider) generate(context.Context, advisorContext) ([]byte, error) {
	return nil, errors.New("unknown remote outcome")
}
func TestCampaignFailureConsumesReservation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	n, status, err := runCampaignAttempt(context.Background(), p, failedCampaignProvider{})
	if n != 1 || status != "PROVIDER_ERROR" || err != nil {
		t.Fatal(n, status, err)
	}
	n, status, err = runCampaignAttempt(context.Background(), p, mockAdvisorProvider{})
	if n != 2 || status != "SUPPORTED" || err != nil {
		t.Fatal(n, status, err)
	}
}
