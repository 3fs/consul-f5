package main

type VirtualServerCatalog struct {
    Items []struct {
        Name string         `json:"name"`
        Description string  `json:"description"`
        Destination string  `json:"destination"`
        Enabled bool        `json:"enabled"`
        Mask string         `json:"mask"`
        Pool string         `json:"pool"`
        Source string       `json:"source"`
    }                       `json:"items"`
}

type VirtualServer struct {
        Name string         `json:"name"`
        Description string  `json:"description"`
        Destination string  `json:"destination"`
        Enabled bool        `json:"enabled"`
        Mask string         `json:"mask"`
        Pool string         `json:"pool"`
        Source string       `json:"source"`
}

type BigipPoolCatalog struct {
	Pools []struct {
        Fullpath string     `json:"fullPath"`
		Name string 		`json:"name"`
		Description string  `json:"description"`
	} 						`json:"items"`
}

type BigipNodeCatalog struct {
	Nodes []struct {
		Address string 		`json:"address"`
		Name string 		`json:"name"`
	} 						`json:"items"`
}

type ConsulCatalog struct {
	Members []struct {
		Address string 		`json:"address"`
		Name    string 		`json:"name"`
	} 						`json:"members"`
	Name  string 			`json:"name"`
	Pools []struct {
		Members []struct {
			Address string 	`json:"address"`
			Name    string 	`json:"name"`
		} 					`json:"members"`
		Pool string 		`json:"pool"`
        Fullpath string     `json:"fullPath"`
	} 						`json:"pools"`
}

type Member struct {
        Name string         `json:"name"`
        Address string      `json:"address"`
}

type PoolMember struct{
    Name string             `json:"name"`
}

type Members struct{
    Members[] PoolMember    `json:"members"`
}

type Pool struct {
        Fullpath string     `json:"fullPath"`
        Name string         `json:"name"`
        Members []Member    `json:"members"`
}

type BigIp struct {
        Members []Member    `json:"members"`
        Pools []Pool        `json:"pools"`
}

//Add member to Members
func (box *Members) AddMember(item PoolMember) []PoolMember {
        box.Members = append(box.Members, item)
        return box.Members
}

//Add Member to BigIp.Members
func (box *BigIp) AddItem(item Member) []Member {
        box.Members = append(box.Members, item)
        return box.Members
}

//Add Member to Pool
func (box *Pool) AddMemberItem(item Member) []Member {
        box.Members = append(box.Members, item)
        return box.Members
}

//Add Pool to BigIp.Pools
func (box *BigIp) AddPoolItem(item Pool) []Pool {   
        box.Pools = append(box.Pools, item)
        return box.Pools
}

//Checks if member exists in BigIp.Members
func (box *BigIp) existsMember(item Member) bool {
    for _, b := range box.Members {
        if b == item {
            return true
        }
    }
    return false
}

//Checks if name of member exists
func (box *BigIp) existsMemberName(item string) bool {
    for _, b := range box.Members {
        if b.Name == item {
            return true
        }
    }
    return false
}

//Checks if pool exists in BigIp.Pools
func (box *BigIp) existsPool(item string) bool {
    for _, b := range box.Pools {
        if b.Name == item {
            return true
        }
    }
    return false
}