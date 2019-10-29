package ecoin

import (
	"fmt"
	"testing"
	"time"
)

func TestNewChain(t *testing.T) {
	fmt.Println("init chain starts ......")
	msg := fmt.Sprintf("%s | %s",
		time.Now().Format("2006/01/02 15:04:05"),
		"Eiger created the chain")
	_, err := NewChain("127.0.0.1", 9999, msg)
	if err != nil {
		t.Errorf("init chain failed: %s\n", err)
	} else {
		fmt.Println("init chain success!")
		fmt.Println("Tip: Now you can startnode to listen all requests in the ecare world!")
	}
}

func TestStartNode(t *testing.T) {
	err := StartNode("127.0.0.1", 9999)
	if err != nil {
		t.Error(err)
	}
}