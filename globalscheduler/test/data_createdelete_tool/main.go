/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//This functon creates cluster dataset up to 1000
//usage example: go run main.go -file=test.yaml -max=100
package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	OpenstackIP = "34.218.224.247"
)

// ApiServer : Empty API server struct
type Location struct {
	availabilityzone string
	region           string
	area             string
	city             string
	country          string
	province         string
}

type Flavor struct {
	flavorid      string
	totalcapacity string
}

type Storage struct {
	storagecapacity string
	typeid          string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	filename := flag.String("file", "./cluster.yaml", "cluster file name")
	max := flag.Int("max", 50, "number of clusters")
	flag.Parse()
	if *max > 9999 {
		*max = 1000
	}
	var regions [10]Location
	var flavors [3]Flavor
	var storages [3]Storage

	regions[0].availabilityzone = "NE-1"
	regions[0].region = "NE-1"
	regions[0].area = "NE-1"
	regions[0].city = "NewYork"
	regions[0].country = "US"
	regions[0].province = "NewYork"

	regions[1].availabilityzone = "NW-1"
	regions[1].region = "NW-1"
	regions[1].area = "NW-1"
	regions[1].city = "Bellevue"
	regions[1].country = "US"
	regions[1].province = "Washington"

	regions[2].availabilityzone = "SE-1"
	regions[2].region = "SE-1"
	regions[2].area = "SE-1"
	regions[2].city = "Orlando"
	regions[2].country = "US"
	regions[2].province = "Florida"

	regions[3].availabilityzone = "SW-1"
	regions[3].region = "SW-1"
	regions[3].area = "SW-1"
	regions[3].city = "Austin"
	regions[3].country = "US"
	regions[3].province = "Texas"

	regions[4].availabilityzone = "Central-1"
	regions[4].region = "Central-1"
	regions[4].area = "Central-1"
	regions[4].city = "Chicago"
	regions[4].country = "US"
	regions[4].province = "Illinois"

	regions[5].availabilityzone = "NE-2"
	regions[5].region = "NE-2"
	regions[5].area = "NE-2"
	regions[5].city = "Boston"
	regions[5].country = "US"
	regions[5].province = "Massachusettes"

	regions[6].availabilityzone = "NW-2"
	regions[6].region = "NW-2"
	regions[6].area = "NW-2"
	regions[6].city = "SanFrancisco"
	regions[6].country = "US"
	regions[6].province = "California"

	regions[7].availabilityzone = "SE-2"
	regions[7].region = "SE-2"
	regions[7].area = "SE-2"
	regions[7].city = "Atlanta"
	regions[7].country = "US"
	regions[7].province = "Georgia"

	regions[8].availabilityzone = "SW-2"
	regions[8].region = "SW-2"
	regions[8].area = "SW-2"
	regions[8].city = "LasVegas"
	regions[8].country = "US"
	regions[8].province = "Nevada"

	regions[9].availabilityzone = "Central-2"
	regions[9].region = "Central-2"
	regions[9].area = "Central-2"
	regions[9].city = "Omaha"
	regions[9].country = "US"
	regions[9].province = "Nebraska"

	flavors[0].flavorid = "1"
	flavors[0].totalcapacity = "1000"

	flavors[1].flavorid = "2"
	flavors[1].totalcapacity = "2000"

	flavors[2].flavorid = "3"
	flavors[2].totalcapacity = "3000"

	storages[0].typeid = "sas"
	storages[0].storagecapacity = "256"

	storages[1].typeid = "sata"
	storages[1].storagecapacity = "512"

	storages[2].typeid = "ssd"
	storages[2].storagecapacity = "1024"

	f, err := os.Create(*filename)
	check(err)

	defer f.Close()

	clusterNumber := 0
	for i := 0; i < 7; i++ { //storage
		for j := 0; i < 7; j++ { //favor
			for k := 0; k < len(regions); k++ {
				_, err := f.WriteString("apiVersion: globalscheduler.com/v1\n")
				_, err = f.WriteString("kind: Cluster\n")
				_, err = f.WriteString("metadata:\n")
				s := fmt.Sprintf("  name: cluster%d\n", clusterNumber)
				_, err = f.WriteString(s)
				_, err = f.WriteString("  namespace: default\n")
				_, err = f.WriteString("spec:\n")
				_, err = f.WriteString("  cpucapacity: 8\n")
				_, err = f.WriteString("  eipcapacity: 3\n")
				_, err = f.WriteString("  flavors:\n")
				switch jj := j % 6; jj {
				case 0:
					_, err = f.WriteString("  - flavorid: \"" + flavors[0].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[0].totalcapacity + "\n")
				case 1:
					_, err = f.WriteString("  - flavorid: \"" + flavors[1].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[1].totalcapacity + "\n")
				case 2:
					_, err = f.WriteString("  - flavorid: \"" + flavors[2].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[2].totalcapacity + "\n")
				case 3:
					_, err = f.WriteString("  - flavorid: \"" + flavors[0].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[0].totalcapacity + "\n")
					_, err = f.WriteString("  - flavorid: \"" + flavors[1].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[1].totalcapacity + "\n")
				case 4:
					_, err = f.WriteString("  - flavorid: \"" + flavors[0].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[0].totalcapacity + "\n")
					_, err = f.WriteString("  - flavorid: \"" + flavors[2].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[2].totalcapacity + "\n")
				case 5:
					_, err = f.WriteString("  - flavorid: \"" + flavors[1].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[1].totalcapacity + "\n")
					_, err = f.WriteString("  - flavorid: \"" + flavors[2].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[2].totalcapacity + "\n")
				case 6:
					_, err = f.WriteString("  - flavorid: \"" + flavors[0].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[0].totalcapacity + "\n")
					_, err = f.WriteString("  - flavorid: \"" + flavors[1].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[1].totalcapacity + "\n")
					_, err = f.WriteString("  - flavorid: \"" + flavors[2].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[2].totalcapacity + "\n")
				default:
					_, err = f.WriteString("  - flavorid: \"" + flavors[0].flavorid + "\"\n")
					_, err = f.WriteString("    totalcapacity: " + flavors[0].totalcapacity + "\n")
				}
				_, err = f.WriteString("  geolocation:\n")
				_, err = f.WriteString("    area: " + regions[k].area + "\n")
				_, err = f.WriteString("    city: " + regions[k].city + "\n")
				_, err = f.WriteString("    country: " + regions[k].country + "\n")
				_, err = f.WriteString("    province: " + regions[k].province + "\n")
				_, err = f.WriteString("  ipaddress: " + OpenstackIP + "\n")
				_, err = f.WriteString("  memcapacity: 256\n")
				_, err = f.WriteString("  operator:\n")
				_, err = f.WriteString("      operator: globalscheduler\n")
				_, err = f.WriteString("  region:\n")
				_, err = f.WriteString("      availabilityzone: " + regions[k].availabilityzone + "\n")
				_, err = f.WriteString("      region: " + regions[k].region + "\n")
				_, err = f.WriteString("  serverprice: 10\n")
				_, err = f.WriteString("  storage:\n")
				switch ii := i % 6; ii {
				case 0:
					_, err = f.WriteString("  - storagecapacity: " + storages[0].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[0].typeid + "\n")
				case 1:
					_, err = f.WriteString("  - storagecapacity: " + storages[1].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[1].typeid + "\n")
				case 2:
					_, err = f.WriteString("  - storagecapacity: " + storages[2].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[2].typeid + "\n")
				case 3:
					_, err = f.WriteString("  - storagecapacity: " + storages[0].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[0].typeid + "\n")
					_, err = f.WriteString("  - storagecapacity: " + storages[1].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[1].typeid + "\n")
				case 4:
					_, err = f.WriteString("  - storagecapacity: " + storages[0].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[0].typeid + "\n")
					_, err = f.WriteString("  - storagecapacity: " + storages[2].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[2].typeid + "\n")
				case 5:
					_, err = f.WriteString("  - storagecapacity: " + storages[1].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[1].typeid + "\n")
					_, err = f.WriteString("  - storagecapacity: " + storages[2].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[2].typeid + "\n")
				case 6:
					_, err = f.WriteString("  - storagecapacity: " + storages[0].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[0].typeid + "\n")
					_, err = f.WriteString("  - storagecapacity: " + storages[1].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[1].typeid + "\n")
					_, err = f.WriteString("  - storagecapacity: " + storages[2].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[2].typeid + "\n")
				default:
					_, err = f.WriteString("  - storagecapacity: " + storages[0].storagecapacity + "\n")
					_, err = f.WriteString("    typeid: " + storages[0].typeid + "\n")
				}
				_, err = f.WriteString("  homescheduler: scheduler0\n")
				_, err = f.WriteString("  homedispatcher: dispatcher0\n")
				if clusterNumber < *max-1 {
					_, err = f.WriteString("---\n")
				}
				check(err)
				clusterNumber++
				if clusterNumber >= *max {
					break
				}
			}
			if clusterNumber >= *max {
				break
			}
		}
		if clusterNumber >= *max {
			break
		}
	}
}
