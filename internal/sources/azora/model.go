package azora

// searchResponse maps to GET /api/query
type searchResponse struct {
	Posts      []postItem `json:"posts"`
	TotalCount int        `json:"totalCount"`
}

type postItem struct {
	ID            int             `json:"id"`
	Slug          string          `json:"slug"`
	Title         string          `json:"postTitle"`
	FeaturedImage string          `json:"featuredImage"`
	SeriesType    string          `json:"seriesType"`
	Genres        []genre         `json:"genres"`
	Chapters      []chapterInList `json:"chapters"`
}

// chapterInList is the lightweight chapter object embedded in /api/query results.
type chapterInList struct {
	ID           int     `json:"id"`
	Slug         string  `json:"slug"`
	Number       float64 `json:"number"`
	Title        string  `json:"title"`
	CreatedAt    string  `json:"createdAt"`
	IsLocked     bool    `json:"isLocked"`
	IsAccessible bool    `json:"isAccessible"`
}

// postResponse maps to GET /api/post?postSlug=
type postResponse struct {
	Post postDetail `json:"post"`
}

type postDetail struct {
	ID            int             `json:"id"`
	Slug          string          `json:"slug"`
	Title         string          `json:"postTitle"`
	Description   string          `json:"postContent"`
	FeaturedImage string          `json:"featuredImage"`
	SeriesStatus  string          `json:"seriesStatus"`
	SeriesType    string          `json:"seriesType"`
	Genres        []genre         `json:"genres"`
	Chapters      []chapterInList `json:"chapters"`
}

type genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// chapterResponse maps to GET /api/chapter?chapterId=
type chapterResponse struct {
	Chapter chapterDetail `json:"chapter"`
}

type chapterDetail struct {
	ID                  int         `json:"id"`
	IsShortLinkLocked   bool        `json:"isShortLinkLocked"`
	IsLockedByCoins     bool        `json:"isLockedByCoins"`
	IsPermanentlyLocked bool        `json:"isPermanentlyLocked"`
	Images              []pageImage `json:"images"`
}

type pageImage struct {
	Order int    `json:"order"`
	URL   string `json:"url"`
}
