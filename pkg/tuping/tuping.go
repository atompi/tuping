package tuping

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"gitee.com/autom-studio/tuping/pkg/options"
	"github.com/go-ping/ping"
)

func ICMPing(pingOptions *options.PingOptions) error {
	pinger, err := ping.NewPinger(pingOptions.Host)
	if err != nil {
		return err
	}

	pinger.Count = pingOptions.Count
	pinger.Size = pingOptions.Size
	pinger.Interval = pingOptions.Interval * time.Millisecond
	pinger.TTL = pingOptions.TTL

	// Listen for Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pinger.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Ttl, pkt.Rtt)
	}

	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	fmt.Printf("PING %s (%s) %d\n", pinger.Addr(), pinger.IPAddr(), pinger.Size)
	err = pinger.Run()
	return err
}

type PingResult struct {
	Received int
	Dropped  int
}

type L4Pinger struct {
	L4PingOptions *options.PingOptions
	Result        *PingResult
	interrupted   bool
	outputted     bool
}

func (p *L4Pinger) OutPutMsg(msg string) {
	fmt.Println(msg)
}

func (p *L4Pinger) OutputOnce(msg string) {
	if !p.outputted {
		fmt.Println(msg)
		p.outputted = true
	}
}

func (p *L4Pinger) Wrapper(conn net.Conn, cnt int, err error) {
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		p.Result.Dropped++
		p.OutPutMsg(fmt.Sprintf("%v for seq=%d", err, cnt))
	}
	cnt++
	if p.L4PingOptions.Count <= 0 {
		p.interrupted = false
	} else if cnt >= p.L4PingOptions.Count {
		p.interrupted = true
	}
	time.Sleep(p.L4PingOptions.Interval * time.Millisecond)
}

func (p *L4Pinger) Ping() error {
	cnt := 0
	p.interrupted = false

	var payload []byte
	for i := 0; i < p.L4PingOptions.Size; i++ {
		payload = append(payload, 0x0a)
	}

	if p.L4PingOptions.DNS != "" {
		dns := p.L4PingOptions.DNS
		if !strings.Contains(dns, ":") {
			dns += ":53"
		}

		_, err := net.DialTimeout(p.L4PingOptions.Protocol, dns, time.Duration(p.L4PingOptions.TTL)*time.Millisecond)
		if err != nil {
			return err
		}
		net.DefaultResolver = &net.Resolver{
			PreferGo: false,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial(network, dns)
			},
		}
	}

	remoteAddr := fmt.Sprintf("%s:%d", p.L4PingOptions.Host, p.L4PingOptions.Port)

	for !p.interrupted {
		start1 := time.Now()
		conn, err := net.DialTimeout(
			p.L4PingOptions.Protocol,
			remoteAddr,
			time.Duration(p.L4PingOptions.TTL)*time.Millisecond,
		)
		stop1 := time.Now()
		if err != nil {
			p.Wrapper(nil, cnt, err)
			continue
		}

		realAddr := conn.RemoteAddr().String()
		p.OutputOnce(fmt.Sprintf("PING %s (%s) %d", p.L4PingOptions.Host, realAddr, p.L4PingOptions.Size))

		conn.SetDeadline(time.Now().Add(time.Duration(p.L4PingOptions.TTL) * time.Millisecond))

		start2 := time.Now()
		n, err := conn.Write(payload)
		stop2 := time.Now()
		if err != nil {
			p.Wrapper(conn, cnt, err)
			continue
		}
		if n != p.L4PingOptions.Size {
			err = fmt.Errorf("partial payload written (size=%d)", n)
			p.Wrapper(conn, cnt, err)
			continue
		}

		rtt := stop1.Sub(start1) + stop2.Sub(start2)
		payloadSize := p.L4PingOptions.Size
		transfDirection := "to"

		if p.L4PingOptions.Wait {
			buf := make([]byte, 1024)
			start3 := time.Now()
			packetSize, err := conn.Read(buf)
			stop3 := time.Now()
			if err != nil {
				p.Wrapper(conn, cnt, err)
				continue
			}
			if packetSize <= 0 {
				err = fmt.Errorf("no packet received")
				p.Wrapper(conn, cnt, err)
				continue
			}
			rtt += stop3.Sub(start3)
			payloadSize = packetSize
			transfDirection = "from"
		}

		p.Result.Received++
		p.OutPutMsg(fmt.Sprintf("%d bytes %s %s %s_seq=%d ttl=%d time=%.3f ms",
			payloadSize, transfDirection, realAddr, p.L4PingOptions.Protocol, cnt, p.L4PingOptions.TTL, float64(rtt)/float64(time.Millisecond)))
		p.Wrapper(conn, cnt, nil)
	}

	return nil
}

func (p *L4Pinger) OnFinish() {
	p.interrupted = true
	fmt.Printf("\n--- %s:%d ping statistics ---\n", p.L4PingOptions.Host, p.L4PingOptions.Port)
	total := p.Result.Received + p.Result.Dropped
	packetLoss := float64(p.Result.Dropped) / float64(total) * 100
	fmt.Printf("%d packets transmitted, %d received, %.2f%% packet loss\n", total, p.Result.Received, packetLoss)
	code := 0
	if p.Result.Dropped > 0 {
		code = 2
		if p.Result.Received == 0 {
			code++
		}
	}
	os.Exit(code)
}

func L4Ping(p *L4Pinger) error {
	// Listen for Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			p.OnFinish()
		}
	}()

	err := p.Ping()
	if err != nil {
		return err
	}
	p.OnFinish()
	return nil
}

func NewL4Pinger(pingOptions *options.PingOptions) *L4Pinger {
	l := &L4Pinger{
		L4PingOptions: pingOptions,
		Result:        &PingResult{},
		interrupted:   false,
	}
	return l
}
