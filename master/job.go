package master

import (
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

type job struct {
	lib.Job
	id           int64
	tasks        []*lib.Task
	runningTasks int
	maxTasks     int
}

func (j *job) reachedMaxTasks() bool {
	return j.runningTasks < j.maxTasks
}

// splitJob attempts to split a cracking job into equal sized tasks regardless of the job
// the taskSize represents the number of brute force iterations
func (j *job) splitJob(taskSize int64) {
	if taskSize < int64(j.Iter) {
		panic("taskSize cannot be lower than job.Iter")
	}

	// adjust taskSize depending on the PBKDF2 rounds
	actualTaskSize := taskSize / int64(j.Iter)

	var tasks []*lib.Task
	cands, lens := chunkCandidates(j.Alphabet, j.KeyLen, actualTaskSize)
	for i := range cands {
		tasks = append(tasks, &lib.Task{
			Job:     j.Job,
			JobID:   j.id,
			ID:      int64(i),
			Start:   cands[i],
			TaskLen: lens[i]})
	}
	j.tasks = tasks
}

// chunkCandidates takes a character set and the required length l and splits to chunks of size n
func chunkCandidates(alph lib.Alphabet, l int, n int64) ([][]byte, []int64) {
	cand := alph.InitialCandidate(l)
	var cands [][]byte
	var lens []int64
	for {
		newCand, overflow := nthCandidateFrom(alph, n, cand)
		cands = append(cands, cand)
		if overflow {
			lens = append(lens, countUntilFinal(alph, cand))
			break
		} else {
			lens = append(lens, n)
		}
		cand = newCand
	}
	return cands, lens
}

// nthCandidateFrom computes the n th candidate password from inp
func nthCandidateFrom(alph lib.Alphabet, n int64, inp []byte) ([]byte, bool) {
	l := len(inp)
	z := lib.BytesToBigInt(alph, inp)
	z.Add(z, big.NewInt(n))
	return lib.BigIntToBytes(alph, z, l)
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
func countUntilFinal(alph lib.Alphabet, cand []byte) int64 {
	c := lib.BytesToBigInt(alph, cand)
	f := lib.BytesToBigInt(alph, alph.FinalCandidate(len(cand)))
	diff := f.Sub(f, c)
	return diff.Int64()
}
