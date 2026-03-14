package service

import (
	"github.com/boyter/pincer/common"
	"github.com/rs/zerolog/log"
	"math"
	"time"
)

// StartBackgroundJobs is the method called after the service is
// setup correctly to trigger all the background processing jobs
// should only ever be called once
func (s *Service) StartBackgroundJobs() {
	s.ServiceMutex.Lock()
	defer s.ServiceMutex.Unlock()

	if !s.BackgroundJobsStarted {
		go s.pruneActivity()
		go s.periodicSaveActivity()
	}

	s.BackgroundJobsStarted = true
}

func (s *Service) periodicSaveActivity() {
	for {
		time.Sleep(1 * time.Hour)
		s.SaveActivity()
	}
}

func (s *Service) pruneActivity() {
	for {
		time.Sleep(5 * time.Minute)
		s.ServiceMutex.Lock()
		var loops int

		for s.TotalActivity > MaxTotalActivity {
			var count int64
			var oldestTime int64 = math.MaxInt64
			var oldestKey = ""

			// loop though the first 100 entries looking for the oldest one and purge em
			// because maps should be unordered this shouldn't be a huge issue
			for k, v := range s.UserActivity {
				if v.LastInteractionUnixTimestamp < oldestTime {
					oldestKey = k
				}
				count++
				if count >= 100 {
					break
				}
			}

			log.Info().Str(common.UniqueCode, "95735953").Str("key", oldestKey).Int64("TotalActivity", s.TotalActivity).Msg("removing key")
			s.TotalActivity = s.TotalActivity - int64(len(s.UserActivity[oldestKey].Activity))
			delete(s.UserActivity, oldestKey)

			loops++
			if loops >= 10 {
				break
			}
		}

		s.ServiceMutex.Unlock()
	}
}
