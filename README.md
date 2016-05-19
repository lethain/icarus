
Icarus is a minimalistic blog platform for people who want to write Markdown
or HTML documents in Git and have them served by a minimalistic server which
still provides the basic nice-to-haves: trending content, full-text search
and discovering similar content, while comfortably running on a small VPS.

It relies on quite a few opensource projects:
[Go](https://golang.org/),
[Redis](http://redis.io/),
[Bleve](http://www.blevesearch.com/),
[Bootstrap 3](http://getbootstrap.com/),
and many others.



## Installation

First, you need to install some system dependencies,
like [Glide](https://github.com/Masterminds/glide), and Redis.
On OSX, that might look like:

    brew install glide
    brew install redis

Then you need to install further dependencies via Glide:

    glide install

Then build or install the commands:

    make build

or

    make install

At which point the binaries will either be in `cmd/icarus` and `cmd/icontent`
or will be in `$GOPATH/bin/`.


## Configuring

Icarus is configured via a JSON configuration file, an example of which
is available at `config.json`, and supporting these parameters:

    {
      "server": {
        "loc": "127.0.0.1:8080",
        "proto": "http",
        "domain": "yourblog.com"
      },
      "rss": {
        "path": "/feeds/",
        "title": "Recent Pages"
      },
      "blog": {
        "name": "Your Blog",
        "results_per_page": 10,
        "pages_in_paginator": 10,
        "template_dir": "templates/",
        "static_dir": "static/"
      },
      "redis": {
        "loc": "localhost:6379"
      }
    }

You'll certainly want to change `server.domain` and `blog.name`,
and the easiest way to customize Icarus without changing the code
is to supply your own templates in `templates/` and your own static
assets in `static/` (you'll specify which files are used from `static/`
by customizing your templates).

By convention, you'll probably want to symlink your pages' static assets
into `static/` to allow the Go file server to serve your static assets,
something along the lines of:

    ln -s /Users/will/git/irrational_exuberance/static/ `pwd`/static/blog/

And you should be good to go.

## Running

Once you've followed the Installation steps, you should be able to get
things running via one of these options:

    # if you did make install, recommended
    $GOPATH/bin/icarus
    # if you did make build
    ./cmd/icarus/icarus
    # if you are making changes during development
    make icarus

You can also pass in `--config path/to/config.json` if you're not running
it from the `icarus.git` repository (or want to specify a different configuration
file).

## Adding Pages

Each article is either a Markdown or an HTML file (indicated via a trailing
`.md` or `.html` respectively), but starting with a modified JSON blob:

    "title": "This is my title",
    "summary": "This is an exciting article about...",
    "pub_date": 1184450111,
    "slug": "a-unique-slug",
    "tags": ["python", "programming"],
    "draft": true,


    Start writing your article here. Just make sure
    you have an empty line after your JSON ends.
    The trailing comma is optional.

The supported parameters are:

- `title` is the human readable title for your page,
- `summary` is the human readable description paragraph for a page,
- `pub_date` is an optional timestamp for publishing date, defaults to time it is first sync'd,
- `slug` is a unique URL component, such that `/<slug>/` is the canonical URL for a page,
- `tags` is a list of strings, for tags this page will be added to
    (tags are also used for calculating related/similar pages),
- `draft` default to false and is optional, this governs if your page is included in analytics
    and the various article lists (e.g. a draft is only accessible if you type in its slug
    by hand, they are not discoverable).

From there you use the `icontent` tool to load the content:

    $GOPATH/bin/icarus --config path/to/config.json blog/*.md
    $GOPATH/bin/icarus --config path/to/config.json blog/*.html    

And you're done.

If you want to unpublish a piece of content, the easiest solution right now
is to mark `"draft": true` in the page's configuration and it will be removed
from all indexes.

Conceivably we might want to add a `iremove` command at some point to truly
remove content, or you could just use `"draft": true` to remove the indexes
and then do `redis-cli REM slug.<post-slug>` if you like living on the edge!

## History

For reasons which are hard to explain, I've spent a lot of time over the past
decade building mediocre blog platforms, and <a href="/https://github.com/lethain/icarus">Icarus</a>
is the next in that glorious heritage.

Icarus is a Go reimagining of [Sisyphus](https://github.com/lethain/sisyphus),
which was my second personal blogging platform, inspired by the many, many
mistakes I made in my first generation "Lifeflow" blog and also by ideas
around real-time analytics and content suggestions/ranking that came from
working at Digg.

For the third generation, looking to explore some additional ideas:

1. Writing it in Go and avoiding a heavy-weight framework like Django.
1. Try to make it actually usable by someone other than myself, mostly
    as an exercise and not because I anticipate much adoption.
1. Keep using Bootstrap, maybe [Bootstrap 4](http://blog.getbootstrap.com/2015/08/19/bootstrap-4-alpha/).
1. Be faster than 60ms to load the front page. Haven't profiled this in
    a very long time, but I'd guess Redis lookups are
    responsible for most of that delay (~85 Redis lookups for the
    frontpage to load). So... let's try to hit ~5 ms, which probably means
    keeping everything in-memory (and also updating analytics out-of-band,
    which will be easy in Go). We can use Redis Pub-Sub to invalid the cache
    if we're running with more than one instance.
1. Disqus comments are mostly just ads for me, and should either be removed
    entirely or replaced with something better / different.
2. Move away from Python-only [Whoosh](http://whoosh.readthedocs.io/en/latest/)
    for search.
1. Use [rrssb](https://github.com/kni-labs/rrssb) for better, simpler
    sharing to social sites.
