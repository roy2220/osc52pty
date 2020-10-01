package main

type parser struct {
	captureBeginPattern pattern
	captureEndPattern   pattern
	capturedDataHandler dataHandler
	ignoredDataHandler  dataHandler
	findCaptureBegin    bool
	capturedData        []byte
}

func (p *parser) Init(captureBegin []byte, captureEnd []byte, capturedDataHandler dataHandler, ignoredDataHandler dataHandler) *parser {
	p.captureBeginPattern.Init(captureBegin)
	p.captureEndPattern.Init(captureEnd)
	p.capturedDataHandler = capturedDataHandler
	p.ignoredDataHandler = ignoredDataHandler
	p.findCaptureBegin = true
	return p
}

func (p *parser) FeedData(data []byte) bool {
	var ignoredData []byte

Loop:
	for {
		if p.findCaptureBegin {
			ignoredData = ignoredData[:0]
			i, ok := p.captureBeginPattern.FindStop(data, &ignoredData)

			if len(ignoredData) >= 1 {
				if !p.ignoredDataHandler(ignoredData) {
					return false
				}
			}

			if !ok {
				break Loop
			}

			data = data[i:]
			p.findCaptureBegin = false
		} else {
			i, ok := p.captureEndPattern.FindStop(data, &p.capturedData)

			if !ok {
				break Loop
			}

			if !p.capturedDataHandler(p.capturedData) {
				return false
			}

			p.capturedData = nil
			data = data[i:]
			p.findCaptureBegin = true
		}
	}

	return true
}

type pattern struct {
	raw                 []byte
	kmpNext             []int
	matchedPrefixLength int
}

func (p *pattern) Init(raw []byte) *pattern {
	p.raw = raw
	p.kmpNext = makeKMPNext(raw)
	return p
}

func (p *pattern) FindStop(data []byte, skippedData *[]byte) (int, bool) {
	i, j := 0, p.matchedPrefixLength

	for ; i < len(data) && j < len(p.raw); i, j = i+1, j+1 {
		k := j

		for j >= 0 && data[i] != p.raw[j] {
			j = p.kmpNext[j]
		}

		if j < k {
			if j < 0 {
				*skippedData = append(*skippedData, p.raw[:k]...)
				*skippedData = append(*skippedData, data[i])
			} else {
				*skippedData = append(*skippedData, p.raw[:k-j]...)
			}
		}
	}

	if j < len(p.raw) {
		p.matchedPrefixLength = j
		return 0, false
	}

	p.matchedPrefixLength = 0
	return i, true
}

func makeKMPNext(pattern []byte) []int {
	kmpNext := make([]int, len(pattern))

	for i, j := 0, -1; i < len(pattern); i, j = i+1, j+1 {
		kmpNext[i] = j

		for j >= 0 && pattern[i] != pattern[j] {
			j = kmpNext[j]
		}
	}

	return kmpNext
}

type dataHandler func(data []byte) (ok bool)
