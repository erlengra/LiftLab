package queue

import (
	def "config"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func runBackup(outgoingMsg chan<- def.Message) {
	const filename = "lift_backup"

	var backup queue
	backup.loadFromDisk(filename)

	if !backup.isEmpty() {
		for f := 0; f < def.N_Floors; f++ {
			for b := 0; b < def.N_Buttons; b++ {
				if backup.isOrder(f, b) {
					if b == def.BtnInside {
						SetOrderLocal(f, b)
					} else {
						outgoingMsg <- def.Message{Category: def.NewOrder, Floor: f, Button: b}
					}
				}
			}
		}
	}
	go func() {
		for {
			<-takeBackup
			if err := local.saveToDisk(filename); err != nil {
				log.Println(def.ColR, err, def.ColN)
			}
		}
	}()
}

func (q *queue) saveToDisk(filename string) error {

	data, err := json.Marshal(&q)
	if err != nil {
		log.Println(def.ColR, "json.Marshal() error: Failed to backup.", def.ColN)
		return err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Println(def.ColR, "ioutil.WriteFile() error: Failed to backup.", def.ColN)
		return err
	}
	return nil
}

func (q *queue) loadFromDisk(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		log.Println(def.ColG, "Backup file found, processing...", def.ColN)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Println(def.ColR, "loadFromDisk() error: Failed to read file.", def.ColN)
		}
		if err := json.Unmarshal(data, q); err != nil {
			log.Println(def.ColR, "loadFromDisk() error: Failed to Unmarshal.", def.ColN)
		}
	}
	return nil
}
