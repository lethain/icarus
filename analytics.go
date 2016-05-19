package icarus

import (
	"fmt"
	"log"
	"net/http"
	"time"
)


const AnalyticsBackoff = "analytics.backoff.%v.%v"
const Referrers = "analytics.refer"
const PageReferrers = "analytics.refer.%v"
const UserAgents = "analytics.useragent"
const PageViews = "analytics.pv"
const PageViewBucket = "analytics.pv_bucket"
const PageViewPageBucket = "analytics.pv_bucket.%v"
const HistoricalReferrer = "imported from Google Analytics"
const DirectReferrer = "DIRECT"
const PageViewBonus = 60 * 60 * 24
const DefaultPeriod = 60 * 60


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

func CurrentTimestamp() int64 {
	return time.Now().Unix()
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

func timebucket(period int) (int, error) {
	if period == 0 {
		period = DefaultPeriod
	}
	now := CurrentTimestamp()
	return int(now / int64(period)), nil
}
