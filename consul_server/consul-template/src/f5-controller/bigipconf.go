package main

import (
	"encoding/json"
	"fmt"
	"flag"
	"log"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
	"github.com/BurntSushi/toml"
)

var (
	config        		= flag.String("c", "config.toml", "Path to configuration file")
	c             		= &Config{}
	URL					= ""
	USERNAME			= ""
	PASSWORD			= ""

)

const(
	PUT 				  = "PUT"
	DELETE 				  = "DELETE"
	POST 				  = "POST"
	GET 				  = "GET"
	BIG_IP_BACKUP		  = "/tmp/bigip/bigip_backups/"
	BIG_IP_BACKUP_NODES   = "backup_node.json"
	BIG_IP_BACKUP_VIRTUAL = "backup_virtual.json"
	BIG_IP_BACKUP_POOLS   = "backup_pools.json"
	BIGIP_VIRTUAL_FILE	  = "/tmp/bigip/big_ip_virtual.json"
	BIGIP_NODES_FILE	  = "/tmp/bigip/big_ip_nodes.json"
	BIGIP_POOLS_FILE	  = "/tmp/bigip/big_ip_pools.json"
	CONSUL_CATALOG_FILE   = "/tmp/bigip/consul_catalog.json"
	NODES 		  		  = "/node/" 
	POOLS 		  		  = "/pool/"
	VIRTUAL				  = "/virtual/"
	MEM 				  = 0644
)

type Config struct {
		Bigip map[string]string
		Ssl map[string]bool
}

func main() {
	flag.Parse()
	// read the config
	if _, err := toml.DecodeFile(*config, c); err != nil {
		log.Panic("Failed to read the configuration file", err)
	}
	//set url,username and pasword from config
	URL = "https://"+c.Bigip["host"]+"/mgmt/tm/ltm"
	USERNAME = c.Bigip["username"]
	PASSWORD = c.Bigip["password"]

	//generate test data at BigIp
	//generateTestData()

	//Generate consulCatalog from consul response data
	consulCatalog := prepareConsulCatalog(getConsulCatalog())

	//Generate BigIpCatalog from BigIp response data
	bigipCatalog := prepareBigIpCatalog(getBigipPoolCatalog(),getBigipNodeCatalog())

	//generate catalogs Add, Delete, Update
	boxAdd:=catalogToAdd(consulCatalog,bigipCatalog)
	boxUpd:=catalogToUpdate(consulCatalog,bigipCatalog)
	boxDel:=catalogToDelete(consulCatalog,bigipCatalog)
	
	//remove pools that needs to be deleted from Virtual Servers
	//delete pools
	removePoolFromVS(boxDel,getBigipVirtualCatalog())
	deletePools(boxDel)
	
	//Add new nodes and pools
	addNodes(boxAdd)
	addPools(boxAdd)

	//change members in pools
	modifyPools(boxUpd)
	modifyNodes(boxUpd)	

	//delete nodes that are no longer avalible
	deleteNodes(boxDel)
}

//REST call that adds pool to Bigip
func addPools(bigip BigIp){
	purl:=URL+POOLS
	for _, key := range bigip.Pools {
		resp, _:=postRequest(purl,postPool(key))
    	fmt.Println("POST "+string(resp))
	}
	modifyPools(bigip)
}

//REST call that adds node to Bigip
func addNodes(bigip BigIp){
	purl:=URL+NODES
	for _, key := range bigip.Members {		
		resp, _:=postRequest(purl,postMember(key))
    	fmt.Println("POST "+string(resp))
	}
}

//REST call that changes inf. about nodes at BigIp
func modifyNodes(bigip BigIp){
	for _, key := range bigip.Members {		
	    deleteMember, postMember := modifyMember(key)
	    resp, _:=deleteRequest(deleteMember)
	    fmt.Println("DELETE "+string(resp))
	    durl:=URL+NODES
	    resp, _=postRequest(durl,postMember)
	    fmt.Println("POST "+string(resp))
	}
}

//REST call that changes members of pool at BigIp
func modifyPools(bigip BigIp){
	for _, key := range bigip.Pools {			
		purl:=URL+POOLS+key.Name
		resp, _:=putRequest(purl,putPool(key.Members))
		fmt.Println("PUT "+string(resp))
	}
}

//REST call that delets pool from BigIp
func deletePools(bigip BigIp){
	for _, key := range bigip.Pools {
		resp, _:=	deleteRequest(deletePool(key))
	    fmt.Println("DELETE"+string(resp))
	}
}

//REST call that delets node from BigIp
func deleteNodes(bigip BigIp){
	for _, key := range bigip.Members {		
		deleteMember, _ := modifyMember(key)
		resp, _:=	deleteRequest(deleteMember)
    	fmt.Println("DELETE "+string(resp))
	}
}

//read from file
func readFile(file string) []byte{
	File, e:= ioutil.ReadFile(file)
	if e != nil {
        fmt.Printf("File error: %v\n", e)
        //os.Exit(1)
    }
    return File
}

//write to file
func writeFile(contents []byte, filename string){
	err := ioutil.WriteFile(filename, contents, MEM)
    if err != nil {
    	fmt.Println("ERROR",err)
        panic(err)
    }
}

//Saves file to backup directory
func copyFile(filename string, destination string){
	cpCmd := exec.Command("cp", "-rf", filename, destination)
	err := cpCmd.Run()
	if err != nil {
		fmt.Println("ERROR",err)
        panic(err)
    }
}

