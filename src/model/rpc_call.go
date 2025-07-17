package model

type ParseBlockDataDenoReq struct {
	BlockNum  string `json:"blocknum" form:"blocknum"`
	BlockData Block  `json:"blockdata" form:"blockdata"`
}
