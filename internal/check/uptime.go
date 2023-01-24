// Copyright (c) 2021 deadc0de6

package check

import (
	"fmt"
	"time"
	"strconv"
	"strings"

	"github.com/deadc0de6/checkah/internal/transport"
)

// Uptime the uptime struct
type Uptime struct {
	limitDays int
	options   map[string]string
}

func (c *Uptime) returnCheck(value string, err error) *Result {
	limits := fmt.Sprintf("%d days", c.limitDays)
	return &Result{
		Name:        c.GetName(),
		Description: c.GetDescription(),
		Value:       value,
		Limit:       limits,
		Error:       err,
	}
}

func uptimeFromProc(stdout string) (int, error) {
	lines := strings.Split(stdout, " ")
	if len(lines) < 1 {
		return 0, fmt.Errorf("getting uptime failed")
	}
	seconds := lines[0]

	val, err := strconv.ParseFloat(seconds, 64)
	if err != nil {
		return 0, err
	}

	nbDays := int(int(val) / 60 / 60 / 24)
	return nbDays, nil
}

func uptimeFromLoadavg(stdout string) (int, error) {
	lines := strings.Split(stdout, " ")
	if len(lines) < 1 {
		return 0, fmt.Errorf("getting uptime failed")
	}

	dt := lines[0]
	fields := strings.Split(dt, ":")
	var h, m, s int
	if len(fields) > 2 {
		_, err := fmt.Sscanf(dt, "%d:%d:%d", &h, &m, &s)
		if err != nil {
			return 0, err
		}
	} else {
		_, err := fmt.Sscanf(dt, "%d:%d", &h, &m)
		s = 0
		if err != nil {
			return 0, err
		}
	}

	// transform to days
	nbDays := float32(h) / 24.0
	nbDays += float32(m) / 60.0 / 24.0
	return int(nbDays), nil
}

// Run executes the check
func (c *Uptime) Run(t transport.Transport) *Result {
	cmd := "sysctl kern.boottime | awk '{print $5}' | cut -d',' -f1"
	stdout, _, err := t.Execute(cmd)

	if err != nil {
		return c.returnCheck("", err)
	}

	boot_time_s  := strings.TrimSpace(string(stdout[:]))
	boot_time, _ := strconv.ParseInt(boot_time_s, 10, 64)
	cur_time := time.Now().Unix()
	val := (cur_time - boot_time) / 60 / 60 / 24

	if val > int64(c.limitDays) {
		return c.returnCheck("", fmt.Errorf("uptime above %d days: %d days", c.limitDays, val))
	}

	return c.returnCheck(fmt.Sprintf("%d days", val), nil)
}

// GetName returns the check name
func (c *Uptime) GetName() string {
	return "uptime"
}

// GetDescription get description
func (c *Uptime) GetDescription() string {
	return "uptime"
}

// GetOptions returns the options
func (c *Uptime) GetOptions() map[string]string {
	return c.options
}

// NewCheckUptime creates a disk check instance
func NewCheckUptime(options map[string]string) (*Uptime, error) {
	days := -1
	v, ok := options["days"]
	if ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		days = i
	}

	c := Uptime{
		limitDays: days,
		options:   options,
	}

	return &c, nil
}
