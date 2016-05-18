package main

import(
	"encoding/json"
	 "fmt"
	 "net/http"
	 "io/ioutil"
	 "strings"
	 "crypto/tls"
)

//REST request call
func requestREST(REST string, url string, data []byte)([]byte, string){

	//checks config file for ssl
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if c.Ssl["ssl"]{
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}
	}
	
    client := &http.Client{Transport: tr}

	request, err := http.NewRequest(REST, url, nil)
    
    if(data != nil){
    	payload := strings.NewReader(string(data))
    	request, err = http.NewRequest(REST, url, payload)	
    }

	request.SetBasicAuth(USERNAME, PASSWORD)

	//send request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Sprintf("ERROR-%s", err)
	} else {
		//read response
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Sprintf("ERROR-%s", err)
		}
		return contents, response.Status
	}
    return nil, fmt.Sprintf("ERROR-%s", err)
}

//Send GET request
func getRequest(url string) ([]byte, string) {
	return requestREST(GET,url,nil)
}

//Send DELETE request
func deleteRequest(url string)([]byte, string){
	return requestREST(DELETE,url,nil)
}

//Send PUT request
func putRequest(url string, data []byte)([]byte, string){
	return requestREST(PUT,url,data)
}

//Send POST request
func postRequest(url string, data []byte) ([]byte, string){
	return requestREST(POST,url,data)
}

//Convert Member to json format
func postMember(key Member) []byte{
	jsonMember, _ := json.Marshal(key)
	return jsonMember
}

//Convert Pool to json format
func postPool(key Pool) []byte{
	pool:=&Pool{Name:   key.Name}
	jsonPool, _ := json.Marshal(pool)
	return jsonPool
}

//Convert []Members to json format
func putPool(key []Member) []byte{
	jsonPoolMembers, _ := json.Marshal(membersToJson(key))
	return jsonPoolMembers
}

//Convert VIrtualServer to json format
func putVS(key VirtualServer) []byte{
	jsonVS, _ := json.Marshal(key)
	return jsonVS
}

//Convert Member to json format and return name as string 
func modifyMember(member Member)(string,[]byte){
	jsonMember, _ := json.Marshal(member)
	return URL+NODES+member.Name, jsonMember
}

//Return url for DELETE member request
func deleteMember(member Member) string{
	return URL+NODES+member.Name
}

//Return url for DELETE pool request
func deletePool(pool Pool) string{
	return URL+POOLS+pool.Name
}

//Return url fro DELETE from pool members request
func deleteFromPool(pool string,member string)string{
	return URL+POOLS+pool+"/members/"+member
}

//Return url for DELETE from Virtua lServer request
func deleteFromVS(virtual VirtualServer)string{
	return URL+VIRTUAL+virtual.Name
}