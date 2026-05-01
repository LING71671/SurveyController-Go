package engine

import (
	"strings"
	"testing"
)

func TestModeConcurrencyProfile(t *testing.T) {
	tests := []struct {
		mode                 Mode
		maxWorkerConcurrency int
		maxBrowserSessions   int
		httpFirst            bool
	}{
		{mode: ModeHTTP, maxWorkerConcurrency: LightWorkerConcurrencyBaseline, httpFirst: true},
		{mode: ModeHybrid, maxWorkerConcurrency: LightWorkerConcurrencyBaseline, maxBrowserSessions: BrowserWorkerConcurrencyLimit, httpFirst: true},
		{mode: ModeBrowser, maxWorkerConcurrency: BrowserWorkerConcurrencyLimit, maxBrowserSessions: BrowserWorkerConcurrencyLimit},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			profile, err := ModeConcurrencyProfile(tt.mode)
			if err != nil {
				t.Fatalf("ModeConcurrencyProfile() returned error: %v", err)
			}
			if profile.MaxWorkerConcurrency != tt.maxWorkerConcurrency || profile.MaxBrowserSessions != tt.maxBrowserSessions || profile.HTTPFirst != tt.httpFirst {
				t.Fatalf("profile = %+v, want worker %d browser %d httpFirst %v", profile, tt.maxWorkerConcurrency, tt.maxBrowserSessions, tt.httpFirst)
			}
		})
	}
}

func TestValidateConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		mode        Mode
		concurrency int
		wantErr     string
	}{
		{name: "http baseline", mode: ModeHTTP, concurrency: LightWorkerConcurrencyBaseline},
		{name: "hybrid baseline", mode: ModeHybrid, concurrency: LightWorkerConcurrencyBaseline},
		{name: "browser small pool", mode: ModeBrowser, concurrency: BrowserWorkerConcurrencyLimit},
		{name: "browser too high", mode: ModeBrowser, concurrency: BrowserWorkerConcurrencyLimit + 1, wantErr: "browser mode"},
		{name: "zero", mode: ModeHTTP, concurrency: 0, wantErr: "greater than 0"},
		{name: "unsupported", mode: Mode("magic"), concurrency: 1, wantErr: "unsupported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConcurrency(tt.mode, tt.concurrency)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateConcurrency() returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateConcurrency() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
