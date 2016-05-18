// TODO: move this into its own module w/its own namespace
package icarus

type PageOpt struct {
	Num      int
	Offset   int
	Selected bool
}

type Paginator struct {
	Pages      []PageOpt
	NextOffset int
	PrevOffset int
	HasNext    bool
	HasPrev    bool
	Show       bool
}

/*
Create a new Paginator.

offset is the current offset (in items).
total is the total number of items.
pageSize is the number of items per page.
numPages is the number of pages to display
*/
func NewPaginator(offset int, total int, pageSize int, numPages int) Paginator {
	currPage := offset / pageSize
	lastPage := total / pageSize
	preceeding := (numPages - 1) / 2
	following := (numPages - 1) / 2
	if preceeding+following+1 < numPages {
		preceeding += 1
	}

	p := Paginator{Pages: make([]PageOpt, 0), Show: total > pageSize}

	// setup for pagers
	if currPage > 0 {
		p.PrevOffset = (currPage - 1) * pageSize
		p.HasPrev = true
	}
	if currPage != lastPage {
		p.NextOffset = (currPage + 1) * pageSize
		p.HasNext = true
	}

	// determine start and finish based on various
	// possible scenarios
	start := currPage - preceeding
	finish := currPage + following
	if lastPage <= numPages {
		start = 0
		finish = lastPage
	} else if currPage <= preceeding {
		start = 0
		finish = numPages
	} else if currPage >= lastPage-following {
		start = lastPage - numPages
		finish = lastPage
	}

	// build the PageOpts
	for i := start; i <= finish; i++ {
		p.Pages = append(p.Pages, PageOpt{Num: i + 1, Offset: i * pageSize, Selected: i == currPage})
	}
	return p
}
