package queue

import (
	"../config"
	"log"
)

func CalculateCost(targetFloor, targetButton, prevFloor, currFloor, currDir int) int {
	q := local.deepCopy()

	// q := new(queue)
	// for f := 0; f < config.N_Floors; f++ {
	// 	for b := 0; b < config.N_Buttons; b++ {
	// 		q.matrix[f][b] = local.matrix[f][b]
	// 	}
	// }




	q.setOrder(targetFloor, config.BtnInside, orderStatus{true, "", nil})

	cost := 0
	floor := prevFloor
	dir := currDir

	if currFloor == -1 {
		// Between floors, add 1 cost.
		cost++
	} else if dir != config.DirStop {
		// At floor, but moving, add 2 cost.
		cost += 2
	}
	floor, dir = incrementFloor(floor, dir)

	// Simulate how the lift will move, and accumulate cost until it 'reaches' target.
	// Break after 10 iterations to assure against a stuck loop.
	for n := 0; !(floor == targetFloor && q.StopElevator(floor, dir)) && n < 10; n++ {
		if q.StopElevator(floor, dir) {
			cost += 2
			q.setOrder(floor, config.BtnUp, inactive)
			q.setOrder(floor, config.BtnDown, inactive)
			q.setOrder(floor, config.BtnInside, inactive)
		}
		dir = q.chooseDirection(floor, dir)
		floor, dir = incrementFloor(floor, dir)
		cost += 2
	}
	return cost
}

func incrementFloor(floor, dir int) (int, int) {
	switch dir {
	case config.DirDown:
		floor--
	case config.DirUp:
		floor++
	case config.DirStop:
	default:
		config.CloseConnectionChan <- true
		config.Restart.Run()
	}

	if floor <= 0 && dir == config.DirDown {
		dir = config.DirUp
		floor = 0
	}
	if floor >= config.N_Floors-1 && dir == config.DirUp {
		dir = config.DirDown
		floor = config.N_Floors - 1
	}
	return floor, dir
}
