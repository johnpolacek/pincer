package common

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"testing"
)

func TestAbs(t *testing.T) {
	if Abs(1) != 1 {
		t.Error("1")
	}

	if Abs(-1) != 1 {
		t.Error("1")
	}

	if Abs(-100) != 100 {
		t.Error("100")
	}
}

func TestLogging(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	log.Info().Msg("something")
	log.Error().Msg("something error")
	log.Error().Err(errors.New("inner error")).Msg("something went wrong here")
}

func TestSieveOfEratosthenes(t *testing.T) {
	fmt.Println(SieveOfEratosthenes(1_000_000))
}
