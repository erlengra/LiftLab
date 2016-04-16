package queue

import (
	"../config"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func Backup(networkPackage chan<- config.Message) {
	const filename = "backup"
	var backup queue
	backup.load(filename)

	if !backup.requestOrderInfo(-1, -1, isEmpty) {
		log.Println("Backup not empty")
		for f := 0; f < config.N_Floors; f++ {
			for b := 0; b < config.N_Buttons; b++ {
				if backup.requestOrderInfo(f, b, isInQueue) {
					if b == config.BtnInside {
						AddToLocalQueue(f, b)
					} else {
						networkPackage <- config.Message{Category: config.NewOrder, Floor: f, Button: b}
					}
				}
			}
		}
	}
	go func() {
		for {
			<-takeBackup
			if err := local.Save(filename); err != nil {
				log.Println(err)
			}
		}
	}()
}

func (q *queue) save(filename string) error {
	data, err := json.Marshal(&q)
	if err != nil {
		return err
	}
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (q *queue) load(filename string) error {
	log.Println("looking for backup...")
	err := os.Stat(filename)
	if err == nil {
		log.Println("Backup file found")

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println("\x1b[31;1m","load error","\x1b[0m")
		}
		err := json.Unmarshal(data, q)
		if err != nil {
			log.Println("\x1b[31;1m","load error","\x1b[0m")
		}
	}
	return nil
}
