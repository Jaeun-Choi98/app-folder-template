package redismodel

import "time"

type Sample struct {
	Id   int64
	Name string
}

func (d *Sample) GetId() int64      { return d.Id }
func (d *Sample) SetId(id int64)    { d.Id = id }
func (d *Sample) TableName() string { return "sample" }
func (d *Sample) GetIndexFields() map[string]interface{} {
	return map[string]interface{}{
		"name": d.Name,
	}
}
func (d *Sample) GetTTL() time.Duration {
	return 1 * time.Minute
}
