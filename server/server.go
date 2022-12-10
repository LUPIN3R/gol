package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
)

const ALIVE = 255
const DEAD  = 0

var curTurn int
var curWorld [][]uint8
var aliveNumber int
var stateMutex, endMutex sync.Mutex
var isEndGeneration bool = false
var isPauseGeneration bool = false
var pauseBlock chan bool = make(chan bool)
var isQuitServer bool
var isFinishGOL bool = false
type GameOfLifeCalls struct {}

func calculateAliveNeighbors(world [][]uint8, width int, height int, y, x int) int {
	var aliveNeighborNum = 0
	posBias := [8][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	for i := 0; i < 8; i++ {
		curY := (y + posBias[i][0] + height) % height
		curX := (x + posBias[i][1] + width) % width
		if world[curY][curX] == ALIVE {
			aliveNeighborNum++
		}
	}
	return aliveNeighborNum
}

func nextGeneration(world [][]uint8, startHeight int, endHeight int, width int, height int, myNextWorld [][]uint8) int {
	curAliveNum := 0
	for i := 0; i < endHeight - startHeight; i++ {
		for j := 0; j < width; j++ {
			actualHeight := i + startHeight
			curState := world[actualHeight][j]
			aliveNeighborNum := calculateAliveNeighbors(world, width, height, actualHeight, j)
			if curState == ALIVE && aliveNeighborNum < 2 {
				myNextWorld[i][j] = DEAD
			}
			if (curState == ALIVE && aliveNeighborNum == 2) || (curState == ALIVE && aliveNeighborNum == 3) {
				myNextWorld[i][j] = ALIVE
				curAliveNum++
			}
			if curState == ALIVE && aliveNeighborNum > 3 {
				myNextWorld[i][j] = DEAD
			}
			if curState == DEAD && aliveNeighborNum == 3 {
				myNextWorld[i][j] = ALIVE
				curAliveNum++
			}
		}
	}
	return curAliveNum
}

func (s *GameOfLifeCalls) StartGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	isFinishGOL = false
	stateMutex.Lock()
	curWorld = req.World
	curTurn = 0
	aliveNumber = 0
	stateMutex.Unlock()
	for curTurn < req.Turns {
		if isQuitServer {
			break
		}
		if isEndGeneration {
			endMutex.Lock()
			isEndGeneration = false
			endMutex.Unlock()
			break
		}
		if isPauseGeneration {
			// wait until 'p' is pressed again
			<-pauseBlock
		}
		nextWorld := make([][]uint8, req.ImageHeight)
		for i := range nextWorld {
			nextWorld[i] = make([]uint8, req.ImageWidth)
		}
		curAliveNum := nextGeneration(curWorld, 0, req.ImageHeight, req.ImageWidth, req.ImageHeight, nextWorld[:][:])

		stateMutex.Lock()
		curWorld = nextWorld
		aliveNumber = curAliveNum
		curTurn++
		stateMutex.Unlock()
	}
	fmt.Printf("Finished total turns %d\n", curTurn)
	res.CurTurn = curTurn
	res.World = curWorld
	isFinishGOL = true
	return
}

func (s *GameOfLifeCalls) GetAliveNumber(req stubs.Request, res *stubs.Response) (err error) {
	stateMutex.Lock()
	res.AliveNumber = aliveNumber
	res.CurTurn = curTurn
	stateMutex.Unlock()
	return
}

func (s *GameOfLifeCalls) GetCurrentWorld(req stubs.Request, res *stubs.Response) (err error) {
	stateMutex.Lock()
	res.World = curWorld
	res.CurTurn = curTurn
	stateMutex.Unlock()
	return
}

func (s *GameOfLifeCalls) EndGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	stateMutex.Lock()
	res.World = curWorld
	res.CurTurn = curTurn
	res.AliveNumber = aliveNumber
	stateMutex.Unlock()	
	endMutex.Lock()
	isEndGeneration = true
	endMutex.Unlock()
	return
}

func (s *GameOfLifeCalls) PauseOrContinue(req stubs.Request, res *stubs.Response) (err error) {
	stateMutex.Lock()
	if isPauseGeneration {
		pauseBlock <- false
	}
	isPauseGeneration = !isPauseGeneration
	res.CurTurn = curTurn + 1
	stateMutex.Unlock()
	return
}

func (s *GameOfLifeCalls) QuitServer(req stubs.Request, res *stubs.Response) (err error) {
	stateMutex.Lock()
	res.World = curWorld
	res.CurTurn = curTurn
	res.AliveNumber = aliveNumber
	stateMutex.Unlock()	
	isQuitServer = true
	// wait for StartGameOfLife to finish
	for !isFinishGOL {
		// do Nothing
	}
	fmt.Println("The Server Has Quitted")
	time.Sleep(1 * time.Second)
	os.Exit(0)
	return
}

func main(){
	// NOTE
	pAddr := flag.String("port","12345","Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&GameOfLifeCalls{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
