package icarus

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const AnalyticsBackoff = "analytics.backoff.%v"
const Referrers = "analytics.refer"
const PageReferrers = "analytics.refer.%v"
const UserAgents = "analytics.useragent"
const PageViews = "analytics.pv"
const PageViewBucket = "analytics.pv_bucket"
const PageViewPageBucket = "analytics.pv_bucket.%v"
const HistoricalReferrer = "imported from Google Analytics"
const DirectReferrer = "DIRECT"
const PageViewBonus = 60 * 60 * 24
const RateLimitPeriod = 60
const DaySeconds = 60 * 60 * 24

var FilteredRefs = []string{
	"www.google.com",
	"lethain.com",
	"dev.lethain.com",
	"DIRECT",
	HistoricalReferrer,
}

var BotAgents = []string{
	"-",
	"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
	"mozilla/5.0 (compatible; yahoo! slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
	"disqus/1.0",
	"sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
	"baiduspider+(+http://www.baidu.com/search/spider.htm)",
	"baiduspider+(+http://www.baidu.jp/spider/)",
	"ichiro/4.0 (http://help.goo.ne.jp/door/crawler.html)",
	"java/1.6.0_24",
	"python-urllib/2.6",
	"ia_archiver (+http://www.alexa.com/site/help/webmasters; crawler@alexa.com)",
	"mozilla/5.0 (compatible; topblogsInfo/2.0; +topblogsinfo@gmail.com)",
	"mozilla/5.0 (compatible; baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
}

func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

func IsRateLimited(key string) bool {
	script := `local current
current = redis.call("incr",KEYS[1])
if tonumber(current) == 1 then
    redis.call("expire", KEYS[1], 60)
end
return current
`
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)
	if err != nil {
		log.Printf("error getting redis client: %v", err)
		return true
	}
	ratelimited, err := rc.Cmd("EVAL", script, 2, key, RateLimitPeriod).Int()
	if err != nil {
		log.Printf("error checking ratelimit: %v", err)
		return true
	}
	return ratelimited != 1
}

func ShouldIgnore(p *Page, r *http.Request) bool {
	if strings.HasSuffix(p.Slug, ".png") || strings.HasSuffix(p.Slug, ".ico") {
		return true
	}
	ua := r.UserAgent()
	lua := strings.ToLower(ua)
	for _, botAgent := range BotAgents {
		if botAgent == lua {
			return true
		}
	}
	if p.Draft || strings.HasPrefix(lua, "reeder") || strings.Contains(lua, "bot") {
		return true
	}
	ip := GetIP(r)
	rlKey := fmt.Sprintf(AnalyticsBackoff, ip)
	return IsRateLimited(rlKey)
}

func Referrer(r *http.Request) string {
	// if startswith("www.google") return "www.google.com"
	// TODO: implement
	return DirectReferrer
}

func GetIP(r *http.Request) string {
	if r.Header.Get("X-Forwarded-For") != "" {
		return r.Header.Get("X-Forwarded-For")
	}
	if r.Header.Get("X-Real-IP") != "" {
		return r.Header.Get("X-Real-IP")
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func Track(p *Page, r *http.Request) error {
	if !ShouldIgnore(p, r) {
		rc, err := GetRedisClient()
		defer PutRedisClient(rc)
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

		// tracking referrers
		referKeys := []string{
			Referrers,
			fmt.Sprintf(PageReferrers, p.Slug),
		}
		referrer := Referrer(r)
		for _, key := range referKeys {
			err := rc.Cmd("ZINCRBY", key, 1, referrer).Err
			if err != nil {
				return err
			}
		}
		// total pageviews by user agents
		err = rc.Cmd("ZINCRBY", UserAgents, 1, r.UserAgent()).Err
		if err != nil {
			return err
		}
		// total pageviews by page
		err = rc.Cmd("ZINCRBY", PageViews, 1, p.Slug).Err
		if err != nil {
			return err
		}
		// tracking pageviews, bucketed by day
		bucket := timebucket(DaySeconds)
		bucketKeys := []string{
			PageViewBucket,
			fmt.Sprintf(PageViewPageBucket, p.Slug),
		}
		for _, key := range bucketKeys {
			err := rc.Cmd("ZINCRBY", key, 1, bucket).Err
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func timebucket(period int) int {
	now := CurrentTimestamp()
	return int(now / int64(period))
}