//backup responses of BigIp
func backupBigip(){
	copyFile(BIGIP_NODES_FILE, BIG_IP_BACKUP +
		timeFormat(time.Now()) + BIG_IP_BACKUP_NODES)
    copyFile(BIGIP_POOLS_FILE, BIG_IP_BACKUP +
    	timeFormat(time.Now()) + BIG_IP_BACKUP_POOLS)
    copyFile(BIGIP_VIRTUAL_FILE, BIG_IP_BACKUP +
    	timeFormat(time.Now()) + BIG_IP_BACKUP_VIRTUAL)
}

//time format for backup data
func timeFormat(ts time.Time)string{
	return fmt.Sprintf("%d-%02d-%02d:%02d:%02d:%02d", ts.Year(), 
		ts.Month(), ts.Day(), ts.Hour(), 
		ts.Minute(), ts.Second())
}

//get Bigip response for pools
func getBigipPoolCatalog() *BigipPoolCatalog{
	res:=&BigipPoolCatalog{}
	byt, _ := getRequest(URL+POOLS)
    json.Unmarshal(byt, &res)
    return res
}

//get Bigip response for virtualservers
func getBigipVirtualCatalog() *VirtualServerCatalog{
	res:=&VirtualServerCatalog{}
	byt, _ := getRequest(URL+VIRTUAL)
    json.Unmarshal(byt, &res)
    return res
}

//get Bigip response for nodes
func getBigipNodeCatalog() *BigipNodeCatalog{
	res:=&BigipNodeCatalog{}
	byt, _ := getRequest(URL+NODES)
    json.Unmarshal(byt, &res)
    return res
}

//get consul response
func getConsulCatalog() *ConsulCatalog{
	res:=&ConsulCatalog{}
    json.Unmarshal(readFile(CONSUL_CATALOG_FILE), &res)
    return res
}

//generate BigIp catalog from consul response catalog
func prepareConsulCatalog(res *ConsulCatalog) BigIp{
	bigip := BigIp{}
    members2:=res.Members
    pools:=res.Pools
    for _, key := range members2 {
	    member := Member{}
	    member.Name=key.Name
	    member.Address=key.Address
	    bigip.AddItem(member)
	}
	for _, key := range pools {
		pool := Pool{}
	    pool.Name=key.Pool
	    pool.Fullpath=key.Fullpath
	    for _, key := range key.Members {  
	    	member := Member{}
		    member.Name=key.Name
		    member.Address=key.Address
		    pool.AddMemberItem(member)	 
		}
		bigip.AddPoolItem(pool)
	}
	return bigip
}

//prepare BigIp catalog from REST responses catalog
func prepareBigIpCatalog(res2 *BigipPoolCatalog, res *BigipNodeCatalog) BigIp{
	bigip := BigIp{}
    members:=res.Nodes
    pools:=res2.Pools
    for _, key := range members {
	    member := Member{}
	    member.Name=key.Name
	    member.Address=key.Address
	    bigip.AddItem(member)
	}
	for _, key := range pools {
		pool := Pool{}
	    pool.Name=key.Name
	    pool.Fullpath=key.Fullpath
		bigip.AddPoolItem(pool)
	}
	return bigip
}

//generate catalog of pools and nodes that need to be added
func catalogToAdd(consul BigIp, bigip BigIp) BigIp{
	bigipAdd := BigIp{}
	for _, key := range consul.Members {
		if !bigip.existsMember(key){
	    	bigipAdd.AddItem(key)
		}
	}
	for _, key := range consul.Pools {
		if !bigip.existsPool(key.Name){
	    	bigipAdd.AddPoolItem(key)
		}
	}
	return bigipAdd
}

//generate catalog of pools and nodes that needs to be deleted
func catalogToDelete(consul BigIp, bigip BigIp) BigIp{
	bigipDel := BigIp{}
	for _, key := range bigip.Members {
		if !consul.existsMember(key){
	    	bigipDel.AddItem(key)
		}
	}
	for _, key := range bigip.Pools {
		if !consul.existsPool(key.Name){
	    	bigipDel.AddPoolItem(key)
		}
	}
	return bigipDel
}

//generate catalog of pools and nodes that needs to be updated
func catalogToUpdate(consul BigIp, bigip BigIp) BigIp{
	bigipUpd := BigIp{}
	for _, key := range consul.Pools {
		if bigip.existsPool(key.Name){    
	    	bigipUpd.AddPoolItem(key)
		}
	}
	return bigipUpd
}

//removes node from active pools
func removeNodeFromPool(todelete BigIp, activepools BigIp){
	for _, key := range activepools.Pools {
		for _, name := range getPoolMembers(strings.Split(key.Name, ":")[0]).Nodes {
			if todelete.existsMemberName(strings.Split(name.Name, ":")[0]){
		    	deleteRequest(deleteFromPool(key.Name,name.Name))
			}
		}
    	
	}
}

//remove pool from active virtual servers
func removePoolFromVS(todelete BigIp, activeServer *VirtualServerCatalog){
	for _, key := range activeServer.Items {
		if(key.Pool != ""){	
			for _, pool := range todelete.Pools{
				if pool.Fullpath==key.Pool{
					key.Pool=""
					putRequest(deleteFromVS(key),putVS(key))
				}
			}
		}
	}
}

//return pool members
func getPoolMembers(pool string) *BigipNodeCatalog{
	url:=URL+POOLS+pool+"/members"
	res:=&BigipNodeCatalog{}
    resp, _ := getRequest(url)
    json.Unmarshal(resp, &res)
    return res
}

//return members from consul catalog
func membersToJson(catalog []Member)Members{
	members:=Members{}
	for _, key := range catalog {
		poolmember:=PoolMember{Name: key.Name}	
	    members.AddMember(poolmember)
	}
	return members
}
