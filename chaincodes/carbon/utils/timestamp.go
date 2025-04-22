package utils

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TimestampRFC3339UtcString(bidTS *timestamppb.Timestamp) string {
	if bidTS == nil {
		return ""
	}
	// RFC3339     = "2006-01-02T15:04:05Z07:00"
	// "AsTime" returns UTC time.
	// We also use environment variable TZ=UTC to make sure
	return bidTS.AsTime().Format(time.RFC3339)
}
