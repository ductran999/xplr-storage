package main

import (
	"log"
	"storage-journey/lsm_tree/engine"
)

func main() {
	lsm, err := engine.NewLSMTree("data", 100)
	if err != nil {
		log.Fatalln("start lsm error", err)
	}

	lsm.Put("dog", []byte("animal"))
	lsm.Put("apple", []byte("red"))
	lsm.Put("bird", []byte("animal"))
	lsm.Put("banana", []byte("fruit"))
	lsm.Put("daniel", []byte(`{"first_name":"daniel","age":"27","phone_num":"01234567879"}`))
	lsm.Put("jenny", []byte(`{"first_name":"jenny","age":"27","phone_num":"01234567879"}`))

	val, found := lsm.Get("apple")
	if found {
		log.Println("found apple:", string(val))
	}
}
