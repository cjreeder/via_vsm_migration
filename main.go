package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	via "github.com/cjreeder/via_vsm_migration/via"
	"github.com/spf13/pflag"
)

type ViaList struct {
	gateway_id string
	vianame    string
	ipaddress  string
}

func ReadCsv(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// read file into a variable to be able to usue later
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func SetVSM(vsm string, vianame string, gateway string, address string) error {

	var command via.Command
	command.Command = "VSMInfo"
	command.Param1 = "Set"
	command.Param2 = vsm
	command.Param3 = gateway

	fmt.Printf("Setting IP Info for %s\n", vianame)
	err := via.SendonlyCommand(command, address)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in setting IP on %s\n", vianame))
	}
	return nil
}

func Reboot(vianame string, address string) error {
	var cmd via.command
	command.Command = "Reboot"

	fmt.Printf("Rebooting: %s \n", vianame)
	err := via.SendonlyCommand(cmd, address)
	if err != nil {
		return errors.New(fmt.Sprintf("Error rebooting VIA: %s\n", vianame))
	}
	return nil
}

func workers(i int, vsm string, wg *sync.WaitGroup, requests <-chan ViaList) {
	defer wg.Done()
	for req := range requests {
		fmt.Printf("Worker: Working on %s\n", req.vianame)
		err := SetVSM(vsm, req.vianame, req.gateway_id, req.ipaddress)
		if err != nil {
			fmt.Printf(err)
		}
		err := Reboot(req.vianame, req.ipaddress)
		if err != nil {
			fmt.Printf(err)
		}
	}
	fmt.Printf("Worker Thread: %v - has complete and is now exiting....\n", i)
	time.Sleep(10 * time.Second)
}

func main() {
	var (
		ifile     string
		ofile     string
		vsm       string
		count     int
		maxThread int
		wg        sync.WaitGroup
	)

	pflag.StringVarP(&ifile, "input", "i", "", "Input file containing a list of VIAs to Migrate")
	pflag.StringVarP(&ofile, "output", "o", "", "file to log all output to")
	pflag.StringVarP(&vsm, "vsm", "v", "", "VIA Site Management Server IP Address")
	pflag.IntVar(&count, "channel_count", "c", "1000", "Size of Channel")
	pflag.IntVar(&maxThread, "maxthread", "m", "10", "Maximum Number of Threads")
	pflag.Parse()

	lines, err := ReadCsv(ifile)
	if err != nil {
		fmt.Printf("File cannot be found or read: %v", err.Error())
	}

	var ch = make(chan int, count) // This number 50 can be anything as long as it's larger than xthreads

	// loop through the lines and turn it into an object
	for _, line := range lines {
		data := ViaList{
			vianame:    line[0],
			gateway_id: line[1],
			ipaddress:  line[2],
		}
		fmt.Printf("Changing over %v\n", data.vianame)
		requests <- data
	}
	for i := 0; i < maxThread; i++ {
		wg.Add(1)
		go workers(i, vsm, &wg, requests)
	}
	close(requests) // This tells the goroutines there's nothing else to do
	wg.Wait()       // Wait for the threads to finish}
}
