package icarus

import (
	"log"
	"strconv"
)

type Tag struct {
	Slug  string
	Count int
}

func GetAllTags() ([]Tag, error) {
	rc, err := GetConfiguredRedisClient()
	if err != nil {
		return []Tag{}, err
	}
	tags, err := rc.Cmd("ZREVRANGEBYSCORE", TagZsetByPages, "+inf", "-inf", "WITHSCORES").List()
	if err != nil {
		return []Tag{}, err
	}

	t := []Tag{}
	for i := 0; i < len(tags); i += 2 {
		tag := tags[i]
		count, err := strconv.Atoi(tags[i+1])
		if err != nil {
			log.Printf("error translating tag count to int: %v", err)
			continue
		}
		t = append(t, Tag{Slug: tag, Count: count})
	}
	return t, nil
}
