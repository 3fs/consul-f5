package main

import (
	"encoding/json"
	"flag"
	"log"
	"io/ioutil"
	"strings"
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

	CONSUL_CATALOG_FILE   = "/tmp/bigip/consul_catalog.json"
	//CONSUL_CATALOG_FILE   = "config/consul_catalog.json"

	NODES 		  		  = "/node/" 
	POOLS 		  		  = "/pool/"
	VIRTUAL				  = "/virtual/"
)

type Config struct {
		Bigip map[string]string
		Ssl map[string]bool
		Retry map[string]int
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

	configureBigIp()

}

//REST call that adds pool to Bigip
func addPools(bigip BigIp){
	purl:=URL+POOLS
	for _, key := range bigip.Pools {
		_, status:=postRequest(purl,postPool(key))
    	log.Println("POST\t STATUS OF REQUEST: \t" +status+"\t POOL: \t"+key.Name)
	}
	modifyPools(bigip)
}

//REST call that adds node to Bigip
func addNodes(bigip BigIp){
	purl:=URL+NODES
	for _, key := range bigip.Members {		
		_, status:=postRequest(purl,postMember(key))
    	log.Println("POST\t STATUS OF REQUEST: \t" +status+"\t NODE: \t"+key.Name)
	}
}

//REST call that changes inf. about nodes at BigIp
func modifyNodes(bigip BigIp){
	for _, key := range bigip.Members {		
	    deleteMember, postMember := modifyMember(key)
	    _, status:=deleteRequest(deleteMember)
	    log.Println("DELETE\t STATUS OF REQUEST: \t" +status+"\t NODE: \t"+key.Name)
	    durl:=URL+NODES
	    _, status=postRequest(durl,postMember)
	    log.Println("POST\t STATUS OF REQUEST: \t" +status+"\t NODE: \t"+key.Name)
	}
}

//REST call that changes members of pool at BigIp
func modifyPools(bigip BigIp){
	for _, key := range bigip.Pools {			
		purl:=URL+POOLS+key.Name
		_, status:=putRequest(purl,putPool(key.Members,key.Balancing))
		log.Println("PUT \t STATUS OF REQUEST: \t" +status+"\t POOL: \t"+key.Name)
	}
}

//REST call that delets pool from BigIp
func deletePools(bigip BigIp){
	for _, key := range bigip.Pools {
		_, status:=	deleteRequest(deletePool(key))
	    log.Println("DELETE\t STATUS OF REQUEST: \t" +status+"\t POOL: \t"+key.Name)
	}
}

//REST call that delets node from BigIp
func deleteNodes(bigip BigIp){
	for _, key := range bigip.Members {		
		deleteMember, _ := modifyMember(key)
		_, status:=	deleteRequest(deleteMember)
    	log.Println("DELETE\t STATUS OF REQUEST: \t" +status+"\t NODE: \t"+key.Name)
	}
}

//read from file
func readFile(file string) ([]byte, error){
	File, e:= ioutil.ReadFile(file)
	if e != nil {
        return nil, e
    }
    return File , nil
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
func getConsulCatalog() (*ConsulCatalog,error){
	res:=&ConsulCatalog{}
	file, err := readFile(CONSUL_CATALOG_FILE)
	if err != nil{
		return nil,err
	}
    json.Unmarshal(file, &res)
    return res,nil
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
	    pool.Monitor=key.Monitor
	    pool.Balancing=key.Balancing
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
	    pool.Monitor=key.Monitor
	    pool.Balancing=key.Balancing


	    for _, mem := range getPoolMembers(key.Name).Nodes {
			member := Member{}
		    member.Name=mem.Name
		    member.Address=mem.Address
		    pool.AddMemberItem(member)
		}
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
func membersToJson(catalog []Member,balancing string)Members{
	members:=Members{Balancing: balancing}
	for _, key := range catalog {
		poolmember:=PoolMember{Name: key.Name}	
	    members.AddMember(poolmember)
	}
	return members
}

//send new configuration to f5 BIG IP
func applyConfiguration(consulCatalog BigIp,bigipCatalog BigIp){
            
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

//revert f5 BIG IP to previous consistent configuration
func rollBack(bigipBackupCatalog BigIp, bigipRollbackCatalog BigIp){
	applyConfiguration(bigipBackupCatalog,bigipRollbackCatalog)
}

//Prepare configuration catalogs and sends it to f5 with retrys if it fails
func configureBigIp(){
	//Generate consulCatalog from consul response data
	catalog, er :=getConsulCatalog()
	if er != nil {
		log.Println("ERROR", er)
	}else{
		consulCatalog := prepareConsulCatalog(catalog)
		//Generate BigIpCatalog from BigIp response data
		bigipCatalog := prepareBigIpCatalog(getBigipPoolCatalog(),getBigipNodeCatalog())

		err := retry(c.Retry["retry"],consulCatalog,bigipCatalog)
		if err != 0{
			log.Println("f5 BIG IP \tCONFIGURATION: \tFAILED - after 3 trys")
		}else{
			log.Println("f5 BIG IP \tCONFIGURATION: \tSUCCESS")
	}
	}	
}

//retry if something goes wrong and rollback to last consistent configuration if not successful
func retry(attempts int,consulCatalog BigIp, bigipCatalog BigIp) (err int) {
    for i := 0; ; i++ {
        applyConfiguration(consulCatalog,bigipCatalog)
        bigipAfterConf := prepareBigIpCatalog(getBigipPoolCatalog(),getBigipNodeCatalog())
        if bigipAfterConf.compareWith(consulCatalog) {
            return 0
        }else{
        	log.Println("\nReverting to previous configuration of f5\n")
			rollBack(bigipCatalog,bigipAfterConf)
        }

        log.Println("\nretrying...  retry no:",i+1,"\n")

        if i >= (attempts - 1) {
            break
        }
        
    }
    return 127
}
