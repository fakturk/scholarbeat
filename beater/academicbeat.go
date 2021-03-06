package beater

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/fakturk/academicbeat/config"
)

type Academicbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Academicbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

func (bt *Academicbeat) Run(b *beat.Beat) error {
	logp.Info("academicbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		// Get filepath for currently running script.
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		// Execute python script that pulls google scholar product data.
		fmt.Println(dir+"/scholar.py/scholar.py", " --author \"albert einstein\"")
		// cmd := exec.Command("bash", dir+"/scholar/scholar.py")
		args := []string{dir + "/scholar.py/scholar.py"}
		args = append(args, "--author")
		args = append(args, "albert einstein")
		// cmd := exec.Command("python", args...)
		cmd := exec.Command("python", dir+"/scholar.py/scholar.py", "--author", "albert einstein")
		// var out bytes.Buffer
		// cmd.Stdout = &out
		// cmdErr := cmd.Run()
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			fmt.Println(fmt.Sprint(cmdErr) + ": " + string(out))
			log.Fatal(cmdErr)
		}
		fmt.Printf("%q\n", out)
		// fmt.Println("python script sonrasi")

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}
}

func (bt *Academicbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
