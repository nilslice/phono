package mock

import (
	"sync"
	"time"

	"github.com/dudk/phono"
	"github.com/dudk/phono/pipe"
	"github.com/dudk/phono/pipe/runner"
)

const (
	defaultBufferSize = 512
	defaultSampleRate = 44100

	// PumpMaxInterval is 2 seconds
	PumpMaxInterval = 2000
	// PumpDefaultLimit is 10 messages
	PumpDefaultLimit = 10
)

// Param types
type (
	// Interval between messages in ms
	Interval int
	// Limit messages
	Limit int

	// Counter can be used to check metrics for mocks
	Counter struct {
		m        sync.Mutex
		messages int64
		samples  int64
	}
)

func (c *Counter) advance(buf phono.Buffer) {
	c.m.Lock()
	c.messages++
	c.samples = c.samples + int64(buf.Size())
	c.m.Unlock()
}

func (c *Counter) reset() {
	c.m.Lock()
	defer c.m.Unlock()
	c.messages, c.samples = 0, 0
}

// Count returns message and samples counted
func (c *Counter) Count() (int64, int64) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.messages, c.samples
}

// Pump mocks a pipe.Pump interface
type Pump struct {
	phono.UID
	Counter
	Interval
	Limit
	Value float64
	phono.BufferSize
	phono.NumChannels
	newMessage phono.NewMessageFunc
}

// IntervalParam pushes new interval value for pump
func (p *Pump) IntervalParam(i Interval) phono.Param {
	return phono.Param{
		ID: p.ID(),
		Apply: func() {
			p.Interval = i
		},
	}
}

// LimitParam pushes new limit value for pump
func (p *Pump) LimitParam(l Limit) phono.Param {
	return phono.Param{
		ID: p.ID(),
		Apply: func() {
			p.Limit = l
		},
	}
}

// ValueParam pushes new signal value for pump
func (p *Pump) ValueParam(v float64) phono.Param {
	return phono.Param{
		ID: p.ID(),
		Apply: func() {
			p.Value = v
		},
	}
}

// NumChannelsParam pushes new number of channels for pump
func (p *Pump) NumChannelsParam(nc phono.NumChannels) phono.Param {
	return phono.Param{
		ID: p.ID(),
		Apply: func() {
			p.NumChannels = nc
		},
	}
}

// Pump returns new buffer for pipe
func (p *Pump) Pump() (phono.Buffer, error) {
	if Limit(p.Counter.messages) >= p.Limit {
		return nil, pipe.ErrEOP
	}
	time.Sleep(time.Millisecond * time.Duration(p.Interval))

	buf := phono.Buffer(make([][]float64, p.NumChannels))
	for i := range buf {
		buf[i] = make([]float64, p.BufferSize)
		for j := range buf[i] {
			buf[i][j] = p.Value
		}
	}
	p.Counter.advance(buf)
	return buf, nil
}

// RunPump returns new pump runner
func (p *Pump) RunPump() pipe.PumpRunner {
	return &runner.Pump{
		Pump: p,
		Before: func() error {
			p.reset()
			return nil
		},
	}
}

// Processor mocks a pipe.Processor interface
type Processor struct {
	phono.UID
	Counter
}

// Process implementation for runner
func (p *Processor) Process(buf phono.Buffer) (phono.Buffer, error) {
	p.Counter.advance(buf)
	return buf, nil
}

// RunProcess returns a pipe.ProcessRunner for this processor
func (p *Processor) RunProcess() pipe.ProcessRunner {
	return &runner.Process{
		Processor: p,
		Before: func() error {
			p.reset()
			return nil
		},
	}
}

// Sink mocks up a pipe.Sink interface
// Buffer is not thread-safe, so should not be checked while pipe is running
type Sink struct {
	phono.UID
	Counter
	phono.Buffer
}

// Sink implementation for runner
func (s *Sink) Sink(buf phono.Buffer) error {
	s.Buffer = s.Buffer.Append(buf)
	s.Counter.advance(buf)
	return nil
}

// RunSink returns new sink runner
func (s *Sink) RunSink() pipe.SinkRunner {
	return &runner.Sink{
		Sink: s,
		Before: func() error {
			s.Buffer = nil
			s.reset()
			return nil
		},
	}
}
