package utils

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var mu sync.Mutex
var seededRand *rand.Rand

func init() {
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateOrderID(userID uint) string {
	mu.Lock()
	defer mu.Unlock()

	nowNano := time.Now().UnixNano()
	nanoPart := nowNano % 1000000

	randPart := seededRand.Intn(900) + 100

	return fmt.Sprintf("MNR-%06d%03d%d", nanoPart, randPart, userID)
}

func GenerateReferenceID(userID uint) string {
	mu.Lock()
	defer mu.Unlock()

	nowNano := time.Now().UnixNano()
	nanoPart := nowNano % 1000000

	randPart := seededRand.Intn(900) + 100

	return fmt.Sprintf("MNR-%06d%03d%d", nanoPart, randPart, userID)
}
