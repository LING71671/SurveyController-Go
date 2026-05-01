package engine

import "fmt"

const (
	LightWorkerConcurrencyBaseline = 1000
	BrowserWorkerConcurrencyLimit  = 16
)

type ConcurrencyProfile struct {
	Mode                 Mode
	MaxWorkerConcurrency int
	MaxBrowserSessions   int
	HTTPFirst            bool
}

func ModeConcurrencyProfile(mode Mode) (ConcurrencyProfile, error) {
	parsed, err := ParseMode(mode.String())
	if err != nil {
		return ConcurrencyProfile{}, err
	}
	switch parsed {
	case ModeHTTP:
		return ConcurrencyProfile{
			Mode:                 ModeHTTP,
			MaxWorkerConcurrency: LightWorkerConcurrencyBaseline,
			HTTPFirst:            true,
		}, nil
	case ModeHybrid:
		return ConcurrencyProfile{
			Mode:                 ModeHybrid,
			MaxWorkerConcurrency: LightWorkerConcurrencyBaseline,
			MaxBrowserSessions:   BrowserWorkerConcurrencyLimit,
			HTTPFirst:            true,
		}, nil
	case ModeBrowser:
		return ConcurrencyProfile{
			Mode:                 ModeBrowser,
			MaxWorkerConcurrency: BrowserWorkerConcurrencyLimit,
			MaxBrowserSessions:   BrowserWorkerConcurrencyLimit,
		}, nil
	default:
		return ConcurrencyProfile{}, fmt.Errorf("unsupported engine mode %q", mode)
	}
}

func ValidateConcurrency(mode Mode, concurrency int) error {
	if concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}
	profile, err := ModeConcurrencyProfile(mode)
	if err != nil {
		return err
	}
	if concurrency > profile.MaxWorkerConcurrency {
		return fmt.Errorf("%s mode concurrency must not exceed %d", profile.Mode, profile.MaxWorkerConcurrency)
	}
	return nil
}
