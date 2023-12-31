package ipapi

// Only used information will be parsed
type Ipapi struct {
	Ip          string `json:"ip"`
	CountryName string `json:"country_name,omitempty"`
	CountryCode string `json:"country,omitempty"`
	Region      string `json:"region,omitempty"`
	Org         string `json:"org,omitempty"`
}

type Countries struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Region string `json:"region"`
}
