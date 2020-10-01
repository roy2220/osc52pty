package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	type FeedDataCall struct {
		Data         string
		OK           bool
		CapturedData string
		IgnoredData  string
	}
	for _, tt := range []struct {
		CaptureBegin  string
		CaptureEnd    string
		FeedDataCalls []FeedDataCall
	}{
		{
			CaptureBegin: "<<<",
			CaptureEnd:   ">>>",
			FeedDataCalls: []FeedDataCall{
				{
					Data:         "aaa<<<\x00123>>>ccc",
					CapturedData: "",
					IgnoredData:  "aaa",
					OK:           false,
				},
			},
		},
		{
			CaptureBegin: "<<<",
			CaptureEnd:   ">>>",
			FeedDataCalls: []FeedDataCall{
				{
					Data:         "aaa<<<123>>>ccc\000",
					CapturedData: "123",
					IgnoredData:  "aaa",
					OK:           false,
				},
			},
		},
		{
			CaptureBegin: "<++",
			CaptureEnd:   "++>",
			FeedDataCalls: []FeedDataCall{
				{
					Data:         "xxx<++HAHA++>yy<++YOYO++>z",
					CapturedData: "HAHAYOYO",
					IgnoredData:  "xxxyyz",
					OK:           true,
				},
			},
		},
		{
			CaptureBegin: "<--",
			CaptureEnd:   "-->",
			FeedDataCalls: []FeedDataCall{
				{
					Data:         "aaa",
					IgnoredData:  "aaa",
					CapturedData: "",
					OK:           true,
				},
				{
					Data:         "<-",
					IgnoredData:  "aaa",
					CapturedData: "",
					OK:           true,
				},
				{
					Data:         "-hello<--world-",
					IgnoredData:  "aaa",
					CapturedData: "",
					OK:           true,
				},
				{
					Data:         "->bbb",
					CapturedData: "hello<--world",
					IgnoredData:  "aaabbb",
					OK:           true,
				},
			},
		},
	} {
		t.Run(tt.CaptureBegin+"..."+tt.CaptureEnd, func(t *testing.T) {
			var capturedData []byte
			capturedDataHandler := func(data []byte) bool {
				for i := range data {
					if data[i] == '\x00' {
						return false
					}
				}
				capturedData = append(capturedData, data...)
				return true
			}
			var ignoredData []byte
			ignoredDataHandler := func(data []byte) bool {
				for i := range data {
					if data[i] == '\x00' {
						return false
					}
				}
				ignoredData = append(ignoredData, data...)
				return true
			}
			p := new(parser).Init([]byte(tt.CaptureBegin), []byte(tt.CaptureEnd), capturedDataHandler, ignoredDataHandler)
			for _, fdc := range tt.FeedDataCalls {
				ok := p.FeedData([]byte(fdc.Data))
				assert.Equal(t, fdc.OK, ok)
				assert.Equal(t, fdc.CapturedData, string(capturedData))
				assert.Equal(t, fdc.IgnoredData, string(ignoredData))
			}
		})
	}
}

func TestPattern(t *testing.T) {
	type FindStopCall struct {
		Data        string
		SkippedData string
		I           int
		OK          bool
	}
	for _, tt := range []struct {
		Pattern       string
		FindStopCalls []FindStopCall
	}{
		{
			Pattern: "1212A1212B",
			FindStopCalls: []FindStopCall{
				{
					Data:        "1212",
					I:           0,
					OK:          false,
					SkippedData: "",
				},
				{
					Data:        "12A",
					I:           0,
					OK:          false,
					SkippedData: "12",
				},
				{
					Data:        "121",
					I:           0,
					OK:          false,
					SkippedData: "12",
				},
				{
					Data:        "2A",
					I:           0,
					OK:          false,
					SkippedData: "121212A",
				},
				{
					Data:        "1212B",
					I:           5,
					OK:          true,
					SkippedData: "121212A",
				},
			},
		},
		{
			Pattern: "123123abc",
			FindStopCalls: []FindStopCall{
				{
					Data:        "bbb---",
					I:           0,
					OK:          false,
					SkippedData: "bbb---",
				},
				{
					Data:        "123",
					I:           0,
					OK:          false,
					SkippedData: "bbb---",
				},
				{
					Data:        "123",
					I:           0,
					OK:          false,
					SkippedData: "bbb---",
				},
				{
					Data:        "abc",
					I:           3,
					OK:          true,
					SkippedData: "bbb---",
				},
				{
					Data:        "+++eee",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee",
				},
				{
					Data:        "123",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee",
				},
				{
					Data:        "123",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee",
				},
				{
					Data:        "123",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee123",
				},
				{
					Data:        "ab",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee123",
				},
				{
					Data:        "c",
					I:           1,
					OK:          true,
					SkippedData: "bbb---+++eee123",
				},
				{
					Data:        "de",
					I:           0,
					OK:          false,
					SkippedData: "bbb---+++eee123de",
				},
			},
		},
	} {
		t.Run(tt.Pattern, func(t *testing.T) {
			p := new(pattern).Init([]byte(tt.Pattern))
			var skippedData []byte
			for _, fsc := range tt.FindStopCalls {
				i, ok := p.FindStop([]byte(fsc.Data), &skippedData)
				assert.Equal(t, fsc.I, i)
				assert.Equal(t, fsc.OK, ok)
				assert.Equal(t, fsc.SkippedData, string(skippedData))
			}
		})
	}
}
