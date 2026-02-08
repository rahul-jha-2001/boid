package main

import (
	"context"
	"encoding/binary"
	"log"
	"net/http"
	"sync"
	"time"
	"unsafe"

	"nhooyr.io/websocket"
)

type Server struct {
	sim *Sim

	mu      sync.RWMutex // protects sim world forces (and any future shared edits)
	clients map[*websocket.Conn]struct{}
	paused  bool
}

func NewServer(sim *Sim) *Server {
	return &Server{
		sim:     sim,
		clients: make(map[*websocket.Conn]struct{}),
		paused:  true,
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // localhost dev
	})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()

	// read loop for commands
	ctx := r.Context()
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			break
		}
		s.applyCommand(data)
	}

	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()
}

func (s *Server) applyCommand(data []byte) {
	if len(data) < 1 {
		return
	}
	t := data[0]
	log.Printf("CMD %d len=%d", t, len(data))
	switch t {

	case CmdSetPaused:
		// u8 type, u8 paused(0/1)
		if len(data) < 2 {
			return
		}
		p := data[1] != 0
		s.mu.Lock()
		s.paused = p
		s.mu.Unlock()

	case CmdAddEmitter:
		// u8 type, u32 id, f32 x,y,amp,freq,phase,radius
		if len(data) < 1+4+6*4 {
			return
		}
		id := binary.LittleEndian.Uint32(data[1:5])
		x := f32(data[5:9])
		y := f32(data[9:13])
		amp := f32(data[13:17])
		freq := f32(data[17:21])
		phase := f32(data[21:25])
		radius := f32(data[25:29])

		log.Printf("CmdAddEmitter: id=%d, pos=(%.1f, %.1f), amp=%.1f, freq=%.1f, phase=%.1f, radius=%.1f",
			id, x, y, amp, freq, phase, radius)

		s.mu.Lock()
		s.sim.World.Emitters[id] = &Emitter{
			Id:     id,
			Pos:    V(float64(x), float64(y)),
			AmpN:   float64(amp),
			FreqHz: float64(freq),
			Phase:  float64(phase),
			Radius: float64(radius),
		}
		s.mu.Unlock()

	case CmdMoveEmitter:
		// u8 type, u32 id, f32 x,y
		if len(data) < 1+4+2*4 {
			return
		}
		id := binary.LittleEndian.Uint32(data[1:5])
		x := f32(data[5:9])
		y := f32(data[9:13])

		s.mu.Lock()
		if em := s.sim.World.Emitters[id]; em != nil {
			em.Pos = V(float64(x), float64(y))
		}
		s.mu.Unlock()

	case CmdDeleteEmitter:
		if len(data) < 1+4 {
			return
		}
		id := binary.LittleEndian.Uint32(data[1:5])

		s.mu.Lock()
		delete(s.sim.World.Emitters, id)
		s.mu.Unlock()

	case CmdSetEmitterWave:
		if len(data) < 1+4+3*4 {
			return
		}
		id := binary.LittleEndian.Uint32(data[1:5])
		amp := f32(data[5:9])
		freq := f32(data[9:13])
		phase := f32(data[13:17])

		s.mu.Lock()
		if em := s.sim.World.Emitters[id]; em != nil {
			em.AmpN = float64(amp)
			em.FreqHz = float64(freq)
			em.Phase = float64(phase)
		}
		s.mu.Unlock()

	case CmdCreateArray:
		// u8 type, u32 baseID, u32 count, f32 cx, f32 cy, f32 spacing, f32 radius, f32 amp, f32 freq
		if len(data) < 1+4+4+6*4 {
			return
		}
		baseID := binary.LittleEndian.Uint32(data[1:5])
		count := binary.LittleEndian.Uint32(data[5:9])
		cx := f32(data[9:13])
		cy := f32(data[13:17])
		spacing := f32(data[17:21])
		radius := f32(data[21:25])
		amp := f32(data[25:29])
		freq := f32(data[29:33])

		log.Printf("CmdCreateArray: base=%d n=%d cx=%.1f cy=%.1f spacing=%.1f radius=%.1f amp=%.1f freq=%.1f",
			baseID, count, cx, cy, spacing, radius, amp, freq)

		s.mu.Lock()
		s.sim.World.SetLineArray(baseID, int(count), float64(cx), float64(cy), float64(spacing), float64(radius), float64(amp), float64(freq))
		s.mu.Unlock()

	case CmdSteerArray:
		// u8 type, u32 baseID, u32 count, f32 cx, f32 spacing, f32 freq, f32 c, f32 thetaRad
		if len(data) < 1+4+4+5*4 {
			return
		}
		baseID := binary.LittleEndian.Uint32(data[1:5])
		count := binary.LittleEndian.Uint32(data[5:9])
		cx := f32(data[9:13])
		spacing := f32(data[13:17])
		freq := f32(data[17:21])
		c := f32(data[21:25])
		theta := f32(data[25:29])

		log.Printf("CmdSteerArray: base=%d n=%d cx=%.1f spacing=%.1f freq=%.2f c=%.2f theta=%.2f",
			baseID, count, cx, spacing, freq, c, theta)

		s.mu.Lock()
		s.sim.World.SteerLineArray(baseID, int(count), float64(cx), float64(spacing), float64(freq), float64(c), float64(theta))
		s.mu.Unlock()

	case CmdReset:
		s.mu.Lock()
		s.sim = NewSim(500, 300, 10000)
		s.paused = true
		s.mu.Unlock()
		log.Printf("Simulation reset")
	}
}

func (s *Server) Run() error {
	// Sim loop at 120Hz
	go func() {
		ticker := time.NewTicker(time.Second / 120)
		defer ticker.Stop()
		for range ticker.C {
			s.mu.Lock()
			if !s.paused {
				s.sim.Step(1.0 / 120.0)
			}
			s.mu.Unlock()
		}
	}()

	// Broadcast loop at 30Hz
	go func() {
		ticker := time.NewTicker(time.Second / 30)
		defer ticker.Stop()
		for range ticker.C {
			s.broadcastSnapshot()
		}
	}()

	http.HandleFunc("/ws", s.handleWS)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("Open http://localhost:8080")
	return http.ListenAndServe("127.0.0.1:8080", nil)
}

func (s *Server) broadcastSnapshot() {
	// build snapshot buffer
	s.mu.RLock()
	tick := s.sim.Tick
	parts := s.sim.World.Particles
	s.mu.RUnlock()

	// 1 + 4 + 4 + n*(12)
	n := uint32(len(parts))
	buf := make([]byte, 1+4+4+int(n)*12)
	buf[0] = MsgSnapshot
	binary.LittleEndian.PutUint32(buf[1:5], tick)
	binary.LittleEndian.PutUint32(buf[5:9], n)

	off := 9
	for i := 0; i < int(n); i++ {
		putF32(buf[off:off+4], float32(parts[i].Pos.X))
		off += 4
		putF32(buf[off:off+4], float32(parts[i].Pos.Y))
		off += 4
		putF32(buf[off:off+4], float32(parts[i].Radius))
		off += 4
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// fanout
	for c := range s.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		_ = c.Write(ctx, websocket.MessageBinary, buf)
		cancel()
	}
}

func putF32(b []byte, v float32) { binary.LittleEndian.PutUint32(b, mathF32bits(v)) }
func f32(b []byte) float32       { return mathF32frombits(binary.LittleEndian.Uint32(b)) }

// local float32 bits helpers (avoid importing math package just for Float32bits)
func mathF32bits(f float32) uint32     { return *(*uint32)(unsafe.Pointer(&f)) }
func mathF32frombits(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }
