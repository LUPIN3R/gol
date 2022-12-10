package stubs

var StartGameOfLife = "GameOfLifeCalls.StartGameOfLife"
var GetAliveNumber = "GameOfLifeCalls.GetAliveNumber"
var GetCurrentWorld = "GameOfLifeCalls.GetCurrentWorld"
var EndGameOfLife = "GameOfLifeCalls.EndGameOfLife"
var PauseOrContinue = "GameOfLifeCalls.PauseOrContinue"
var QuitServer = "GameOfLifeCalls.QuitServer"

type Request struct {
	World [][]uint8
	Turns int
	ImageWidth  int
	ImageHeight int
}

type Response struct {
	World [][]uint8
	CurTurn int
	AliveNumber int
}