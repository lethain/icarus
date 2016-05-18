// TODO: move this into its own module w/its own namespace
package icarus


type PageOpt struct {
	Num int
	Offset int
	Selected bool
}

type Paginator struct {
	Pages []PageOpt
	Next bool
	Prev bool
}

/*
Create a new Paginator.

offset is the current offset (in items).
total is the total number of items.
pageSize is the number of items per page.
numPages is the number of pages to display
*/
func NewPaginator(offset int, total int, pageSize int, numPages int) Paginator {
	p := Paginator{Pages: make([]PageOpt, 0)}

	currPage := offset / pageSize
	lastPage := total / pageSize
	preceeding := (numPages - 1) / 2
	following := (numPages - 1) / 2
	if preceeding + following + 1 < numPages {
		preceeding += 1
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
	} else if currPage >= lastPage - following {
		start = lastPage - numPages
		finish = lastPage
	}

	// build the PageOpts
	for i := start; i <= finish; i++ {
		p.Pages = append(p.Pages, PageOpt{Num: i + 1, Offset: i * pageSize, Selected: i == currPage})
	}	
	return p
}
