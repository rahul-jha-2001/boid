package main

import "log"

func main() {
	sim := NewSim(500, 300, 10000)
	srv := NewServer(sim)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
