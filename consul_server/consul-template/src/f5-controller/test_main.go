package main 

import(
	"encoding/json"
	"log"
)

func generateTestData(){
	createData()
	addPoolToVS()
    log.Println("Test data added to f5 \n")
}

func createData(){
	purl:=URL+NODES
	res:=&ConsulCatalog{}
    file, _ := readFile("./test_data_f5/nodes.json")
    json.Unmarshal(file, &res)
    for _, member :=range res.Members{
    	mem:=Member{}
    	mem.Address=member.Address
    	mem.Name=member.Name
        _,status := postRequest(purl,postMember(mem))
    	log.Println("POST NODE: "+mem.Name+ " STATUS OF REQUEST: " +status)
    }
    purl=URL+POOLS
    for _, pool :=range res.Pools{
    	mem:=Pool{}
    	membersToAdd:=&BigIp{}
    	mem.Name=pool.Pool
    	mem.Fullpath=pool.Fullpath
        mem.Monitor=pool.Monitor
        mem.Balancing=pool.Balancing
        _, status := postRequest(purl,postPool(mem))
    	log.Println("POST POOL: "+mem.Name+ " STATUS OF REQUEST: " +status)
    	for _, member :=range pool.Members{
    		poolmember:=Member{Name: member.Name}
    		membersToAdd.AddItem(poolmember)
    		
    	}
    	purl:=URL+POOLS+mem.Name
        _, status =putRequest(purl,putPool(membersToAdd.Members,mem.Balancing))
    	log.Println("PUT POOL NODES: "+mem.Name+ " STATUS OF REQUEST: " +status)
    }
}

func addPoolToVS(){
	res:=&VirtualServerCatalog{}
    file, _ := readFile("./test_data_f5/virtual.json")
    json.Unmarshal(file, &res)
    for _, key:=range res.Items{
        _, status := putRequest(deleteFromVS(key),putVS(key))
    	log.Println("PUT POOL: "+key.Name+ " STATUS OF REQUEST: " +status)
    }
}