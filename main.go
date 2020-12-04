package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cjreeder/via_networking_script/via"
	"github.com/fatih/color"
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

func SetVSM(vianame string, oldaddress string, ipaddress string, subnetmask string, gateway string, dns string) error {
	defer color.Unset()
	color.Set(color.FgYellow)

	address := oldaddress

	var command via.Command
	command.Command = "IpSetting"
	command.Param1 = ipaddress
	command.Param2 = subnetmask
	command.Param3 = gateway
	command.Param4 = dns
	command.Param5 = vianame

	fmt.Printf("Setting IP Info for %s\n", vianame)
	err := via.SendonlyCommand(command, address)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in setting IP on %s\n", vianame))
	}
	return nil
}

func Reboot() error {
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
		ifile string
		ofile string
		vsm   string
		count int
		wg    sync.WaitGroup
	)

	pflag.StringVarP(&ifile, "input", "i", "", "Input file containing a list of VIAs to Migrate")
	pflag.StringVarP(&ofile, "output", "o", "", "file to log all output to")
	pflag.StringVarP(&vsm, "vsm", "v", "", "VIA Site Management Server IP Address")
	pflag.IntVar(&count, "Processing Number", "n", "1000", "Size of Channel")
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
