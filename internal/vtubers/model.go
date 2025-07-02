package vtubers

type VTuber struct {
	VTuberRendered
	VTuberMeta
}

type VTuberRendered struct {
	YouTubeID     string `db:"youtube_id"`
	YouTubeHandle string `db:"youtube_handle"`
	PictureURL    string `db:"picture_url"`
	OriginalName  string `db:"original_name"`
	EnglishName   string `db:"english_name"`
	OshiMark      string `db:"oshi_mark"`
	Zodiac        string `db:"zodiac"`
	Affiliation   string `db:"affiliation"`
	Birthday      string `db:"birthday"`
	DebutDate     string `db:"debut_date"`
	Gender        string `db:"gender"`
	Height        string `db:"height"`
	Fanbase       string `db:"fanbase"`
	Status        string `db:"status"`
}

type VTuberMeta struct {
	ID       int    `json:"id" db:"id"`
	Link     string `json:"link" db:"link"`
	Modified string `json:"modified" db:"modified"`
}

type Channel struct {
	ID        string `db:"id"`
	Handle    string `db:"handle"`
	Name      string `db:"name"`
	AvatarURL string `db:"avatar_url"`
	BannerURL string `db:"banner_url"`
}
