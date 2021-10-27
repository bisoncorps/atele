package inspector

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

// DFMetrics : Metrics used by DF
type DFMetrics struct {
	Size        float64
	Used        float64
	Available   float64
	PercentFull int
}

// DF : Parsing the `df` output for disk monitoring
type DF struct {
	fields
	// The values read from the command output string are defaultly in KB
	RawByteSize string
	// We want do display disk values in GB
	DisplayByteSize string
	// Parse only device that start with this e.g /dev/sd
	DeviceStartsWith string
	// Mount point to examine
	MountPoint string
	// Values of metrics being read
	Values []DFMetrics
}

// Parse : run custom parsing on output of the command
func (i *DF) Parse(output string) {
	var values []DFMetrics
	log.Debug("Parsing ouput string in DF inspector")
	lines := strings.Split(output, "\n")
	for index, line := range lines {
		// skip title line
		if index == 0 {
			continue
		}
		columns := strings.Fields(line)
		if len(columns) == 6 {
			percent := columns[4]
			if len(percent) > 1 {
				percent = percent[:len(percent)-1]
			} else if percent == `-` {
				percent = `0`
			}
			percentInt, err := strconv.Atoi(percent)
			if err != nil {
				log.Fatalf(`Error Parsing Percent Full: %s `, err)
			}
			if columns[5] == i.MountPoint {
				values = append(values, i.createMetric(columns, percentInt))
			} else if strings.HasPrefix(columns[0], i.DeviceStartsWith) &&
				i.MountPoint == "" {
				values = append(values, i.createMetric(columns, percentInt))
			}
		}
	}
	i.Values = values
}

func (i DF) createMetric(columns []string, percent int) DFMetrics {
	return DFMetrics{
		Size:        NewByteSize(columns[1], i.RawByteSize).format(i.DisplayByteSize),
		Used:        NewByteSize(columns[2], i.RawByteSize).format(i.DisplayByteSize),
		Available:   NewByteSize(columns[3], i.RawByteSize).format(i.DisplayByteSize),
		PercentFull: percent,
	}
}

// NewDF : Initialize a new DF instance
func NewDF() *DF {
	return &DF{
		fields: fields{
			Type:    Command,
			Command: `df -a`,
		},
		RawByteSize:     `KB`,
		DisplayByteSize: `GB`,
		MountPoint:      `/`,
	}

}