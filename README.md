
Icarus is a Go reimagining of [Sisyphus](https://github.com/lethain/sisyphus),
which was my second personal blogging platform, inspired by the many, many
mistakes I made in my first generation "Lifeflow" blog and also by ideas
around real-time analytics and content suggestions/ranking that came from
working at Digg.

For the third generation, looking to explore some additional ideas:

1. Be faster than 60ms to load the front page. Haven't profiled this in
    a very long time, but I'd guess Redis lookups are
    responsible for most of that delay (~85 Redis lookups for the
    frontpage to load). So... let's try to hit ~5 ms, which probably means
    keeping everything in-memory (and also updating analytics out-of-band,
    which will be easy in Go). We can use Redis Pub-Sub to invalid the cache
    if we're running with more than one instance.
1. Disqus comments are mostly just ads for me, and should either be removed
    entirely or replaced with something better / different.
2. Try to hack together an in-memory or Redis-backed solution for my very
    simple search needs. [Whoosh](http://whoosh.readthedocs.io/en/latest/)
    worked surprisingly well for the last one, but would prefer to reduce
    dependencies (or at least something more common like Sphinx).
1. Use [rrssb](https://github.com/kni-labs/rrssb) for better, simpler
    sharing to social sites.
2. ??

Keep doing:

1. Use Bootstrap, maybe [Bootstrap 4](http://blog.getbootstrap.com/2015/08/19/bootstrap-4-alpha/).
2. Use Redis for persistence.
3. ...

Also in terms of system administration and deployment, a few changes
in how I'm intending to deploy (although, you can do whatever you want,
I suppose):

1. The last version is running on Apache2 and mod_python,
    and then also fronted by Nginx for static assets and
    terminating connections. Let's just do Nginx fronting
    the Go server.
2. Get a [let's encrypt SSL/TKS cert](https://www.nginx.com/blog/free-certificates-lets-encrypt-and-nginx/)
    and then also enable [http2](https://www.nginx.com/blog/nginx-1-9-5/)
    both at the Nginx layer and the Go layer.
3. Give [SaltStack](https://saltstack.com/) a go for setup and deployment.


## Writing Posts

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

And that's all there is to it.


## Installation

First, you need to install some system dependencies,
like [Glide](https://github.com/Masterminds/glide), and Redis.
On OSX, that might look like:

    brew install glide
    brew install redis

Then you need to install further dependencies via Glide:

    glide install


## Deployment

Deployment is basically checking out this repo along with the repo
containing your content, then using the `icontent` tool to sync your
repository.

You'll also want to sym-link your static assets into somewhere they
can be served, depending on your setup, that might look like this:

    ln -s /Users/will/git/irrational_exuberance/static/ `pwd`/static/blog/

Realistically you'll probably want to be fronting this with Nginx for
serving your static assets (probably with far-future expires and such).