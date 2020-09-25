package main

type parser struct {
	CaptureBegin        []byte
	CaptureEnd          []byte
	CapturedDataHandler func([]byte) bool
	IgnoredDataHandler  func([]byte) bool

	findCaptureBegin    bool
	captureBeginPattern pattern
	captureEndPattern   pattern
	capturedData        []byte
}

func (p *parser) Init() *parser {
	p.findCaptureBegin = true
	p.captureBeginPattern = pattern{Raw: p.CaptureBegin}
	p.captureBeginPattern.Init()
	p.captureEndPattern = pattern{Raw: p.CaptureEnd}
	p.captureEndPattern.Init()
	return p
}

func (p *parser) FeedData(data []byte) bool {
	ignoredData := []byte(nil)

Loop:
	for {
		if p.findCaptureBegin {
			ignoredData = ignoredData[:0]
			i, ok := p.captureBeginPattern.FindStop(data, &ignoredData)

			if len(ignoredData) >= 1 {
				if !p.IgnoredDataHandler(ignoredData) {
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

			if !p.CapturedDataHandler(p.capturedData) {
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
	Raw []byte

	kmpNext             []int
	matchedPrefixLength int
}

func (p *pattern) Init() *pattern {
	p.kmpNext = makeKMPNext(p.Raw)
	return p
}

func (p *pattern) FindStop(data []byte, skippedData *[]byte) (int, bool) {
	i, j := 0, p.matchedPrefixLength

	for ; i < len(data) && j < len(p.Raw); i, j = i+1, j+1 {
		k := j

		for j >= 0 && data[i] != p.Raw[j] {
			j = p.kmpNext[j]
		}

		if j < k {
			if j < 0 {
				*skippedData = append(*skippedData, p.Raw[:k]...)
				*skippedData = append(*skippedData, data[i])
			} else {
				*skippedData = append(*skippedData, p.Raw[:k-j]...)
			}
		}
	}

	if j < len(p.Raw) {
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
