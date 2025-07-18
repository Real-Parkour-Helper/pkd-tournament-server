package util

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
)

func ParseTime(pb string) (uint64, error) {
	pb = strings.TrimSpace(pb)
	minutes, rest, found := strings.Cut(pb, ":")
	if !found {
		err := errors.New("invalid time format: failed to parse minutes")
		slog.Warn(err.Error())
		return 0, err
	}

	seconds, rest, found := strings.Cut(rest, ".")
	if !found {
		err := errors.New("invalid time format: failed to parse seconds")
		slog.Warn(err.Error())
		return 0, err
	}

	minutesInt, err := strconv.ParseInt(minutes, 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	secondsInt, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	if len(rest) == 0 {
		err := errors.New("invalid time format: failed to parse milliseconds")
		slog.Warn(err.Error())
		return 0, err
	}

	millisecondsInt, err := strconv.ParseInt(rest[0:1], 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	return uint64(minutesInt*60*1000 + secondsInt*1000 + millisecondsInt), nil
}
