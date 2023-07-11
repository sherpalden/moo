package game

import (
	"fmt"
	"sync"
)

var (
	nums = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	// DebugMode makes log output enable
	DebugMode = true
)

type (
	// Question returns hits, blow by guess []int
	Question func(guess []int) (hits, blow int)
	// Estimate returns a next guess
	Estimate func(q Question) (guess []int)
	// Game means one game field
	Game struct {
		difficulty int
		answer     []int
	}
)

// NewGame returns a new game field
func NewGame(d int) *Game {
	if d < 1 || d > 9 {
		fmt.Println(d, "is invalid moo digit, difficulty set to", d)
		d = 4
	}
	return &Game{
		difficulty: d,
		answer:     GetMooNum(d),
	}
}

// GetDifficulty returns digits
func (x *Game) GetDifficulty() int {
	return x.difficulty
}

// GetAnswer returns answer
func (x *Game) GetAnswer() []int {
	return x.answer
}

// GetQuestion returns a question func
func (x *Game) GetQuestion(count *int) Question {
	*count = 0
	return func(g []int) (h, b int) {
		*count++
		h = x.GetHit(g)
		b = x.GetBlow(g)
		if DebugMode {
			fmt.Println(g, ": hits:", h, "blow:", b)
		}
		return h, b
	}
}

// GetHit returns hit count in this game
func (x *Game) GetHit(g []int) int {
	return GetHit(g, x.answer)
}

// GetBlow returns blow count in this game
func (x *Game) GetBlow(g []int) int {
	return GetBlow(g, x.answer)
}

// Equals returns bool which guess = answer
func (x *Game) Equals(g []int) bool {
	return Equals(g, x.answer)
}

// GetHit returns hit
func GetHit(guess []int, answer []int) int {
	count := 0
	if len(guess) != len(answer) {
		return 0
	}
	for i, v := range answer {
		if guess[i] == v {
			count++
		}
	}
	return count
}

// GetBlow returns blow
func GetBlow(guess []int, answer []int) int {
	count := 0
	if len(guess) != len(answer) {
		return 0
	}
	for i, g := range guess {
		for j, a := range answer {
			if g == a && i != j {
				count++
			}
		}
	}
	return count
}

// returns the estimator to guess the answer
func GetEstimater(difficulty int) Estimate {
	type guessPosition struct {
		pos int
		val int
	}
	operate := func(q Question, pos int, gs chan<- guessPosition) {
		g := make([]int, difficulty)
		h, _ := q(g)
		for {
			g[pos]++
			newh, _ := q(g)
			if newh > h {
				gs <- guessPosition{pos: pos, val: g[pos]}
				break
			} else if newh < h {
				gs <- guessPosition{pos: pos, val: 0}
				break
			}
		}
	}
	return func(q Question) []int {
		guess := make([]int, difficulty)
		guessStream := func() <-chan guessPosition {
			gs := make(chan guessPosition)
			wg := sync.WaitGroup{}

			for i := 0; i < difficulty; i++ {
				wg.Add(1)
				go func(pos int) {
					defer wg.Done()
					operate(q, pos, gs)
				}(i)
			}

			go func() {
				wg.Wait()
				close(gs)
			}()

			return gs
		}()

		for g := range guessStream {
			guess[g.pos] = g.val
		}
		return guess
	}
}
