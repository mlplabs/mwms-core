package model

type Cell struct {
	Id            int64
	Name          string
	WhsId         int64 `json:"whs_id"`     // Id склада (может быть именован)
	ZoneId        int64 `json:"zone_id"`    // Id зоны назначения (может быть именован)
	SectionId     int   `json:"section_id"` // Id секции/блока (может быть именован)
	PassageId     int   `json:"passage_id"` // Id проезда (может быть именован)
	RackId        int   `json:"rack_id"`    // Id стеллажа (может быть именован)
	Floor         int   `json:"floor"`
	IsSizeFree    bool  `json:"is_size_free"`
	IsWeightFree  bool  `json:"is_weight_free"`
	NotAllowedIn  bool  `json:"not_allowed_in"`
	NotAllowedOut bool  `json:"not_allowed_out"`
	IsService     bool  `json:"is_service"`
	//Size          SpecificSize `json:"size"`
}
