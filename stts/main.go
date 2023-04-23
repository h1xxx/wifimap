package main

import (
	"flag"
	"os"
	"regexp"
	"time"

	fp "path/filepath"

	"github.com/mdlayher/wifi"
)

type sttsT struct {
	uptime time.Duration
	loads  [3]float64
	procs  int
	mem    memT

	cpu1Temps  []string
	cpu2Temps  []string
	driveTemps []string
	moboTemps  []string

	wifiBss  *wifi.BSS
	wifiInfo *wifi.StationInfo

	batLevel    string
	batTimeLeft string
}

type memT struct {
	total  int
	used   int
	free   int
	shared int
	buffer int
	cache  int
	avail  int
}

type varsT struct {
	meminfoFd *os.File

	cpu1TempHwmon string
	cpu2TempHwmon string
	cpu1TempFds   []*os.File
	cpu2TempFds   []*os.File

	driveTempHwmons []string
	driveTempFds    []*os.File

	moboTempHwmons []string
	i2cMoboTemps   []string
	moboTempFds    []*os.File

	miscHwmonNames []string
	miscI2cNames   []string

	wifiClient *wifi.Client
	wifiIface  *wifi.Interface

	batCapacityFd   *os.File
	batEnergyFd     *os.File
	batEnergyFullFd *os.File
	batPowerFd      *os.File
	batStatusFd     *os.File
}

type argsT struct {
	bench *bool
}

var args argsT

func init() {
	args.bench = flag.Bool("b", false, "perform a benchmark")
}

func main() {
	flag.Parse()

	var vars varsT
	getVars(&vars)

	var st sttsT
	getAllInfo(&st, &vars)
	prettyPrint(st, vars)

	if *args.bench {
		doBench(&vars)
	}

	closeFiles(&vars)
}

func getAllInfo(st *sttsT, vars *varsT) {
	getSysinfo(st, vars)
	readCpuTemps(st, vars)
	readDriveTemps(st, vars)
	readMoboTemps(st, vars)
	getWifiInfo(st, vars)
	getBatInfo(st, vars)
}

func getVars(vars *varsT) {
	var err error
	vars.meminfoFd, err = os.Open("/proc/meminfo")
	errExit(err)

	hwmonDetect(vars)
	i2cDetect(vars)
	detectWlan(vars)
	detectBat(vars)

	vars.cpu1TempFds = openHwmon(vars.cpu1TempHwmon, "temp.*_input")
	vars.cpu2TempFds = openHwmon(vars.cpu2TempHwmon, "temp.*_input")

	for _, hwmon := range vars.driveTempHwmons {
		vars.driveTempFds = append(vars.driveTempFds,
			openHwmon(hwmon, "temp.*_input")...)
	}

	for _, hwmon := range vars.moboTempHwmons {
		vars.moboTempFds = append(vars.moboTempFds,
			openHwmon(hwmon, "temp.*_input")...)
	}

	vars.moboTempFds = append(vars.moboTempFds,
		openFiles(vars.i2cMoboTemps)...)
}

func openHwmon(hwmonDir string, ex string) []*os.File {
	var fds []*os.File

	re := regexp.MustCompile(ex)

	hwmonFiles, err := os.ReadDir(hwmonDir)
	if err != nil {
		return fds
	}

	for _, hwmonFile := range hwmonFiles {
		if !re.MatchString(hwmonFile.Name()) {
			continue
		}

		file := fp.Join(hwmonDir, hwmonFile.Name())

		fd, err := os.Open(file)
		if err != nil {
			continue
		}

		fds = append(fds, fd)
	}

	return fds
}

func openFiles(files []string) []*os.File {
	var fds []*os.File

	for _, file := range files {
		fd, err := os.Open(file)
		if err != nil {
			continue
		}

		fds = append(fds, fd)
	}

	return fds
}

func closeFiles(vars *varsT) {
	vars.meminfoFd.Close()

	for _, fd := range vars.cpu1TempFds {
		fd.Close()
	}

	for _, fd := range vars.cpu2TempFds {
		fd.Close()
	}

	for _, fd := range vars.driveTempFds {
		fd.Close()
	}

	for _, fd := range vars.moboTempFds {
		fd.Close()
	}

	if vars.wifiClient != nil {
		vars.wifiClient.Close()
	}

	if vars.batCapacityFd != nil {
		vars.batCapacityFd.Close()
	}

	if vars.batEnergyFd != nil {
		vars.batEnergyFd.Close()
	}

	if vars.batEnergyFullFd != nil {
		vars.batEnergyFullFd.Close()
	}

	if vars.batPowerFd != nil {
		vars.batPowerFd.Close()
	}

	if vars.batStatusFd != nil {
		vars.batStatusFd.Close()
	}
}
