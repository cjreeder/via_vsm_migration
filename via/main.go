package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cjreeder/via_networking_script/via"
	"github.com/fatih/color"
)

type ViaList struct {
	vianame    string
	oldaddress string
	ipaddress  string
	subnetmask string
	gateway    string
	dns        string
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

func SetNetwork(vianame string, oldaddress string, ipaddress string, subnetmask string, gateway string, dns string) error {
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

func GetNetwork(vianame string, ipaddress string) {
	defer color.Unset()
	color.Set(color.FgGreen)
	file, err := os.OpenFile("/tmp/VIA_Network_Check.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	defer file.Close()

	var command via.Command
	command.Command = "IpInfo"
	fmt.Printf("Getting IP Info for %s\n", vianame)
	resp, err := via.SendCommand(command, ipaddress)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	result := vianame + " - " + resp
	_, err = file.WriteString(result)
	file.Sync()
}

func main() {
	lines, err := ReadCsv("/home/creeder/Desktop/tnrb_via_replacement_addresses2.csv")
	if err != nil {
		fmt.Printf("File cannot be found or read: %v", err.Error())
	}

	// loop through the lines and turn it into an object
	for _, line := range lines {
		data := ViaList{
			vianame:    line[0],
			oldaddress: line[1],
			ipaddress:  line[2],
			subnetmask: line[3],
			gateway:    line[4],
			dns:        line[5],
		}
		fmt.Printf("Changing over %v\n", data.vianame)
		err := SetNetwork(data.vianame, data.oldaddress, data.ipaddress, data.subnetmask, data.gateway, data.dns)
		if err != nil {
			fmt.Printf("%v returned an error: %v\n", data.vianame, err)
		} else {
			fmt.Printf("Change over script sent to %v\n", data.vianame)
		}
		time.Sleep(120 * time.Second)
		GetNetwork(data.vianame, data.ipaddress)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
	}
}
