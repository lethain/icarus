package icarus

import (
	"fmt"
	"strconv"
)

const TagZsetByTime = "tags_by_times"
const TagZsetByPages = "tags_by_pages"
const TagPagesZsetByTime = "tag_pages_by_time.%v"
const TagPagesZsetByTrend = "tag_pages_by_trend.%v"
const PageZsetByTime = "pages_by_time"
const PageZsetByTrend = "pages_by_trend"
const PageString = "page.%v"
const SimilarPagesByTrend = "similar_pages.%v"
const SimilarPagesExpire = 60 * 60 * 24

func SlugsForList(list string, offset int, count int, reverse bool) ([]string, error) {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return []string{}, err
	}
	cmd := "ZRANGE"
	if reverse {
		cmd = "ZREVRANGE"
	}
	return rc.Cmd(cmd, list, offset, offset+count).List()
}

func PagesInList(list string) (int, error) {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return 0, err
	}
	return rc.Cmd("ZCARD", list).Int()
}

func PagesForList(list string, offset int, count int, reverse bool) ([]*Page, error) {
	slugs, err := SlugsForList(list, offset, count, reverse)
	if err != nil {
		return []*Page{}, err
	}
	return PagesFromRedis(slugs)
}

// Get up to N preceeding or following pages.
func Surrounding(p *Page, num int, reverse bool) ([]*Page, error) {
	start := fmt.Sprintf("(%v", p.PubDate().Unix())
	end := "+inf"
	cmd := "ZRANGEBYSCORE"
	if reverse {
		cmd = "ZREVRANGEBYSCORE"
		end = "-inf"
	}
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return []*Page{}, err
	}
	slugs, err := rc.Cmd(cmd, PageZsetByTime, start, end, "LIMIT", 0, num).List()
	if err != nil {
		return []*Page{}, err
	}
	return PagesFromRedis(slugs)
}

func RecentPages(offset int, count int) ([]*Page, error) {
	return PagesForList(PageZsetByTime, offset, count, true)
}

func TrendingPages(offset int, count int) ([]*Page, error) {
	return PagesForList(PageZsetByTrend, offset, count, true)
}

func SimilarPages(p *Page, offset int, count int) ([]*Page, error) {
	if len(p.Tags) == 0 {
		return []*Page{}, nil
	}
	similarKey := fmt.Sprintf(SimilarPagesByTrend, p.Slug)

	// first, let's check if it's already been generated,
	// in which case we can skip regenerating it
	pgs, err := PagesForList(similarKey, offset, count, true)
	if len(pgs) > 0 || err != nil {
		return pgs, err
	}

	// union all slugs from all the page's tags,
	// relying on articles appearing in multiple tags
	// having their scored summed such that they are
	// the highest scoring pages
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return []*Page{}, err
	}
	cmds := []string{similarKey, strconv.Itoa(len(p.Tags))}
	for _, tag := range p.Tags {
		cmds = append(cmds, fmt.Sprintf(TagPagesZsetByTrend, tag))
	}
	err = rc.Cmd("ZUNIONSTORE", cmds).Err
	if err != nil {
		return []*Page{}, err
	}
	err = rc.Cmd("ZREM", similarKey, p.Slug).Err
	if err != nil {
		return []*Page{}, err
	}
	err = rc.Cmd("EXPIRE", similarKey, SimilarPagesExpire).Err
	if err != nil {
		return []*Page{}, err
	}

	// ok, let's try retrieving those slugs a second time
	return PagesForList(similarKey, offset, count, true)
}

func RegisterPageTag(p *Page, tag string) error {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return err
	}
	now := p.PubDate().Unix()
	timeKey := fmt.Sprintf(TagPagesZsetByTime, tag)
	trendKey := fmt.Sprintf(TagPagesZsetByTrend, tag)

	err = rc.Cmd("ZADD", timeKey, "NX", now, p.Slug).Err
	if err != nil {
		return fmt.Errorf("error adding to tag recent: %v", err)
	}
	err = rc.Cmd("ZADD", trendKey, "NX", now, p.Slug).Err
	if err != nil {
		return fmt.Errorf("error adding to tag trending: %v", err)
	}
	c, err := rc.Cmd("ZCOUNT", trendKey, "-inf", "+inf").Int()
	if err != nil {
		return fmt.Errorf("error calculating number of articles in tag: %v", err)
	}
	err = rc.Cmd("ZADD", TagZsetByPages, c, tag).Err
	if err != nil {
		return fmt.Errorf("error updating count of articles in tag: %v", err)
	}
	return nil
}

func UnregisterPageTag(p *Page, tag string) error {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return err
	}
	timeKey := fmt.Sprintf(TagPagesZsetByTime, tag)
	trendKey := fmt.Sprintf(TagPagesZsetByTrend, tag)

	err = rc.Cmd("ZREM", timeKey, p.Slug).Err
	if err != nil {
		return err
	}
	err = rc.Cmd("ZREM", trendKey, p.Slug).Err
	if err != nil {
		return err
	}

	c, err := rc.Cmd("ZCOUNT", trendKey, "-inf", "+inf").Int()
	if err != nil {
		return fmt.Errorf("error calculating number of articles in tag: %v", err)
	}
	err = rc.Cmd("ZADD", TagZsetByPages, c, tag).Err
	if err != nil {
		return fmt.Errorf("error updating count of articles in tag: %v", err)
	}
	return nil
}

func RegisterPage(p *Page) error {
	now := p.PubDate().Unix()
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return err
	}
	err = rc.Cmd("ZADD", PageZsetByTime, "NX", now, p.Slug).Err
	if err != nil {
		return fmt.Errorf("error adding page to recent pages: %v", err)
		return err
	}
	err = rc.Cmd("ZADD", PageZsetByTrend, "NX", now, p.Slug).Err
	if err != nil {
		return fmt.Errorf("error adding page to trending pages: %v", err)
	}
	for _, tag := range p.Tags {
		err := RegisterPageTag(p, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func UnregisterPage(p *Page) error {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return err
	}
	err = rc.Cmd("ZREM", PageZsetByTime, p.Slug).Err
	if err != nil {
		return err
	}
	err = rc.Cmd("ZREM", PageZsetByTrend, p.Slug).Err
	if err != nil {
		return err
	}
	for _, tag := range p.Tags {
		err := UnregisterPageTag(p, tag)
		if err != nil {
			return err
		}
	}
	return nil

}
