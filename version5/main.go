package main

func main() {
	server := NewServer("192.168.56.105", 8888)
	server.Start()
}
