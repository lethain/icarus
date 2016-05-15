package icarus

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strconv"
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
	return rc.Cmd(cmd, list, offset, count).List()
}

func PagesForList(list string, offset int, count int, reverse bool) ([]*Page, error) {
	slugs, err := SlugsForList(list, offset, count, reverse)
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
	log.Printf("Track(%v)", p.Slug)
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

func timebucket(period int) (int, error) {
	if period == 0 {
		period = DefaultPeriod
	}
	nowStr := time.Now().Format(time.RFC850)
	now, err := strconv.ParseInt(nowStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(now / int64(period)), nil
}

func addPageToTag(p *Page, tag string) {
	/*
    if cli.zrank(TAG_PAGES_ZSET_BY_TIME % tag_slug, page_slug) is None:
        created = created or int(time.time())
        cli.zadd(TAG_PAGES_ZSET_BY_TIME % tag_slug, page_slug, created)
        cli.zadd(TAG_PAGES_ZSET_BY_TREND % tag_slug, page_slug, created)
        cli.zincrby(TAG_ZSET_BY_PAGES, tag_slug, 1)
*/
}

func registerPage(p *Page) {
	// cli.zadd(PAGE_ZSET_BY_TIME, slug, page['pub_date'])
        // cli.zadd(PAGE_ZSET_BY_TREND, slug, page['pub_date'])
        // for tag in page['tags']:
	//    add_page_to_tag(tag, slug, created=page['pub_date'], cli=cli)
}
