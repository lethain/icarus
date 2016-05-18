package icarus

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const TagZsetByTime = "tags_by_times"
const TagZsetByPages  = "tags_by_pages"
const TagPagesZsetByTime = "tag_pages_by_time.%v"
const TagPagesZsetByTrend = "tag_pages_by_trend.%v"
const PageZsetByTime = "pages_by_time"
const PageZsetByTrend = "pages_by_trend"
const PageString = "page.%v"
const SimilarPagesByTrend = "similar_pages.%v"
const SimilarPagesExpire = 60 * 5
const PageViewBonus = (60 * 60 * 24) * 1.0
const DefaultPeriod = 60 * 60

const AnalyticsBackoff = "analytics.backoff.%v.%v"
const Referrers = "analytics.refer"
const PageReferrers = "analytics.refer.%v"
const UserAgents = "analytics.useragent"
const PageViews = "analytics.pv"
const PageViewBucket = "analytics.pv_bucket"
const PageViewPageBucket = "analytics.pv_bucket.%v"
const HistoricalReferrer = "imported from Google Analytics"
const DirectReferrer = "DIRECT"


var FilteredRefs = []string{
	"www.google.com",
	"lethain.com",
	"dev.lethain.com",
	"DIRECT",
	HistoricalReferrer,
}

var BotAgents = []string {
	"-",
	"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
	"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
	"Disqus/1.0",
	"Sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
	"Baiduspider+(+http://www.baidu.com/search/spider.htm)",
	"Baiduspider+(+http://www.baidu.jp/spider/)",
	"ichiro/4.0 (http://help.goo.ne.jp/door/crawler.html)",
	"Java/1.6.0_24",
	"Python-urllib/2.6",
	"ia_archiver (+http://www.alexa.com/site/help/webmasters; crawler@alexa.com)",
	"Mozilla/5.0 (compatible; TopBlogsInfo/2.0; +topblogsinfo@gmail.com)",
	"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
}

func ShouldIgnore(p *Page, r *http.Request) bool {
	/*
    if not slug.endswith('.png') and \
            useragent not in BOT_AGENTS and \
            'subscribers' not in lowered and \
            'bot' not in lowered and \
            not useragent.startswith('Reeder'):
*/
	// TODO: implement
	if p.Draft {
		log.Printf("ignoring %v because draft", p.Slug)
		return true
	}
	return false
}

func Referrer(r *http.Request) string {
	// if startswith("www.google") return "www.google.com"
	// TODO: implement
	return DirectReferrer
}

func SlugsForList(list string, offset int, count int, reverse bool) ([]string, error) {
	rc, err := GetConfiguredRedisClient()
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
	rc, err := GetConfiguredRedisClient()
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
	rc, err := GetConfiguredRedisClient()

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

func Track(p *Page, r *http.Request) error {
	if !ShouldIgnore(p, r) {
		rc, err := GetConfiguredRedisClient()
		if err != nil {
			return err
		}
		err = rc.Cmd("ZINCRBY", PageZsetByTrend, PageViewBonus, p.Slug).Err
		if err != nil {
			return err
		}
		for _, tag := range p.Tags {
			err := rc.Cmd("ZINCRBY", fmt.Sprintf(TagPagesZsetByTrend, tag), PageViewBonus, p.Slug).Err
			if err != nil {
				return err
			}

		}
		err = TrackAnalytics(p, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func TrackAnalytics(p *Page, r *http.Request) error {
	log.Printf("TrackAnalytics(%v)", p.Slug)
	return nil
}

// Search for pages by query string.
func Search(q string) []*Page {
	return make([]*Page, 0)
}

func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

func timebucket(period int) (int, error) {
	if period == 0 {
		period = DefaultPeriod
	}
	now := CurrentTimestamp()
	return int(now / int64(period)), nil
}

func RegisterPageTag(p *Page, tag string) error {
	rc, err := GetConfiguredRedisClient()
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
	rc, err := GetConfiguredRedisClient()
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
	rc, err := GetConfiguredRedisClient()
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
	rc, err := GetConfiguredRedisClient()
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

