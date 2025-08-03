package utils

type Virtual struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Icon         string            `json:"icon"`
	Desc         string            `json:"desc"`
	Auth         map[string]string `json:"auth"`
	Devices      []Device          `json:"devices"`
	EnterpriseId string            `json:"enterpriseId"`
}

type Device struct {
	ID        string                 `json:"id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Icon      string                 `json:"icon"`
	Desc      string                 `json:"desc"`
	Auth      map[string]string      `json:"auth"`
	VirtualId string                 `json:"virtualId"`
	Location  map[string]string      `json:"location"`
	Config    map[string]interface{} `json:"config"`
	Network   map[string]string      `json:"network"`
}

type Auth struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	Password    string         `json:"passworrd"`
	Virtual     string         `json:"virtual"`
	VirtualId   string         `json:"virtualId"`
	Permissions AuthPermission `json:"permissions"`
}

type AuthPermission struct {
	Subscribers []string `json:"subscribers"`
	Publichers  []string `json:"publishers"`
}
