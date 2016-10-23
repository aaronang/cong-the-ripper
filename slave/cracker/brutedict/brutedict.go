package brutedict

import (
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

type BruteDict struct {
	task  *lib.Task
	queue chan []byte
	quit  chan bool
}

func New(task *lib.Task) (bd *BruteDict) {
	bd = &BruteDict{
		task:  task,
		queue: make(chan []byte),
		quit:  make(chan bool),
	}

	go bd.list(task.Alphabet, task.Start, task.TaskLen)
	return
}

func (bd *BruteDict) Next() (candidate []byte) {
	select {
	case candidate = <-bd.queue:
	case <-bd.quit:
	}
	return
}

func (bd *BruteDict) Close() {
	close(bd.queue)
}

func (bd *BruteDict) list(alph lib.Alphabet, start []byte, length int64) {
	defer func() { recover() }()
	len := len(start)
	currentComb := lib.BytesToBigInt(alph, start)

	for length > 0 {
		byteArray, overflow := lib.BigIntToBytes(alph, currentComb, len)
		if overflow {
			break
		}
		bd.queue <- lib.ReverseArray(byteArray)
		currentComb.Add(currentComb, big.NewInt(1))
		length--
	}

	bd.quit <- true
}
