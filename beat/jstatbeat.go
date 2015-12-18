package beat

import (
	"bufio"
	"errors"
	"github.com/elastic/libbeat/beat"
	"github.com/elastic/libbeat/cfgfile"
	"github.com/elastic/libbeat/common"
	"github.com/elastic/libbeat/logp"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var blanks = regexp.MustCompile(`\s+`)

type Jstatbeat struct {
	// from configuration
	interval string
	name     string

	// state
	pid     string
	isAlive bool
}

func (jsb *Jstatbeat) Config(b *beat.Beat) error {
	var config ConfigSettings
	err := cfgfile.Read(&config, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	jsb.name = *config.Input.Name

	if config.Input.Interval != nil {
		jsb.interval = *config.Input.Interval
	} else {
		jsb.interval = "5000"
	}

	return nil
}

func (jsb *Jstatbeat) Setup(b *beat.Beat) error {
	cmd := exec.Command("jps")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logp.Err("Error get stdout pipe: %v", err)
		return err
	}

	cmd.Start()

	// TODO: handle error when 'jps' command cannot be executed.

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, " ")

		if len(items) == 2 && items[1] == jsb.name {
			jsb.pid = items[0]
			break
		}
	}
	cmd.Wait()

	if len(jsb.pid) == 0 {
		logp.Err("No target process: %v", jsb.name)
		return errors.New("No target process: " + jsb.name)
	}

	return nil
}

func (jsb *Jstatbeat) Run(b *beat.Beat) error {
	jsb.isAlive = true

	cmd := exec.Command("jstat", "-gc", "-t", jsb.pid, jsb.interval)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		logp.Err("Error get stdout pipe: %v", err)
		return err
	}

	cmd.Start()
	defer killProcess(cmd)

	scanner := bufio.NewScanner(stdout)
	var keys []string
	var version string
	for jsb.isAlive && scanner.Scan() {
		line := scanner.Text()

		values := blanks.Split(line, -1)

		if len(values) > 2 && values[0] == "Timestamp" {
			keys = values

			if strings.Contains(line, "CCSC") {
				version = "java8"
			} else {
				version = "java5"
			}

			continue
		}

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       version,
		}

		for i, key := range keys {
			if len(key) == 0 {
				continue
			}
			event[key] = toFloat(values[i+1])
		}
		b.Events.PublishEvent(event)
	}

	return nil
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			logp.Err("Error killing jstat process: %v", err)
		}
	}
}

func toFloat(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)

	if err != nil {
		logp.Err("Cannot parser to float. Ignore this value: %v", err)
		return 0
	}

	return value
}

func (jsb *Jstatbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (jsb *Jstatbeat) Stop() {
	jsb.isAlive = false
}
