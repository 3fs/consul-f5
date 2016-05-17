package main 

import(
	"encoding/json"
	"fmt"
)

func generateTestData(){
	createData()
	addPoolToVS()
}

func createData(){
	purl:=URL+NODES
	res:=&ConsulCatalog{}
    json.Unmarshal(readFile("./test_data_f5/nodes.json"), &res)
    for _, member :=range res.Members{
    	mem:=Member{}
    	mem.Address=member.Address
    	mem.Name=member.Name
        resp,_ := postRequest(purl,postMember(mem))
    	fmt.Println("POST "+string(resp))
    }
    purl=URL+POOLS
    for _, pool :=range res.Pools{
    	mem:=Pool{}
    	membersToAdd:=&BigIp{}
    	mem.Name=pool.Pool
    	mem.Fullpath=pool.Fullpath
        resp, _ := postRequest(purl,postPool(mem))
    	fmt.Println("POST "+string(resp))
    	for _, member :=range pool.Members{
    		poolmember:=Member{Name: member.Name}
    		membersToAdd.AddItem(poolmember)
    		
    	}
    	purl:=URL+POOLS+mem.Name
        resp, _ =putRequest(purl,putPool(membersToAdd.Members))
    	fmt.Println("PUT "+string(resp))
    }
}

func addPoolToVS(){
	res:=&VirtualServerCatalog{}
    json.Unmarshal(readFile("./test_data_f5/virtual.json"), &res)
    for _, key:=range res.Items{
        resp, _ := putRequest(deleteFromVS(key),putVS(key))
    	fmt.Println("PUT "+string(resp))
    }
}