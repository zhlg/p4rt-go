/*
 * Copyright 2020-present Brian O'Connor
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"bytes"
	"flag"
	"fmt"
	"time"

	iris "github.com/kataras/iris/v12"
	p4 "github.com/p4lang/p4runtime/go/p4/v1"

	"github.com/bocon13/p4rt-go/pkg/acl"
	"github.com/bocon13/p4rt-go/pkg/p4rt"
)

//var writeReples sync.WaitGroup
//var failedWrites uint32
var devices map[string]string
var verbose *bool
var writeTraceChanSlice []chan p4rt.WriteTrace
var cliSlice []p4rt.P4RuntimeClient

//var count *uint64

func main() {
	//target = flag.String("target", "localhost:28000", "")
	verbose = flag.Bool("verbose", false, "")
	//count = flag.Uint64("count", 1, "")

	flag.Parse()

	app := iris.Default()

	app.Get("/devices", getDevices)
	app.Put("/devices", addDevice)
	app.Get("/pipeconf", getDevicesPipeConf)
	app.Put("/pipeconf", setDevicesPipeConf)

	app.Get("/devices/{id:string}/pipeconf/tables", getDeviceTables)

	app.Post("/rest/v2/acls", acl.AddAcls)
	app.Delete("/rest/v2/acls/{id:string}", acl.DeleteAcl)
	app.Patch("/rest/v2/acls/{id:string}", acl.ModifyAcl)
	app.Get("/rest/v2/acls", acl.GetAllAcls)
	app.Get("/rest/v2/acls/{id:string}", acl.GetAcl)

	app.Post("/rest/v2/acls/{id:string}/acl_entries", acl.AddAclEntries)
	app.Delete("/rest/v2/acls/{id:string}/acl_entries", acl.DeleteAclEntries)
	app.Patch("/rest/v2/acls/{pid:string}/acl_entries/{id:uint32}", acl.ModifyAclEntries)
	app.Get("/rest/v2/acls/{pid:string}/acl_entries", acl.GetAllAclEntries)
	app.Get("/rest/v2/acls/{pid:string}/acl_entries/{id:uint32}", acl.GetAclEntries)
	app.Put("/rest/v2/acls/{pid:string}/acl_entries/", acl.UpdateAclEntries)

	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}

func getDeviceTables(ctx iris.Context) {

}

func getDevices(ctx iris.Context) {
	if len(devices) == 0 {
		ctx.Writef("No Device.")
		return
	}

	var stringBuilder bytes.Buffer
	for key, value := range devices {
		stringBuilder.WriteString(key)
		stringBuilder.WriteString("=")
		stringBuilder.WriteString(value)
		stringBuilder.WriteString(",")
	}

	devStr := stringBuilder.String()
	ctx.Writef("devices: [%s]\n", devStr)
}

//TODO: get the device info from atomix
func addDevice(ctx iris.Context) {
	devices = ctx.URLParams()

	num := len(devices)
	if num != 0 {
		ctx.Writef("Add %d device successfully.\n", num)
	} else {
		ctx.Writef("No device added.\n")
	}
}

type devPipeConf struct {
	Name     string `json:"name" validate:"required"`
	P4info   string `json:"p4_info" validate:"required"`
	PipeConf string `json:"pipe_conf" validate:"required"`
}

type deviceAllConf struct {
	id uint64
	dpc devPipeConf
	cli p4rt.P4RuntimeClient
	flows string
}

var devicesAllConf []deviceAllConf

func getDevicesPipeConf(ctx iris.Context) {
	var num int
	for _, devpc := range devicesAllConf {
		if devpc.dpc.Name == "" {
			ctx.Writef("device name is empty, device num %d, devpc : %v.\n", len(devicesAllConf), devpc)
			continue
		}

		num++
		ctx.Writef("get devid %d devicePipeConf: %v.\n", devpc.id, devpc.dpc)
	}
	if num == 0 {
		ctx.Writef("No Device PipeConf.")
	}
}

func setDevicesPipeConf(ctx iris.Context) {
	if len(devices) == 0 {
		ctx.Writef("The device is empty, please add the device first.\n")
		return
	}

	var devicesPipeConf []devPipeConf

	if err := ctx.ReadJSON(&devicesPipeConf); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().Title("Read Json failure").DetailErr(err))
		//ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}

	start := time.Now()

	for devid, devpc := range devicesPipeConf {
		if devpc.Name == "" {
			ctx.Writef("device %d name is empty, devices num %d, devpc : %v.\n", devid, len(devicesPipeConf), devpc)
			continue
		}

		dev, ok := devices[devpc.Name]
		if !ok {
			ctx.Writef("device name %s don't match.\n", devpc.Name)
			return
		}

		ctx.Writef("device name %s match %s.\n", devpc.Name, dev)
		fmt.Printf("set device %d PipeConf: %v.\n", devid+1, devpc)

		client, err := p4rt.GetP4RuntimeClient(dev, 1)
		if err != nil {
			panic(err)
		}

		err = client.SetMastership(p4.Uint128{High: 0, Low: 1})
		if err != nil {
			panic(err)
		}

		err = client.SetForwardingPipelineConfig(devpc.P4info, devpc.PipeConf)
		if err != nil {
			panic(err)
		}

		// Set up write tracing for test
		writeTraceChan := make(chan p4rt.WriteTrace, 100)
		client.SetWriteTraceChan(writeTraceChan)

		writeTraceChanSlice = append(writeTraceChanSlice, writeTraceChan)
		cliSlice = append(cliSlice, client)

		var device deviceAllConf

		device.id = uint64(devid+1)
		device.dpc = devpc
		device.cli = client

		devicesAllConf = append(devicesAllConf, device)
		ctx.Writef("hello %v, set p4runtime pipeconf successfully.\n", devpc)
	}

	//doneChan := make(chan bool)
	//go func() {
	//	var writeCount, lastCount uint64
	//	printInterval := 1 * time.Second
	//	ticker := time.Tick(printInterval)
	//	for {
	//		for _, v := range writeTraceChanSlice {
	//			if v != nil {
	//				select {
	//				case trace := <-v:
	//					writeCount += uint64(trace.BatchSize)
	//					if writeCount >= (*count)*uint64(len(writeTraceChanSlice)) {
	//						doneChan <- true
	//						return
	//					}
	//				case <-ticker:
	//					if *verbose {
	//						fmt.Printf("\033[2K\rWrote %d of %d (~%.1f flows/sec)...",
	//							writeCount, *count, float64(writeCount-lastCount)/printInterval.Seconds())
	//						lastCount = writeCount
	//					}
	//				}
	//			}
	//		}
	//	}
	//}()

	// Send the flow entries
	//writeReples.Add(int((*count)*uint64(len(cliSlice))))
	//start := time.Now()
	//for _, client := range(cliSlice) {
	//	if client != nil {
	//		go SendTableEntries(client, *count)
	//	}
	//}
	// Wait for all writes to finish
	//<-doneChan
	duration := time.Since(start).Seconds()
	//fmt.Printf("\033[2K\r%f seconds, %d writes, %f writes/sec\n",
	//	duration, (*count)*uint64(len(cliSlice)), float64((*count)*uint64(len(cliSlice)))/duration)
	//writeReples.Wait()
	//fmt.Printf("Number of failed writes: %d\n", failedWrites)
	fmt.Printf("set %d devices pipeline time consuming : %f\n", len(devices), duration)
}
